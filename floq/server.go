package main

func server(root *ast) error {

	flo := compile(root) 

	flo.run_group.Add(int(run_count))

	//  fire the first "run"
	flo.get()

	<- make(chan bool)      //  wait forever

	return nil		//  not reached
}
