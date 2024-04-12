package graph

type TreeTopoNode struct {
	ID       string          `json:"id"`
	Node     *Node           `json:"node"`
	Children []*TreeTopoNode `json:"children"`
}

func CreateTreeNode(node *Node) *TreeTopoNode {
	return &TreeTopoNode{
		ID:       node.ID,
		Node:     node,
		Children: make([]*TreeTopoNode, 0),
	}
}

func SliceToTree(process_nodes []*TreeTopoNode, net_nodes []*TreeTopoNode, ppid string) []*TreeTopoNode {
	newarr := make([]*TreeTopoNode, 0)

	for _, node := range process_nodes {
		if node.Node.Metrics["Ppid"] == ppid {
			node.Children = SliceToTree(process_nodes, net_nodes, node.Node.Metrics["Pid"])
			newarr = append(newarr, node)
		}
	}

	for _, netnode := range net_nodes {
		if netnode.Node.Metrics["Pid"] == ppid {
			newarr = append(newarr, netnode)
		}
	}

	return newarr
}
