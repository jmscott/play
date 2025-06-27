package main

import (
	"time"
)

func server(root *ast) error {

	flo, _, err := compile(root) 
	if err != nil {
		return err
	}

	flo.resolved = make(chan struct{})

	close(flo.resolved)	//  fires first flow

	for {
		time.Sleep(time.Second)
	}
	return err
}
