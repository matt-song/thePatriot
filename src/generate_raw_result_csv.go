package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pborman/getopt/v2"
)

var (
	enableDebug = false // default disable debug mode, unless use -D
	mtrFolder   = ""
	outputFile  = ""
)

func init() {

	/* get the options */
	optDebug := getopt.Bool('D', "Display debug message") // enable DEBUG mode
	optHelp := getopt.Bool('h', "Help")                   // print the help message
	// optUrl := getopt.String('l', "the link from the vultr which contain the site list")
	optMtrFolder := getopt.String('b', "/usr/local/sbin", "Folder which holds the mtr binary, default:") // mtr utility
	optOutputFolder := getopt.String('o', ".", "The output of the result, default:")

	getopt.Parse()
	enableDebug = *optDebug
	mtrFolder = *optMtrFolder
	outputFile = *optOutputFolder
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
	mtrTest(allTestLink, outputFile)
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
func mtrTest(urls []string, outputFolder string) {

	curTime := time.Now()
	nowDate := curTime.Format("2006-01-01_12-01")
	outputFile := outputFolder + "/" + "speedTestReport_" + nowDate + ".csv"
	plog("DEBUG", "The output file is ["+outputFile+"]")

	for _, url := range urls {
		targetSite := strings.Split(url, "/")[2]
		// plog("INFO", "Working on site: ["+targetSite+"]...")
		plog("INFO", "Calling mtr to test the speed of site: ["+targetSite+"]...")
		testCommand := "cd " + mtrFolder + "; sudo " + mtrFolder + "/mtr " + targetSite + " -r -w -c 2 -C" // send 60 pings to target site
		resultCSV := runCommand(testCommand, false)

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
}
func getURL() (listOfURL []string) {

	plog("INFO", "Getting the URL list...")
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
