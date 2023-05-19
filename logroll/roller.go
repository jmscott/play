package logroll

import (
	"errors"
	"fmt"
	"os"
	"time"
)

//  A rollable file.
type Roller struct {
	base_name string

	directory   string
	path        string
	file_perm   os.FileMode
	file        *os.File
	file_suffix string

	read_c    chan ([]byte)
	request_c chan (chan interface{})
	done_c    chan (interface{})

	poll_roll_tick time.Duration

	//  rate to roll file for hz driver
	hz_tick time.Duration

	driver      *driver
	driver_data interface{}

	pre_roll_callback  roll_callback
	post_roll_callback roll_callback

	client_data interface{}
}

type roll_option func(roll *Roller) roll_option

var roller_default = Roller{
	directory:      ".",
	poll_roll_tick: 3 * time.Second,
	file_perm:      00640,
	file_suffix:    "log",
	hz_tick:        10 * time.Minute,
}

//  A client callback invoked by Roller.
type roll_callback func(client_data interface{}) (msgs [][]byte)

type driver struct {
	name      string
	open      func(*Roller) error
	close     func(*Roller) error
	roll      func(*Roller, time.Time) error
	poll_roll func(*Roller, time.Time) (bool, error)
}

// Atomically read messages from a channel and write to rollable file.
func (roll *Roller) read() {

	//  invoke client defined callback, panicing on error
	call := func(what string, cb roll_callback) {
		if cb == nil {
			return
		}
		for _, msg := range cb(roll.client_data) {
			_, err := roll.file.Write(msg)
			if err != nil {
				roll.panic("roll("+what+").Write", err)
			}
		}
	}

	for {
		now := time.Now()
		select {

		case <-roll.done_c:
			return

		//  request to roll the file
		case reply := <-roll.request_c:
			call("pre", roll.pre_roll_callback)
			err := roll.driver.roll(roll, now)
			if err != nil {
				roll.panic("roll.Write", err)
			}
			call("post", roll.post_roll_callback)
			reply <- new(interface{})

		//  write a message to file
		case msg := <-roll.read_c:
			_, err := roll.file.Write(msg)
			if err == nil {
				continue
			}
			os.Stderr.Write(msg)
			roll.panic("Write", err)
		}
	}
}

func (roll *Roller) poll_roll() {

	tick := time.NewTicker(roll.poll_roll_tick)
	for now := range tick.C {
		rollable, err := roll.driver.poll_roll(roll, now)

		if err != nil {
			roll.panic("poll_roll", err)
		}
		if rollable == false {
			continue
		}
		reply := make(chan interface{})
		roll.request_c <- reply
		<-reply
	}
}

func (roll *Roller) panic(what string, err error) {

	var driver_name string

	if roll.driver != nil {
		driver_name = roll.driver.name
	}
	fmt.Fprintf(
		os.Stderr,
		"%s: PANIC: %s(%s): Logger.%s() failed: %s\n",
		time.Now().Format("2006/01/02 15:04:05"),
		roll.base_name,
		driver_name,
		what,
		err,
	)
	panic(err)
}

func ValidHzTick(tick time.Duration) error {
	if tick < time.Second {
		return errors.New(fmt.Sprintf("hz tick < 1s: %s", tick))
	}
	return nil
}

func HzTick(tick time.Duration) roll_option {
	return func(roll *Roller) roll_option {
		previous := roll.hz_tick
		roll.hz_tick = tick
		return HzTick(previous)
	}
}

//  Directory() sets the directory containing the rolled file.
//  The default directory is "."
func Directory(directory string) roll_option {
	return func(roll *Roller) roll_option {
		previous := roll.directory
		roll.directory = directory
		return Directory(previous)
	}
}

//  PreRollCallback() sets the callback to invoke immediately before rolling
//  the underlying file.  The callback can not invoke roll.Write()!
//  To write messages before closing the underlying file have the callback
//  return [][]byte.
func PreRollCallback(callback roll_callback) roll_option {
	return func(roll *Roller) roll_option {
		previous := roll.pre_roll_callback
		roll.pre_roll_callback = callback
		return PreRollCallback(previous)
	}
}

//  Set client data passed to pre/post roll callbacks.
func RollClientData(client_data interface{}) roll_option {
	return func(roll *Roller) roll_option {
		previous := roll.client_data
		roll.client_data = client_data
		return RollClientData(previous)
	}
}

//  PostRollCallback() sets the callback to invoke immediately after rolling
//  the underlying file.  The callback can not invoke roll.Write()!
//  To write messages to the newly rolled file have the callback return
//  [][]byte.
func PostRollCallback(callback roll_callback) roll_option {
	return func(roll *Roller) roll_option {
		previous := roll.post_roll_callback
		roll.post_roll_callback = callback
		return PostRollCallback(previous)
	}
}

func FileSuffix(file_suffix string) roll_option {
	return func(roll *Roller) roll_option {
		previous := roll.file_suffix
		roll.file_suffix = file_suffix
		return FileSuffix(previous)
	}
}

// Open a File Roller with a base file name, a driver and variable number of
// options.  The base file is the prefix for the rolled file names.  The driver
// can be "Dow" or "Hz".  The "Dow" driver creates and rolls files with the
// day of the week embedded the file name.  The "Hz" driver creates a file
// named <base_name>.<suffix> and then rolls to a file
// <base_name>-YYYYMMDD_hhmmss[-+]hrmi.<suffix> at regular rate of time.
// An error is returned or the Roller is ready to accept bytes via
// logroll.Write().
func OpenRoller(
	base_name,
	driver_name string,
	options ...roll_option,
) (*Roller, error) {
	if base_name == "" {
		return nil, errors.New("empty roll base name")
	}
	roll := &Roller{}
	*roll = roller_default
	roll.base_name = base_name

	switch driver_name {
	case "Dow":
		roll.driver = dow_driver
	case "Hz":
		roll.driver = hz_driver
	default:
		return nil, errors.New("unknown roll driver: " + driver_name)
	}

	// evaluate variadic options
	for _, opt := range options {
		opt(roll)
	}

	// open specific roller
	err := roll.driver.open(roll)
	if err != nil {
		return nil, err
	}

	// request to close log file
	roll.done_c = make(chan (interface{}))

	// open the logger message channel
	roll.read_c = make(chan ([]byte))

	// open the roll request chan
	roll.request_c = make(chan (chan interface{}))

	// start background to read requests for messages
	go roll.read()

	//
	go roll.poll_roll()

	return roll, nil
}

func (roll *Roller) Close() error {

	if roll == nil || roll.driver == nil {
		return nil
	}
	if roll.file != nil {
		roll.done_c <- new(interface{})

		// Note: in the past Sync() seems to fix issues with incomplete
		//       log writes.
		//roll.file.Sync()
	}
	return roll.driver.close(roll)
}

func (roll *Roller) Write(msg []byte) error {
	roll.read_c <- msg
	return nil
}
