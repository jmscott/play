package main

func server(root *ast) error {

	flo := compile(root) 

	go func() {
		osx_wg.Wait()
		exit(0)
	}()

	close(flo.resolved)	//  fire the first flow

	<- make(chan bool)

	return nil
}
