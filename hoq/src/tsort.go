//  toplogically sort graph using the kahn algorithm.
//
//	https://en.wikipedia.org/wiki/Topological_sorting#Kahn.27s_algorithm
//

package main

import "strings"

func tsort(graph []string) (order []string) {

	edge := make(map[string][]string)
	node := make(map[string]bool)
	inbound := make(map[string]uint64)
	root := make(map[string]bool)

	for _, e := range graph {

		pair := strings.Split(e, " ")

		source := pair[0]
		target := pair[1]
		
		node[source] = true
		node[target] = true
		if source != target {
			edge[source] = append(edge[source], target)
			inbound[target]++
		}
	}
	for n := range node {
		if inbound[n] == 0 {
			root[n] = true
		}
	}

	visited := uint64(0)
	order = make([]string, 0)

	for len(root) > 0 {
		var r string

		//  delete an arbitrary element from root set 

		for r = range root {
			break
		}
		delete(root, r)

		order = append(order, r)
		visited++
		for _, t := range edge[r] {
			inbound[t]--
			if inbound[t] == 0 {
				root[t] = true
			}
		}
	}
	if visited == uint64(len(node)) {
		return order
	}
	return nil
}
