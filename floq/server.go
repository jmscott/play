package main

func server(root *ast) error {

	flo := &flow{}
	flo, _, err := compile(root) 
	if err != nil {
		return err
	}
	close(flo.resolved)

	return err
}
