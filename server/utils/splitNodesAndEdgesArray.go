package utils

import "gitee.com/openeuler/PilotGo-plugin-topology-server/graph"

func SplitEdgesByBreakpoint(arr []*graph.Edge, n int) [][]*graph.Edge {
	length := len(arr)
	if length == 0 {
		return nil
	}

	size := length / n
	result := make([][]*graph.Edge, n)

	for i := 0; i < n; i++ {
		start := i * size
		end := (i + 1) * size

		if end > length {
			end = length
		}

		result = append(result, arr[start:end])
	}

	return result
}

func SplitNodesByBreakpoint(arr []*graph.Node, n int) [][]*graph.Node {
	length := len(arr)
	if length == 0 {
		return nil
	}
	
	size := length / n
	result := make([][]*graph.Node, n)

	for i := 0; i < n; i++ {
		start := i * size
		end := (i + 1) * size

		if end > length {
			end = length
		}

		result = append(result, arr[start:end])
	}

	return result
}
