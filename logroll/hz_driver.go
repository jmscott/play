package logroll

import (
	"fmt"
	"os"
	"time"
)

var hz_driver = &driver{
	name: "Hz",

	open:      (*Roller).hz_open,
	close:     (*Roller).hz_close,
	roll:      (*Roller).hz_roll,
	poll_roll: (*Roller).hz_poll_roll,
}

var zero_time time.Time

func (roll *Roller) hz_open() (err error) {

	path := roll.directory +
		string(os.PathSeparator) +
		roll.base_name +
		"." +
		roll.file_suffix

	//  move an existing log file to time stamped version

	_, err = os.Stat(path)
	if err == nil {
		ts_path := roll.hz_path(time.Now())
		err = os.Rename(path, ts_path)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	mode := os.O_CREATE | os.O_WRONLY
	roll.file, err = os.OpenFile(path, mode, roll.file_perm)
	if err != nil {
		return err
	}
	roll.path = path
	roll.driver_data = time.Now().Add(roll.hz_tick)
	return nil
}

func (roll *Roller) hz_close() error {

	if roll.file == nil {
		return nil
	}
	f := roll.file
	roll.driver_data = nil
	err := f.Close()
	if err != nil {
		return err
	}
	roll.file = nil
	return os.Rename(roll.path, roll.hz_path(time.Now()))
}

func (roll *Roller) hz_poll_roll(now time.Time) (bool, error) {

	return now.After(roll.driver_data.(time.Time)), nil
}

func tzo2file_name(sec int) string {

	min := (sec % 3600) / 60 //  hours
	if min < 0 {
		min = -min
	}
	return fmt.Sprintf("%+03d%02d", sec/3600, min)
}

func (roll *Roller) hz_roll(now time.Time) error {

	roll.driver_data = now.Add(roll.hz_tick)

	err := roll.file.Close()
	roll.file = nil
	if err != nil {
		return err
	}

	//  move log/<name>.txn to log/<name>-YYYYMMDD_HHMMSS[+-]hhmm

	err = os.Rename(roll.path, roll.hz_path(now))
	if err != nil {
		return err
	}

	mode := os.O_TRUNC | os.O_CREATE | os.O_WRONLY
	roll.file, err = os.OpenFile(roll.path, mode, roll.file_perm)
	return err
}

func (roll *Roller) hz_path(now time.Time) string {
	_, tzo := now.Zone()
	return roll.directory +
		string(os.PathSeparator) +
		roll.base_name +
		"-" +
		fmt.Sprintf("%d%02d%02d_%02d%02d%02d%s",
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			now.Second(),
			tzo2file_name(tzo),
		) +
		"." +
		roll.file_suffix
}
