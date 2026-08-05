package main

func server(root *ast) {

	//  compile pass2 ast.  the first flo created compile().

	compile(root) 

	//  wake up all flow operators pateitnly waiting for compilation
	//  to complete

	close(compiling)

	//  wait forever, such is the burden of a server

	<- make(chan interface{})
}
