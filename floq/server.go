package main

func server(root *ast) {

	//  compile pass2 ast.  the first flo created compile().

	flo := compile(root) 

	//  wake up all flow operators patiently waiting for compilation
	//  to complete

	close(flo.compiling)

	//  wait forever, such is the burden of a server

	<- make(chan interface{})
}
