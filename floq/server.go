package main

func server(root *ast) error {

	flo, concurrent_count := compile(root) 

	con_group.gmux.Add(int(concurrent_count))

	go flo.cop(concurrent_count)

	// wake up all flow operators wired during compilation
	close(compiling)	//  wake upflow operators

	//  wait forever
	<- make(chan interface{})

	return nil		//  not reached
}
