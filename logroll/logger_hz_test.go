package logroll

import (
	"testing"
	"time"
)

const (
	hz_heartbeat_tick = 2 * time.Second
	hz_sleep_tick     = 30 * time.Minute
)

func TestHzHeartbeat(t *testing.T) {

	roll, err := OpenRoller("test", "Hz",
		HzTick(20*time.Second),
		Directory("tmp"),
	)
	if err != nil {
		t.Fatalf("OpenRoller() failed: %s", err)
		return
	}

	log, err := OpenLogger(roll, HeartbeatTick(hz_heartbeat_tick))
	if err != nil {
		t.Fatalf("OpenLogger() failed: %s", err)
		return
	}

	// test print functions

	log.INFO("TestHzHeartbeat: INFO: hi")
	log.WARN("TestHzHeartbeat: hi")
	log.ERROR("TestHzHeartbeat: hi")

	defer log.Close()

	go func() {
		tick := time.NewTicker(time.Second)
		for range tick.C {
			log.INFO("tick")
		}
	}()

	//  test heartbeat

	log.INFO("sleep tick: %s", hz_sleep_tick)
	time.Sleep(hz_sleep_tick)
	log.INFO("awoke from %s sleep", hz_sleep_tick)
}
