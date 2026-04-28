package main

func server(root *ast) error {

	flo := compile(root) 
WTF("flo.op_count: %d", flo.op_count)

	// wake up all flow operators wired during compilation
	close(compiling)

	//  wait forever
	<- make(chan interface{})

	return nil		//  not reached
}
