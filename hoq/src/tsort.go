//  toplogically sort a directed graph using the kahn algorithm.
//  return the order or nil.
//
//	https://en.wikipedia.org/wiki/Topological_sorting#Kahn.27s_algorithm
//
//  the graph is a list of edges.  each edge is a pair of nodes separated by a
//  space character.
//
//  eventually need to return a cycle, to aid debuging.

package main

import "strings"

func tsort(graph []string) (order []string) {

	edge := make(map[string][]string)
	node := make(map[string]bool)
	inbound := make(map[string]uint64)
	root := make(map[string]bool)

	//  build the node{}, edge{}, and inbound{} maps of graph

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

	//  build the root{} map

	for n := range node {
		if inbound[n] == 0 {
			root[n] = true
		}
	}

	visited := 0
	order = make([]string, 0)

	//  while the root set has elements
	//	select any root element, say r1
	//	delete r1 from root set
	//	add r1 to order list
	//	increment count of visited nodes
	//	visit each target, say tN, of r1
	//		decrement tN inbound node count
	//		add tN to root set if inbound count <= 0
	//  ordered if visited count == node count

	for len(root) > 0 {
		var r1 string

		for r1 = range root {
			break
		}
		delete(root, r1)

		order = append(order, r1)
		visited++

		//  have any of the nodes that r1 points to themselves become
		//  roots?  if so, then add them to the root set

		for _, tN := range edge[r1] {
			inbound[tN]--
			if inbound[tN] <= 0 {
				root[tN] = true
			}
		}
	}
	if visited == len(node) {
		return order
	}
	return nil
}
