/*
TBD:
1. find a better way to import the csv to db
2. consider to get rid of gpdb, use local db file instead.
3. add support for Bandwagonhost and DigitalOcean
4. need remove the csv file on target host
*/
package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/pborman/getopt/v2"
)

var (
	enableDebug         = false // default disable debug mode, unless use -D
	mtrFolder           = ""
	outputFolder        = ""
	dbName              = ""
	dbHost              = ""
	dbUser              = ""
	dbPassword          = ""
	dbPort              = ""
	threads             = ""
	db                  *sql.DB
	mtrReportTable      = "mtr_report"
	downloadReportTable = "download_report"
	reportTable         = "final_report"
	vendor              = "vultr"
)

func init() {

	/* get the options */
	optDebug := getopt.Bool('D', "Display debug message")                                                // enable DEBUG mode
	optHelp := getopt.Bool('H', "Help")                                                                  // print the help message
	optMtrFolder := getopt.String('b', "/usr/local/sbin", "Folder which holds the mtr binary, default:") // path of mtr utility
	optOutputFolder := getopt.String('o', ".", "The location of the csv file, default:")

	/* db connection info: */
	optDBName := getopt.String('d', "thePatriot", "Database to connect, default:")
	optDBHost := getopt.String('h', "mdw", "Database to connect, default:")
	optDBUser := getopt.String('u', "gpadmin", "the user of DB, default:")
	optDBPassword := getopt.String('p', "abc123", "the password of DB user, default:")
	optDBPort := getopt.String('P', "5432", "the port of DB, default:")
	optVendor := getopt.String('v', "vultr", "the vendor of the VPS, aviliable: vultr, linode; default:")
	optthreads := getopt.String('t', "5", "How many threads running in parallel, default:")

	getopt.Parse()
	enableDebug = *optDebug
	mtrFolder = *optMtrFolder
	outputFolder = *optOutputFolder
	dbName = *optDBName
	dbHost = *optDBHost
	dbUser = *optDBUser
	dbPassword = *optDBPassword
	dbPort = *optDBPort
	vendor = *optVendor
	threads = *optthreads

	if *optHelp {
		getopt.Usage()
		os.Exit(0)
	}
}
func main() {

	// get current run timestampe
	curTime := time.Now()
	dateFormatA := curTime.Format("2006-01-01_15-04-05")
	dateFormatB := curTime.Format("2006-01-02 15:04:00")

	// step1: TBD, checking if the required tools like mtr has been installed
	checkRequirement()

	// step2: get the url list from vultr.com
	allTestLink := getURL(vendor)

	// step4: connect to db
	connectDB()

	// step3: run mtr against all site, save the result to csv. this has to be done at local mac so the result is real
	mtrTest(allTestLink, outputFolder, dateFormatA)

	// step5: load the csv to DB
	// loadCsvToDB(dateFormatA, outputFolder)

	// step6: test the speed of the download against each site
	testDownloadSpeed(allTestLink, dateFormatB, vendor)

	// step7: load the data into db and get the result
	generateReport("aa", dateFormatB)

}

func connectDB() {

	DBconnStr := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser + " dbname=" + dbName + " password=" + dbPassword + " sslmode=disable"
	plog("DEBUG", "The DB conn string is: "+DBconnStr)
	plog("INFO", "Connecting to DB ["+dbName+"]...")

	dbConn, err := sql.Open("postgres", DBconnStr)
	if err != nil {
		plog("FATAL", err.Error())
	}
	db = dbConn
}

func loadCsvToDB(csvFile string) {

	/* Import the csv */
	CsvFileBase := filepath.Base(csvFile)
	scpCommand := "scp " + csvFile + " " + dbUser + "@" + dbHost + ":/tmp/"
	runCommand(scpCommand, true)

	// runQueryWithNoOutput("truncate " + mtrReportTable)
	copyQuery := "copy " + mtrReportTable + " from '/tmp/" + CsvFileBase + "' csv"
	runQueryWithNoOutput(copyQuery)
	// cleanCsvCommand := "ssh " + dbUser + "@" + dbHost + ` "rm -f /tmp/` + CsvFileBase + `"`
	// runCommand(cleanCsvCommand, false)

}

func runQueryWithNoOutput(query string) {
	plog("DEBUG", "will run query: ["+query+"]...")
	_, error := db.Query(query)
	if error != nil {
		plog("FATAL", "Failed to run the query ["+query+"], error: "+error.Error())
	}
}

func testDownloadSpeed(targetSites []string, testTimeStamp string, vendor string) {

	testDuration := 10 // test download for x secs
	downloadSpeed := "n/a"
	var fileSize = 100 * 1024 * 1024 // 100 MB

	/* create a work folder */
	workFolder := fmt.Sprintf("/tmp/.testSpeed.%d", os.Getpid())
	plog("INFO", "Start to testing the download speed, creating the temp folder ["+workFolder+"]...")
	err := os.MkdirAll(workFolder, 0755)
	if err != nil {
		plog("FATAL", "Failed to create temp folder ["+workFolder+"], exit!")
	}
	defer os.RemoveAll(workFolder) // clean the work folder if abnormally exit

	runQueryWithNoOutput("truncate " + downloadReportTable)
	for _, testURL := range targetSites {

		switch vendor {
		case "vultr":
			plog("INFO", "testing download speed for link: "+testURL)
		case "linode": // http://speedtest.toronto1.linode.com  -> http://speedtest.toronto1.linode.com/100MB-toronto1.bin
			fileName := strings.Split(testURL, ".")[1]
			testURL = testURL + "/100MB-" + fileName + ".bin"
		default:
			plog("FATAL", "unsupported vendor ["+vendor+"], exit!")
		}
		hostName := strings.Split(testURL, "/")[2]
		getFileCMD := "curl " + testURL + " -o /dev/null -m " + strconv.Itoa(testDuration) + " 2>&1 | grep 'Operation timed out'"
		downloadSummary := runCommand(getFileCMD, false)
		plog("DEBUG", "The downloadSummary is ["+downloadSummary+"]")

		if len(downloadSummary) == 0 { // download finished within x sec
			downloadSpeed = strconv.Itoa(fileSize / testDuration / 1024)
			plog("DEBUG", "The download for site "+hostName+" has finsished, downloadSummary is ["+downloadSummary+"]")
		} else {
			plog("DEBUG", "The download for site "+hostName+" has not finsished, downloadSummary is ["+downloadSummary+"]")
			downloadedSizeString := strings.Split(downloadSummary, " ")[9]
			// check if the output is number
			match, _ := regexp.MatchString("([0-9]+)", downloadedSizeString)
			if match {
				downloadedSize, _ := strconv.Atoi(downloadedSizeString)
				if downloadedSize > 0 {
					downloadSpeed = strconv.Itoa(downloadedSize / testDuration / 1024)
				} else {
					downloadSpeed = "0"
				}
			} else {
				plog("DEBUG", "The download for site "+hostName+" has finsished, something goes wrong, set speed as 0")
				plog("ERROR", "Failed to get download speed for site ["+testURL+"]...")
				downloadSpeed = "0"
			}
		}
		insertQuery := "insert into " + downloadReportTable + " values('" + testTimeStamp + "', '" + vendor + "', '" + hostName + "', '" + downloadSpeed + "')"
		runQueryWithNoOutput(insertQuery)
	}
}

func generateReport(csvFile string, date string) {

	plog("INFO", "Clean up invalid result from download report...")
	cleanUpQuery := `update ` + downloadReportTable + ` set result = 0 where result !~ '^[0-9]+$';`
	runQueryWithNoOutput(cleanUpQuery)

	plog("INFO", "Generating final report table...")
	InsertReportQuery := `insert into ` + reportTable + ` 
    select 
		to_timestamp(testdate, 'YYYY-MM-DD HH24:MI')::timestamp as testDate,
		vendor,
        hostname,
    	result::int as Speed,
        -- CASE when result > 0 then result::int ELSE 0 END as Speed, 
        avg_lossrate,
        max_lossrate,
		avg_latency,
		max_latency
    from 
        (select host as hostname,
            max(lossrate) as max_lossrate, 
            avg(lossrate)::numeric(100,2) as avg_lossrate, 
            max(worst) as max_latency, 
            max(avg) as avg_latency
            from ` + mtrReportTable + `
            where ip != '???' 
            group by host ) m,
        ` + downloadReportTable + ` d
    where d.host = m.hostname 
	order by max_lossrate,Speed;`
	runQueryWithNoOutput(InsertReportQuery)

	plog("INFO", "Generating final report...")
	getReportQuery := `select 
		hostname,
		speed,
		avg_lossrate,
		max_lossrate,
		avg_latency,
		max_latency 
	from ` + reportTable + ` 
	where testdate = '` + date + `'
	and vendor = '` + vendor + `'
	order by 2 desc`

	plog("DEBUG", "will run query ["+getReportQuery+"]")
	rows, queryErr := db.Query(getReportQuery)
	if queryErr != nil {
		plog("FATAL", queryErr.Error())
	}
	var host, speed, avgLoss, maxLoss, avgLatency, maxLatency string
	fmt.Printf("%-30s %15s %15s %15s %15s %15s\n", "HostName", "speed(KB/s)", "avgLossRate", "maxLossRate", "avgLatency", "maxLatency")
	for rows.Next() {
		err := rows.Scan(&host, &speed, &avgLoss, &maxLoss, &avgLatency, &maxLatency)
		if err != nil {
			plog("FATAL", err.Error())
		}
		fmt.Printf("%-30s %15s %15s %15s %15s %15s\n", host, speed, avgLoss, maxLoss, avgLatency, maxLatency)
	}
}

/* checking if the required tools has been installed */
func checkRequirement() {
	plog("INFO", "Checking if we have installed MTR...")
	if len(mtrFolder) > 0 {
		plog("INFO", "searching mtr under: ["+mtrFolder+"]...")
		testMtrCommand := "cd " + mtrFolder + ";" + mtrFolder + "/mtr --version"
		result := runCommand(testMtrCommand, false)
		if len(result) == 0 {
			plog("FATAL", "Unable to find mtr binary, please specific the path by adding -b [/path/to/mtr] option.")
		} else {
			plog("INFO", "Found the mtr binary under ["+mtrFolder+"]!")
		}
	}
}

/* test the speed with mtr against each site of the url, save the output into local storage as csv format */
func mtrTest(urls []string, outputFolder string, curTimestamp string) {

	testDuration := "60" // how long will mtr take to check the network.
	// outputFile := outputFolder + "/" + "speedTestReport_" + nowDate + ".csv"
	// plog("DEBUG", "The output file is ["+outputFile+"]")

	// clean up the mtrReportTable
	runQueryWithNoOutput("truncate " + mtrReportTable)

	// Start multi thrad to handle all the urls
	plog("INFO", "Starting to run mtr test with ["+threads+"] threads")
	var wg sync.WaitGroup
	var jobs = make(chan string, 2) // channel has buffer length = 2

	numThreads, _ := strconv.Atoi(threads)
	for i := 0; i < numThreads; i++ { // limit the thread
		go func() {
			for job := range jobs {
				runMtrUtility(job, outputFolder, testDuration, curTimestamp) // run the function
				wg.Done()

			}
		}()
	}
	for _, url := range urls {
		jobs <- url
		wg.Add(1)
		plog("INFO", "Starting thread to handle url ["+url+"]...")
	}
	wg.Wait()
	plog("INFO", "All done, please review the output file under: ["+outputFolder+"]!")
	// csv header: Mtr_Version,Start_Time,Status,Host,Hop,Ip,Loss%,Snt, ,Last,Avg,Best,Wrst,StDev
}

func runMtrUtility(url string, outputFolder string, testDuration string, nowDate string) {

	targetSite := strings.Split(url, "/")[2]
	plog("INFO", "[Worker] Calling mtr to test the speed of site: ["+targetSite+"]...")
	testCommand := "cd " + mtrFolder + "; sudo " + mtrFolder + "/mtr " + targetSite + " -r -w -c " + testDuration + " -C | grep -v \"^Mtr_Version\" | sed 's/;/,/g' " // send 60 pings to target site
	resultCSV := runCommand(testCommand, false) + "\n"

	/* write the output to outputFile */
	outputFile := outputFolder + "/" + "speedTestReport_" + targetSite + "_" + nowDate + ".csv"
	plog("DEBUG", "The output file is ["+outputFile+"]")

	fd, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		plog("FATAL", "Failed to open the output file ["+outputFile+"], reason: ["+err.Error()+"]")
	}
	message := []byte(resultCSV)
	if _, err := fd.Write(message); err != nil {
		fd.Close() // ignore error; Write error takes precedence
		plog("FATAL", "Failed to write to the output file ["+outputFile+"], reason: ["+err.Error()+"]")
	}
	loadCsvToDB(outputFile)
}

func getURL(vendor string) (listOfURL []string) {

	getLinkCmd := ""
	switch vendor {
	case "vultr":
		getLinkCmd = "curl -s https://www.vultr.com/resources/faq/#downloadspeedtests | grep 100MB | awk -F 'href=' '{print $2}' | grep -v ipv6 | grep https | awk '{print $1}' | sed 's/\\\"//g'"
		plog("DEBUG", "will check the vultr site, command: "+getLinkCmd)
	case "linode":
		getLinkCmd = ` curl -s https://www.linode.com/speed-test/ | grep o-button | grep 'speedtest'  | awk '{print $4}' | awk -F'=' '{print $2}' | sed 's/\"//g' | sed 's/\/$//g'`
		plog("DEBUG", "will check the linode site, command: "+getLinkCmd)
	default:
		plog("FATAL", "unsupported vendor ["+vendor+"], exit!")
	}

	// plog("INFO", "Getting the URL list, please input the password for sudo when prompt...")
	// cmd := "curl -s https://www.vultr.com/resources/faq/#downloadspeedtests | grep 100MB | awk -F 'href=' '{print $2}' | grep -v ipv6 | grep https | awk '{print $1}' | sed 's/\\\"//g'"
	curlOutput := runCommand(getLinkCmd, true)
	if len(curlOutput) == 0 {
		plog("FATAL", "curl command failed, Unable to get site list from vultr.com, please check the URL!")
	}
	var allSite = strings.Split(curlOutput, "\n")
	return allSite
}

func runCommand(cmd string, errorOut bool) (output string) {

	plog("DEBUG", "Execute command ["+cmd+"]...")

	out, err := exec.Command("bash", "-c", cmd).Output()
	outputFinal := strings.TrimSpace(string(out)) // remove the new line at the end
	plog("DEBUG", "The output is: ["+string(outputFinal)+"]")

	if err != nil {
		if errorOut == false {
			plog("ERROR", "Failed to exeute command ["+cmd+"]")
			plog("ERROR", "The error message is ["+err.Error()+"]")
		} else {
			plog("ERROR", "Failed to exeute command ["+cmd+"]")
			plog("FATAL", "The error message is ["+err.Error()+"]")
		}
	}
	return string(outputFinal)
}

func plog(logLevel string, message string) {

	// define the color code here:
	lightRed := "\033[38;5;9m"
	red := "\033[38;5;1m"
	green := "\033[38;5;2m"
	yellow := "\033[38;5;3m"
	cyan := "\033[38;5;14m"
	//darkBlue := "\033[38;5;25m"
	normal := "\033[39;49m"

	var colorCode string
	var errorOut = 0

	switch logLevel {
	case "INFO":
		colorCode = green
	case "WARN":
		colorCode = yellow
	case "ERROR":
		colorCode = lightRed
	case "FATAL":
		colorCode = red
		errorOut = 1
	case "DEBUG":
		if enableDebug == true {
			colorCode = cyan
		} else {
			return
		}
	default:
		colorCode = normal
	}
	curTime := time.Now()
	fmt.Printf("%s"+curTime.Format("2006-01-02 15:04:05")+" [%s] %s\n", colorCode, logLevel, message)
	fmt.Printf("%s", normal)
	if errorOut == 1 {
		os.Exit(1)
	}
}
