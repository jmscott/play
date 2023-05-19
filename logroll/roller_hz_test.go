package logroll

import (
	"testing"
	"time"
)

func (roll *Roller) write(msg string) {
	roll.Write([]byte("TestHz: " + msg + "\n"))
}

func TestHz(t *testing.T) {

	roll, err := OpenRoller("test", "Hz",
		Directory("tmp"),
		FileSuffix("txn"),
		HzTick(time.Second*10),
	)
	if err != nil {
		t.Fatalf("OpenRoller() failed: %s", err)
		return
	}
	defer roll.Close()

	// test print functions

	stop_time := time.Now().Add(time.Minute + 10*time.Second)
	tick := time.NewTicker(time.Second)
	for now := range tick.C {
		roll.Write([]byte(now.Format("2006/01/02 15:04:05") + "\n"))
		if now.After(stop_time) {
			tick.Stop()
			break
		}
	}
}
