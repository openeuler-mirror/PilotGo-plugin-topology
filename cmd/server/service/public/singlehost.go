/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package public

// import (
// 	"strings"

// 	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
// 	"gitee.com/openeuler/PilotGo-plugin-topology-server/processor"
//  "gitee.com/openeuler/PilotGo-plugin-topology/server/global"
// 	"github.com/pkg/errors"
// )

// func SingleHostService(uuid string) ([]*meta.Node, []*meta.Edge, []error, []error) {
// 	dataprocesser := processor.CreateDataProcesser()

// 	// TODO: 临时定义agentnum
// 	agentnum := 0
// 	nodes, edges, collect_errlist, process_errlist := dataprocesser.Process_data(agentnum)
// 	if len(collect_errlist) != 0 || len(process_errlist) != 0 {
// 		for i, cerr := range collect_errlist {
// 			collect_errlist[i] = errors.Wrap(cerr, " ")
// 		}

// 		for i, perr := range process_errlist {
// 			process_errlist[i] = errors.Wrap(perr, " ")
// 		}
// 	}

// 	single_nodes := []*meta.Node{}
// 	for _, node1 := range nodes.Nodes {
// 		if node1.UUID == uuid {
// 			repeat_node := false
// 			for _, node2 := range single_nodes {
// 				if node2.ID == node1.ID {
// 					repeat_node = true
// 				}
// 			}

// 			if !repeat_node {
// 				single_nodes = append(single_nodes, node1)
// 			}
// 		}
// 	}

// 	single_edges := []*meta.Edge{}
// 	for _, edge1 := range edges.Edges {
// 		if strings.Split(edge1.Src, global.NODE_CONNECTOR)[0] == uuid {
// 			repeat_edge := false
// 			for _, edge2 := range single_edges {
// 				if edge2.ID == edge1.ID {
// 					repeat_edge = true
// 				}
// 			}

// 			if !repeat_edge {
// 				single_edges = append(single_edges, edge1)
// 			}
// 		}
// 	}

// 	return single_nodes, single_edges, collect_errlist, process_errlist
// }

/*
	// if len(collect_errlist) != 0 && len(process_errlist) != 0 {
	// 	for i, cerr := range collect_errlist {
	// 		collect_errlist[i] = errors.Wrap(cerr, " ")
	// 	}

	// 	for i, perr := range process_errlist {
	// 		process_errlist[i] = errors.Wrap(perr, " ")
	// 	}

	// 	return nil, nil, collect_errlist, process_errlist
	// } else if len(collect_errlist) != 0 && len(process_errlist) == 0 {
	// 	for i, cerr := range collect_errlist {
	// 		collect_errlist[i] = errors.Wrap(cerr, " ")
	// 	}

	// 	return nil, nil, collect_errlist, nil
	// } else if len(collect_errlist) == 0 && len(process_errlist) != 0 {
	// 	for i, perr := range process_errlist {
	// 		process_errlist[i] = errors.Wrap(perr, " ")
	// 	}

	// 	return nil, nil, nil, process_errlist
	// }
*/
