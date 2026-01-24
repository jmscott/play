package main

func server(root *ast) error {

	flo := compile(root) 

	close(flo.resolved)	//  fire the first flow

	<- make(chan bool)	//  wait forever	

	return nil
}
