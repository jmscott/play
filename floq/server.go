package main

import (
)

func server(root *ast) error {

	flo, _, err := compile(root) 
	if err != nil {
		return err
	}

	close(flo.resolved)	//  fires first flow

	return err
}
