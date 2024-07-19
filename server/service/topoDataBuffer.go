package service

import (
	"strconv"
	"sync"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/graph"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"
	"github.com/pkg/errors"
)

// 在parent_pid的子进程树中搜索target_pid
func searchTargetPid(process_map map[int32][]int32, parent_pid, target_pid int32) bool {
	find_target := false
	for _, sub_pid := range process_map[parent_pid] {
		if sub_pid == target_pid {
			find_target = true
			break
		}
		find_target = searchTargetPid(process_map, sub_pid, target_pid)
		if find_target {
			return true
		}
	}
	return find_target
}

// 判断两个process node是否属于同一个应用
func IsSamePstreeBranch(old_node, new_node *graph.Node, new_process_slice []*graph.Process) bool {
	var small_pid, big_pid int32
	old_node_pid, _ := strconv.Atoi(old_node.Metrics["Pid"])
	new_node_pid, _ := strconv.Atoi(new_node.Metrics["Pid"])
	if old_node_pid > new_node_pid {
		small_pid = int32(new_node_pid)
		big_pid = int32(old_node_pid)
	} else {
		small_pid = int32(old_node_pid)
		big_pid = int32(new_node_pid)
	}

	new_process_map := make(map[int32][]int32)
	for _, process := range new_process_slice {
		new_process_map[int32(process.Pid)] = process.Cpid
	}
	return (old_node.Metrics["Ppid"] == new_node.Metrics["Ppid"] || searchTargetPid(new_process_map, small_pid, big_pid))
}

// 更新全局图数据缓存
func UpdateGlobalTopoDataBuffer(custom_topodata *graph.TopoDataBuffer) {
	if graph.Global_TopoDataBuffer == nil || graph.Global_TopoDataBuffer.TopoConfId != custom_topodata.TopoConfId {
		graph.Global_TopoDataBuffer = custom_topodata
	} else {
		var wg sync.WaitGroup
		for uuid, global_node_slice := range graph.Global_TopoDataBuffer.Nodes.LookupByUUID {
			wg.Add(1)
			go func(_uuid string, _global_node_slice []*graph.Node) {
				defer wg.Done()
				// 初始化最新图数据中节点的待匹配状态，若未在缓存图数据中发现对应节点，则该节点为新增节点
				custom_node_matched_state_map := make(map[string]bool)
				for _, custom_node := range custom_topodata.Nodes.LookupByUUID[_uuid] {
					custom_node_matched_state_map[custom_node.ID] = false
				}
				// 初始化缓存图数据中节点的待匹配状态，若未在新图数据中发现对应节点，则该节点为删减节点
				global_node_matched_state_map := make(map[string]bool)
				for _, global_node := range _global_node_slice {
					global_node_matched_state_map[global_node.ID] = false
				}
				// 用新图数据更新缓存图数据
				for _, global_node := range _global_node_slice {
					for _, custom_node := range custom_topodata.Nodes.LookupByUUID[_uuid] {
						if global_node.Name == custom_node.Name {
							if global_node.Metrics["Pid"] == custom_node.Metrics["Pid"] || IsSamePstreeBranch(global_node, custom_node, agentmanager.Global_AgentManager.GetAgent_T(_uuid).Processes_2) {
								// 更新节点数据
								global_node.LayoutAttr = custom_node.LayoutAttr
								global_node.Metrics = custom_node.Metrics
								global_node.Network = custom_node.Network
								global_node.Tags = custom_node.Tags
								global_node.Unixtime = custom_node.Unixtime
								// 更新边数据
								custom_edge_id_slice_any, ok := custom_topodata.Edges.Node_Edges_map.Load(custom_node.ID)
								if !ok {
									continue
								}
								for _, custom_edge_id := range custom_edge_id_slice_any.([]string) {
									custom_edge_any, ok := custom_topodata.Edges.Lookup.Load(custom_edge_id)
									if !ok {
										continue
									}
									custom_edge := custom_edge_any.(*graph.Edge)

									global_edge_id_slice, ok := graph.Global_TopoDataBuffer.Edges.Node_Edges_map.Load(global_node.ID)
									if !ok {
										continue
									}
									for _, global_edge_id := range global_edge_id_slice.([]string) {
										global_edge_any, ok := graph.Global_TopoDataBuffer.Edges.Lookup.Load(global_edge_id)
										if !ok {
											continue
										}
										global_edge := global_edge_any.(*graph.Edge)
										if custom_edge.Type == global_edge.Type {
											global_edge.Dir = custom_edge.Dir
											global_edge.DBID = custom_edge.DBID
											global_edge.DstID = custom_edge.DstID
											global_edge.Metrics = custom_edge.Metrics
											global_edge.SrcID = custom_edge.SrcID
											global_edge.Tags = custom_edge.Tags
											global_edge.Unixtime = custom_edge.Unixtime
											break
										}
									}
								}
								// 更新待匹配状态
								custom_node_matched_state_map[custom_node.ID] = true
								global_node_matched_state_map[global_node.ID] = true

								break
							}
						}
					}
				}
				// 将新图数据中的新增节点及边添加到缓存图数据中
				for _, custom_node := range custom_topodata.Nodes.LookupByUUID[_uuid] {
					if !custom_node_matched_state_map[custom_node.ID] {
						graph.Global_TopoDataBuffer.Nodes.Add(custom_node)
						custom_edge_id_slice_any, ok := custom_topodata.Edges.Node_Edges_map.Load(custom_node.ID)
						if !ok {
							continue
						}
						for _, custom_edge_id := range custom_edge_id_slice_any.([]string) {
							custom_edge_any, ok := custom_topodata.Edges.Lookup.Load(custom_edge_id)
							if !ok {
								continue
							}
							custom_edge := custom_edge_any.(*graph.Edge)
							graph.Global_TopoDataBuffer.Edges.Add(custom_edge)
						}
					}
				}
				// 删减缓存图数据中过期的节点及边
				for _, global_node := range _global_node_slice {
					if !global_node_matched_state_map[global_node.ID] {
						err := graph.Global_TopoDataBuffer.Nodes.Remove(global_node)
						if err != nil {
							err = errors.Wrap(err, "->")
							errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
							continue
						}
						global_edge_id_slice_any, ok := graph.Global_TopoDataBuffer.Edges.Node_Edges_map.Load(global_node.ID)
						if !ok {
							continue
						}
						for _, global_edge_id := range global_edge_id_slice_any.([]string) {
							global_edge_any, ok := graph.Global_TopoDataBuffer.Edges.Lookup.Load(global_edge_id)
							if !ok {
								continue
							}
							global_edge := global_edge_any.(*graph.Edge)
							err := graph.Global_TopoDataBuffer.Edges.Remove(global_edge.ID)
							if err != nil {
								err = errors.Wrap(err, "->")
								errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
							}
						}
					}
				}

			}(uuid, global_node_slice)
		}
		wg.Wait()
	}
}
