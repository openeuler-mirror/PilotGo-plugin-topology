package utils

import "gitee.com/openeuler/PilotGo-plugin-topology-server/meta"

func SplitEdgesByBreakpoint(arr []*meta.Edge, n int) [][]*meta.Edge {
	length := len(arr)
	if length == 0 {
		return nil
	}

	size := length / n
	result := make([][]*meta.Edge, n)

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

func SplitNodesByBreakpoint(arr []*meta.Node, n int) [][]*meta.Node {
	length := len(arr)
	if length == 0 {
		return nil
	}
	
	size := length / n
	result := make([][]*meta.Node, n)

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
