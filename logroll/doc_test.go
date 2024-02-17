package logroll_test

import (
	"github.com/jmscott/play/logroll"
	"time"
)

func ExampleRoller() (err error) {

	// Open a rollable transaction file in
	//
	//	data/example.txn
	//
	// that rolls example.txn every 10 mminutes to a file named
	//
	//	data/example-YYYYMMDD_HHMMSS[-+]hhmm.txn
	//
	// where YYYYMMDD_HHMMSS is the time of day in local timezone
	// and [-+]hhmm is the timezone offset in hours and minutes.
	roll, err := logroll.OpenRoller(
		"example",
		"Hz",
		logroll.HzTick(10*time.Minute),
		logroll.Directory("data"),
		logroll.FileSuffix("txn"),
	)

	// ...

	//  write a date stamped transaction record "OPEN-TXN"
	err = roll.Write([]byte(
		time.Now().Format("2006/01/02 15:04:05") + "\t" + "OPEN-TXN\n"),
	)

	// ...

	//  write a "CLOSE-TXN" record
	err = roll.Write([]byte(
		time.Now().Format("2006/01/02 15:04:05") +
			"\t" +
			"CLOSE-TXN\n",
	))

	return roll.Close()
}

func ExampleOpenLogger() (*logroll.Logger, error) {

	// Open a rollable log file, typically for a server
	//
	//	log/example-Tue.log
	//
	// that truncates and rolls to new log file near midnight local time.
	//
	//	log/example-Wed.log
	//
	roll, err := logroll.OpenRoller("example", "Dow",
		logroll.Directory("log"),
		logroll.FileSuffix("log"),
	)
	if err != nil {
		return nil, err
	}
	return logroll.OpenLogger(roll, logroll.HeartbeatTick(time.Minute))
}
