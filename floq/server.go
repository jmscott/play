package main

func server(root *ast) error {

	flo := compile(root) 

WTF("server: close(%p)", flo.resolved)
	close(flo.resolved)	//  fire the first flow
WTF("server: closed")

	<- make(chan bool)	//  wait forever	

	return nil
}
