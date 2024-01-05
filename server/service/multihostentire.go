package service

// import (
// 	"fmt"

// 	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
// 	"gitee.com/openeuler/PilotGo-plugin-topology-server/processor"
// 	"github.com/pkg/errors"
// )

// func MultiHostEntireService() ([]*meta.Node, []*meta.Edge, []error, []error) {
// 	dataprocesser := processor.CreateDataProcesser()

// 	// TODO: 临时定义agentnum
// 	agentnum := 0
// 	nodes, edges, collect_errlist, process_errlist := dataprocesser.Process_data(agentnum)
// 	if len(collect_errlist) != 0 || len(process_errlist) != 0 {
// 		for i, cerr := range collect_errlist {
// 			collect_errlist[i] = errors.Wrap(cerr, "**3")
// 		}

// 		for i, perr := range process_errlist {
// 			process_errlist[i] = errors.Wrap(perr, "**7")
// 		}
// 	}

// 	hostids := []string{}
// 	multi_nodes_map := make(map[string]*meta.Node)
// 	multi_nodes := []*meta.Node{}
// 	multi_edges_map := make(map[string]*meta.Edge)
// 	multi_edges := []*meta.Edge{}

// 	// 添加 host node
// 	for _, node := range nodes.Nodes {
// 		if node.Type == "host" {
// 			if _, ok := multi_nodes_map[node.ID]; !ok {
// 				multi_nodes_map[node.ID] = node
// 				multi_nodes = append(multi_nodes, node)
// 			}

// 			hostids = append(hostids, node.ID)
// 		}
// 	}

// 	for _, edge := range edges.Edges {
// 		if edge.Type == "tcp" || edge.Type == "udp" {
// 			if _, ok := multi_edges_map[edge.ID]; !ok {
// 				multi_edges_map[edge.ID] = edge
// 				multi_edges = append(multi_edges, edge)
// 			}

// 			if _, ok := multi_nodes_map[nodes.Lookup[edge.Src].ID]; !ok {
// 				multi_nodes_map[nodes.Lookup[edge.Src].ID] = nodes.Lookup[edge.Src]
// 				multi_nodes = append(multi_nodes, nodes.Lookup[edge.Src])
// 			}

// 			if _, ok := multi_nodes_map[nodes.Lookup[edge.Dst].ID]; !ok {
// 				multi_nodes_map[nodes.Lookup[edge.Dst].ID] = nodes.Lookup[edge.Dst]
// 				multi_nodes = append(multi_nodes, nodes.Lookup[edge.Dst])
// 			}
// 		} else if edge.Type == "server" || edge.Type == "client" {
// 			if _, ok := multi_edges_map[edge.ID]; !ok {
// 				multi_edges_map[edge.ID] = edge
// 				multi_edges = append(multi_edges, edge)
// 			}

// 			if _, ok := multi_nodes_map[nodes.Lookup[edge.Src].ID]; !ok {
// 				multi_nodes_map[nodes.Lookup[edge.Src].ID] = nodes.Lookup[edge.Src]
// 				multi_nodes = append(multi_nodes, nodes.Lookup[edge.Src])
// 			}

// 			if _, ok := multi_nodes_map[nodes.Lookup[edge.Dst].ID]; !ok {
// 				multi_nodes_map[nodes.Lookup[edge.Dst].ID] = nodes.Lookup[edge.Dst]
// 				multi_nodes = append(multi_nodes, nodes.Lookup[edge.Dst])
// 			}

// 			// 添加process-to-process、process-to-host edge
// 			start_nodeid := edge.Dst
// 			for {
// 				if nodes.Lookup[start_nodeid].Metrics["Ppid"] != "1" {
// 					// 添加 process-to-process edge
// 					edgeid := start_nodeid + "_belong_" + nodes.Lookup[start_nodeid].UUID + "_process_" + nodes.Lookup[start_nodeid].Metrics["Ppid"]
// 					edge1, ok := edges.Lookup.Load(edgeid)
// 					if !ok {
// 						fmt.Printf("%+v\n", errors.Errorf("faild to load edge from edges.lookup: %s**2", edgeid))
// 					}

// 					if _, ok := multi_edges_map[edge1.(*meta.Edge).ID]; !ok {
// 						multi_edges_map[edge1.(*meta.Edge).ID] = edge1.(*meta.Edge)
// 						multi_edges = append(multi_edges, edge1.(*meta.Edge))
// 					}

// 					// 添加 target process node
// 					target_node := nodes.Lookup[edge1.(*meta.Edge).Dst]

// 					if _, ok := multi_nodes_map[target_node.ID]; !ok {
// 						multi_nodes_map[target_node.ID] = target_node
// 						multi_nodes = append(multi_nodes, target_node)
// 					}

// 					start_nodeid = nodes.Lookup[start_nodeid].UUID + "_process_" + nodes.Lookup[start_nodeid].Metrics["Ppid"]

// 					continue
// 				}

// 				// 添加 process-to-1 edge
// 				edgeid_to_1 := start_nodeid + "_belong_" + nodes.Lookup[start_nodeid].UUID + "_process_" + nodes.Lookup[start_nodeid].Metrics["Ppid"]
// 				edge_to_1, ok := edges.Lookup.Load(edgeid_to_1)
// 				if !ok {
// 					fmt.Printf("%+v\n", errors.Errorf("faild to load edge from edges.lookup: %s**2", edgeid_to_1))
// 				}

// 				if _, ok := multi_edges_map[edge_to_1.(*meta.Edge).ID]; !ok {
// 					multi_edges_map[edge_to_1.(*meta.Edge).ID] = edge_to_1.(*meta.Edge)
// 					multi_edges = append(multi_edges, edge_to_1.(*meta.Edge))
// 				}

// 				// 添加 1-to-host edge
// 				start_nodeid = nodes.Lookup[start_nodeid].UUID + "_process_" + nodes.Lookup[start_nodeid].Metrics["Ppid"]
// 				edgeid_to_host := ""
// 				for _, hostid := range hostids {
// 					if nodes.Lookup[start_nodeid].UUID == nodes.Lookup[hostid].UUID {
// 						edgeid_to_host = start_nodeid + "_belong_" + nodes.Lookup[hostid].ID
// 						break
// 					}
// 				}

// 				edge_to_host, ok := edges.Lookup.Load(edgeid_to_host)
// 				if !ok {
// 					fmt.Printf("%+v\n", errors.Errorf("faild to load edge from edges.lookup: %s**2", edgeid_to_host))
// 				}

// 				if _, ok := multi_edges_map[edge_to_host.(*meta.Edge).ID]; !ok {
// 					multi_edges_map[edge_to_host.(*meta.Edge).ID] = edge_to_host.(*meta.Edge)
// 					multi_edges = append(multi_edges, edge_to_host.(*meta.Edge))
// 				}

// 				// 添加 process 1 node
// 				target_node := nodes.Lookup[start_nodeid]

// 				if _, ok := multi_nodes_map[target_node.ID]; !ok {
// 					multi_nodes_map[target_node.ID] = target_node
// 					multi_nodes = append(multi_nodes, target_node)
// 				}

// 				break
// 			}

// 		}
// 	}

// 	return multi_nodes, multi_edges, collect_errlist, process_errlist
// }
