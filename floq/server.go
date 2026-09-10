package main

func server(root *ast) {

	for i := uint8(0);  i < floq_flows;  i++ {

		flo := compile(root)
		close(flo.compiling)
	}

	//  wait forever, such is the burden of a server

	<- make(chan interface{})
}
