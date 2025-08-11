//  Note: this code is have baked.  please ignore
package main

func stdio(root *ast) (err error) {

	flowA, err := compile(root) 
	if err != nil {
		return err
	}

	close(flowA.resolved)	//  fire the first flow

	for {
		<- flowA.resolved
	}
	return
}
