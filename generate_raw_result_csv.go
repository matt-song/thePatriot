package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pborman/getopt/v2"
)

var (
	enableDebug = false // default disable debug mode, unless use -D
	_MTR        = ""
	outputFile  = ""
)

func init() {

	// get the options
	optDebug := getopt.Bool('D', "Display debug message") // enable DEBUG mode
	optHelp := getopt.Bool('h', "Help")                   // print the help message
	// optUrl := getopt.String('l', "the link from the vultr which contain the site list")
	optMtr := getopt.String('b', "/usr/sbin/mtr", "path to mtr binary, default: /usr/local/sbin/mtr") // mtr utility
	optOutput := getopt.String('o', "./speedTest.$$.out", "the output of the result")

	getopt.Parse()
	enableDebug = *optDebug
	_MTR = *optMtr
	outputFile = *optOutput
	if *optHelp {
		getopt.Usage()
		os.Exit(0)
	}
}
func main() {

	// step1: TBD, checking if the required tools like mtr has been installed
	checkRequirement()
	// step2: get the url list from vultr.com
	allTestLink := getURL() // the link url is like: https://par-fr-ping.vultr.com/vultr.com.100MB.bin

	// step3: run mtr against all site, save the result to csv
	mtrTest(allTestLink, outputFile)

}

/* checking if the required tools has been installed */
func checkRequirement() {
	plog("INFO", "Checking if we have installed MTR...")
	if len(_MTR) > 0 {
		plog("INFO", "searching path: ["+_MTR+"]...")
		testMtrCommand := _MTR + " --version"
		result := runCommand(testMtrCommand, false)
		if len(result) == 0 {
			plog("FATAL", "Unable to find mtr binary, please specific the path by adding -b [/path/to/mtr] option.")
		}
	}
}

/* test the speed with mtr against each site of the url, save the output into local storage as csv format */
func mtrTest(urls []string, outputFile string) {

	for _, url := range urls {
		targetSite := strings.Split(url, "/")[2]
		plog("INFO", "Working on site: ["+targetSite+"]...")
		plog("INFO", "Calling mtr to test the speed of site: ["+targetSite+"]...")
		// send 60 pings to target site
		testCommand := "sudo " + _MTR + " " + targetSite + " -r -w -c 60 -C"
		resultCSV := runCommand(testCommand, false)
		message := []byte(resultCSV)
		err := ioutil.WriteFile("outputFile", message, 0644)
		if err != nil {
			plog("FATAL", "Failed to write to output file ["+outputFile+"], reason: ["+err.Error()+"]")
		}
	}

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
