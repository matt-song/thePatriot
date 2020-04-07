/*
TBD:
1. change the program from single thread to muliple thread.
2. find a better way to import the csv to db
3. consider to get rid of gpdb, use local db file instead.
4. re-write the sql so it can only query the current date (where date=xxxxx)
5. remove the minLoss column, its useless (local site always 0)
*/
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/pborman/getopt/v2"
)

var (
	enableDebug = false // default disable debug mode, unless use -D
	mtrFolder   = ""
	outputFile  = ""
	dbName      = ""
	dbHost      = ""
	dbUser      = ""
	dbPassword  = ""
	dbPort      = ""
)

func init() {

	/* get the options */
	optDebug := getopt.Bool('D', "Display debug message") // enable DEBUG mode
	optHelp := getopt.Bool('H', "Help")                   // print the help message
	// optUrl := getopt.String('l', "the link from the vultr which contain the site list")
	optMtrFolder := getopt.String('b', "/usr/local/sbin", "Folder which holds the mtr binary, default:") // mtr utility
	optOutputFolder := getopt.String('o', ".", "The output of the result, default:")
	// db connection info:
	optDBName := getopt.String('d', "thePatriot", "Database to connect")
	optDBHost := getopt.String('h', "aio1", "Database to connect")
	optDBUser := getopt.String('u', "gpadmin", "dbuser")
	optDBPassword := getopt.String('p', "abc123", "password")
	optDBPort := getopt.String('P', "5432", "port of DB")

	getopt.Parse()
	enableDebug = *optDebug
	mtrFolder = *optMtrFolder
	outputFile = *optOutputFolder
	dbName = *optDBName
	dbHost = *optDBHost
	dbUser = *optDBUser
	dbPassword = *optDBPassword
	dbPort = *optDBPort

	if *optHelp {
		getopt.Usage()
		os.Exit(0)
	}
}
func main() {

	// step1: TBD, checking if the required tools like mtr has been installed
	checkRequirement()

	// step2: get the url list from vultr.com
	allTestLink := getURL()

	// step3: run mtr against all site, save the result to csv. this has to be done at local mac so the result is real
	csvReport := mtrTest(allTestLink, outputFile)

	// step4: load the data into db and get the result
	generateReport(csvReport)
}

func generateReport(csvFile string) {

	DBconnStr := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser + " dbname=" + dbName + " password=" + dbPassword + " sslmode=disable"
	plog("DEBUG", "The DB conn string is: "+DBconnStr)
	plog("INFO", "Connecting to DB...")

	db, err := sql.Open("postgres", DBconnStr)
	if err != nil {
		log.Fatal(err)
	}

	/* Import the csv */
	CsvFileBase := filepath.Base(csvFile)
	scpCommand := "scp " + csvFile + " " + dbUser + "@" + dbHost + ":/tmp/" + CsvFileBase
	runCommand(scpCommand, true)
	_, importErr := db.Query("copy test_report from '/tmp/" + CsvFileBase + "' csv")
	if importErr != nil {
		log.Fatal(importErr)
	} else {
		plog("INFO", "CSV file has been imported, generating report now...")
	}

	// generate the result
	query := `select
	host, 
	min(lossrate) as min_lossrate, 
	max(lossrate) as max_lossrate, 
	avg(lossrate)::numeric(100,2) as avg_lossrate, 
	max(worst) as max_worst, 
	max(avg) as max_avg_ping 
	from test_report  
		where ip != '???' group by host order by 3`

	rows, queryErr := db.Query(query)
	if queryErr != nil {
		log.Fatal(queryErr)
	}
	var host, minLoss, maxLoss, avgLoss, maxWorstPing, avgWorstPing string
	fmt.Printf("%-30s %10s %10s %10s %10s %10s\n", "HostName", "minLoss", "maxLoss", "avgLoss", "maxWorstPing", "avgWorstPing")
	for rows.Next() {
		err = rows.Scan(&host, &minLoss, &maxLoss, &avgLoss, &maxWorstPing, &avgWorstPing)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%-30s %10s %10s %10s %10s %10s\n", host, minLoss, maxLoss, avgLoss, maxWorstPing, avgWorstPing)
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
func mtrTest(urls []string, outputFolder string) (csvFile string) {

	curTime := time.Now()
	nowDate := curTime.Format("2006-01-01_15-04-05")
	outputFile := outputFolder + "/" + "speedTestReport_" + nowDate + ".csv"
	plog("DEBUG", "The output file is ["+outputFile+"]")

	for _, url := range urls {
		targetSite := strings.Split(url, "/")[2]
		// plog("INFO", "Working on site: ["+targetSite+"]...")
		plog("INFO", "Calling mtr to test the speed of site: ["+targetSite+"]...")
		testCommand := "cd " + mtrFolder + "; sudo " + mtrFolder + "/mtr " + targetSite + " -r -w -c 60 -C | grep -v \"^Mtr_Version\" " // send 60 pings to target site
		resultCSV := runCommand(testCommand, false) + "\n"

		/* write the output to outputFile */
		fd, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			plog("FATAL", "Failed to open the output file ["+outputFile+"], reason: ["+err.Error()+"]")
		}
		message := []byte(resultCSV)
		if _, err := fd.Write(message); err != nil {
			fd.Close() // ignore error; Write error takes precedence
			plog("FATAL", "Failed to write to the output file ["+outputFile+"], reason: ["+err.Error()+"]")
		}
	}
	plog("INFO", "All done, please review the output file: ["+outputFile+"]!")
	/*csv header: Mtr_Version,Start_Time,Status,Host,Hop,Ip,Loss%,Snt, ,Last,Avg,Best,Wrst,StDev, */
	return outputFile
}
func getURL() (listOfURL []string) {

	plog("INFO", "Getting the URL list, please input the password for sudo when prompt...")
	cmd := "curl -s https://www.vultr.com/resources/faq/#downloadspeedtests | grep 100MB | awk -F 'href=' '{print $2}' | grep -v ipv6 | grep https | awk '{print $1}' | sed 's/\\\"//g'"
	curlOutput := runCommand(cmd, true)
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
