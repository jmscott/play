package logroll

import (
	"testing"
	"time"
)

const heartbeat_tick = 2 * time.Second

func TestDowHeartbeat(t *testing.T) {

	roll, err := OpenRoller("test", "Dow")
	if err != nil {
		t.Fatalf("OpenRoller() failed: %s", err)
		return
	}

	log, err := OpenLogger(roll, HeartbeatTick(heartbeat_tick))
	if err != nil {
		t.Fatalf("OpenLogger() failed: %s", err)
		return
	}

	// test print functions

	log.INFO("TestDowHeartbeat: INFO: hi")
	log.WARN("TestDowHeartbeat: hi")
	log.ERROR("TestDowHeartbeat: hi")

	defer log.Close()

	//  test heartbeat

	sleep_tick := heartbeat_tick + time.Second
	log.INFO("sleep tick: %s", sleep_tick)
	time.Sleep(sleep_tick)
	log.INFO("awoke from %s sleep", sleep_tick)
}
