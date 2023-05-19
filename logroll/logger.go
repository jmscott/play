// Note: need to track change in timezone!
// Note: think about renamed Logger to ServerLogger
package logroll

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// a rollable log file typical for server process
type Logger struct {
	roll               *Roller
	heartbeat_tick     time.Duration
	client_data        interface{}
	pre_roll_callback  log_callback
	post_roll_callback log_callback
}

//  A client callback invoked by Logger.
type log_callback func(client_data interface{}) (msgs []string)

// variadic options passed to function OpenLogger()
type log_option func(log *Logger) log_option

var logger_default = Logger{
	heartbeat_tick: 10 * time.Second,
}

//  Set client data passed to pre/post log callbacks.
func LogClientData(client_data interface{}) log_option {
	return func(log *Logger) log_option {
		previous := log.client_data
		log.client_data = client_data
		return LogClientData(previous)
	}
}

// Validate a heartbeat tick.
// The tick must be validated before calling HeartbeatTick()
func ValidHeartbeatTick(tick time.Duration) error {
	if tick < 0 {
		return errors.New(
			fmt.Sprintf("heart beat tick < 0: %s", tick))
	}
	return nil
}

// How often is a heartbeat "alive" message written to the log file.
// The default value is 10 seconds.  The heartbeat is not validated.
// Call ValidHeartbeatTick() to validate.
func HeartbeatTick(tick time.Duration) log_option {
	return func(log *Logger) log_option {
		previous := log.heartbeat_tick
		log.heartbeat_tick = tick
		return HeartbeatTick(previous)
	}
}

// PreLogRollCallback() sets the callback to invoke immediately before rolling
// the underlying log file.  The callback can not invoke INFO(), WARN() or
// ERROR().  To write messages before closing the underlying file have the
// callback return [][]byte.
//
// The client_data passed to the callback is set with the option LogClientData
func PreLogRollCallback(callback log_callback) log_option {
	return func(log *Logger) log_option {
		previous := log.pre_roll_callback
		log.pre_roll_callback = callback
		return PreLogRollCallback(previous)
	}
}

// PostLogRollCallback() sets the callback to invoke immediately after rolling
// the underlying log file.  The callback can not invoke INFO(), WARN() or
// ERROR().  To write messages before closing the underlying file have the
// callback return [][]byte.
//
// The client_data passed to the callback is set with the option LogClientData
func PostLogRollCallback(callback log_callback) log_option {
	return func(log *Logger) log_option {
		previous := log.post_roll_callback
		log.post_roll_callback = callback
		return PostLogRollCallback(previous)
	}
}

//  format a server timestamped message for Roller
func (log *Logger) record(format string, args ...interface{}) []byte {
	return []byte(
		time.Now().Format("2006/01/02 15:04:05") +
			": " +
			fmt.Sprintf(format, args...) +
			"\n",
	)
}

func call_log_roll_pre(client_data interface{}) (msgs [][]byte) {

	log := client_data.(*Logger)

	//  tack final messages onto slice returned to roller
	add := func(m [][]byte, format string, args ...interface{}) [][]byte {
		return append(m, log.record(
			log.roll.base_name+": "+format, args...),
		)
	}
	if log.pre_roll_callback != nil {
		for msg := range log.pre_roll_callback(log.client_data) {
			msgs = add(msgs, "%s", msg)
		}
	}
	for _, msg := range log.raw_preamble() {
		msgs = append(msgs, msg)
	}
	msgs = add(msgs, "rolling log file")

	return
}

func call_log_roll_post(client_data interface{}) (msgs [][]byte) {

	log := client_data.(*Logger)

	//  tack first messages onto slice returned to roller
	add := func(m [][]byte, format string, args ...interface{}) [][]byte {
		return append(m, log.record(
			log.roll.base_name+": "+format, args...),
		)
	}

	msgs = add(msgs, "rolled log file")
	for _, msg := range log.raw_preamble() {
		msgs = append(msgs, msg)
	}

	//  tack on messages returned by post roll callback
	if log.post_roll_callback != nil {
		for msg := range log.post_roll_callback(log.client_data) {
			msgs = add(msgs, "%s", msg)
		}
	}
	return
}

func (log *Logger) raw_preamble() (msg [][]byte) {

	roll := log.roll

	ep, err := os.Executable()
	if err == nil {
		msg = append(msg, log.record("executable path: %s", ep))
	} else {
		msg = append(msg, log.record("os.Executable() failed: %s", err))
	}
	msg = append(msg, log.record("roll directory: %s", roll.directory))
	msg = append(msg, log.record("roll log path: %s", roll.path))
	msg = append(msg, log.record("hz tick: %s", roll.hz_tick))
	msg = append(msg, log.record("heartbeat tick: %s", log.heartbeat_tick))

	env := os.Environ()
	msg = append(msg, log.record("process environment: %d vars", len(env)))
	for _, kv := range env {
		msg = append(msg, log.record("	%s", kv))
	}
	return
}

func (log *Logger) preamble() {
	for _, msg := range log.raw_preamble() {
		log.roll.Write(msg)
	}
}

func OpenLogger(roll *Roller, options ...log_option) (*Logger, error) {
	if roll.pre_roll_callback != nil {
		return nil, errors.New("Roller.PreRollCalback must be nil")
	}
	if roll.post_roll_callback != nil {
		return nil, errors.New("Roller.PostRollCalback must be nil")
	}

	//  set custom callbacks invoked by roller
	roll.pre_roll_callback = call_log_roll_pre
	roll.post_roll_callback = call_log_roll_post

	log := &Logger{}
	*log = logger_default
	log.roll = roll
	roll.client_data = log

	// evaluate variadic options.
	for _, opt := range options {
		opt(log)
	}

	log.INFO("hello, world")
	log.preamble()

	go log.heartbeat()

	return log, nil
}

// Note: what happens if logger is closed!
func (log *Logger) heartbeat() {

	tick := time.NewTicker(log.heartbeat_tick)
	for range tick.C {
		if log.roll == nil {
			break
		}
		log.INFO("alive")
	}
}

func (log *Logger) Close() error {

	if log.roll == nil {
		return nil
	}
	log.INFO("good bye, cruel world")
	err := log.roll.Close()
	log.roll = nil
	return err
}

func (log *Logger) INFO(format string, args ...interface{}) {
	log.roll.Write(log.record(format, args...))
}

func (log *Logger) ERROR(format string, args ...interface{}) {
	log.roll.Write(log.record("ERROR: "+format, args...))
}

func (log *Logger) WARN(format string, args ...interface{}) {
	log.roll.Write(log.record("WARN: "+format, args...))
}
