package main

func server(root *ast) error {

	flo, err := compile(root) 
	if err != nil {
		return err
	}

	go func() {
		osx_wg.Wait()
		exit(0)
	}()

	close(flo.resolved)	//  fire the first flow

	<- make(chan bool)

	return nil
}
