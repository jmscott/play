package main

func server(root *ast) error {

	compile(root) 

	// wake up all flow operators wired during compilation
	close(compiling)

	//  wait forever
	<- make(chan interface{})

	return nil		//  not reached
}
