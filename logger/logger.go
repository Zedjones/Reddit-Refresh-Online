package logger

import (
	"fmt"
	"log"
	"os"
)

const logFileName = "rr_online.log"

/*
Log is the global logger to be used in the project
*/
var Log *log.Logger

func init() {
	logFile, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Error opening log file: %s", logFileName))
	}
	Log = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
}
