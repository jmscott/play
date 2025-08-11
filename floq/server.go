package main

func server(root *ast) error {

	flo, err := compile(root) 
	if err != nil {
		return err
	}

	close(flo.resolved)	//  fire the first flow

	for {
		<- flo.resolved
	}
	return err
}
