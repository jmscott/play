package main

import (
	"github.com/jmscott/work/logroll"

	"fmt"
	"os"
	"time"
)

var log *logroll.Logger

func log_init(name string) {

	roll, err := logroll.OpenRoller(name, "Dow",
		logroll.Directory("log"),
		logroll.FileSuffix("log"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logroll.OpenRoller() failed: %s", err)
		os.Exit(1)
	}
	log, err = logroll.OpenLogger(roll,
		logroll.HeartbeatTick(10*time.Second),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logroll.OpenLogger() failed: %s", err)
		os.Exit(1)
	}
}

func INFO(format string, args ...interface{}) {
	log.INFO(format, args...)
}

func ERROR(format string, args ...interface{}) {
	log.ERROR("ERROR: "+format, args...)
}

func WARN(format string, args ...interface{}) {

	log.WARN("WARN: "+format, args...)
}

func leave(exit_status int) {

	if db != nil {
		INFO("closing sql database ...")
		err := db.Close()
		db = nil
		if err != nil {
			ERROR("db.Close() failed", err)
		}
	}

	//  seems to force async i/o
	os.Stderr.Sync()

	if log != nil {
		err := log.Close()
		log = nil
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"\n%s: rasqld: ERROR: "+
					"close(log file) failed: %s\n",
				time.Now().Format("2006/01/02 15:04:05"),
				err,
			)
		}
	}
	os.Exit(exit_status)
}

func die(format string, args ...interface{}) {

	if log == nil {
		os.Stderr.Write([]byte(
			fmt.Sprintf("ERROR: "+format+"\n", args...),
		))
	} else {
		ERROR(format, args...)
	}
	leave(1)
}
