/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package graph

import (
	"strings"
	"sync"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/global"
	"github.com/pkg/errors"
)

type Edges struct {
	Lock           sync.Mutex
	Node_Edges_map sync.Map // key: node_id, value: []edge_id
	Lookup         sync.Map // key: edge_id, value: *Edge
	Edges          []*Edge
}

type Edge struct {
	DBID     int64               `json:"dbid"`
	ID       string              `json:"id"`
	Type     string              `json:"Type"`
	SrcID    int64               `json:"sourceid"`
	DstID    int64               `json:"targetid"`
	Src      string              `json:"source"`
	Dst      string              `json:"target"`
	Dir      string              `json:"dir"`
	Unixtime string              `json:"unixtime"`
	Tags     []string            `json:"tags"`
	Metrics  []map[string]string `json:"metrics"`
}

// 网络边镜像id检测：多个goruntine并发添加、访问、修改相同的edge实例
func (e *Edges) Add(edge *Edge) {
	if edge.Type == global.EDGE_TCP || edge.Type == global.EDGE_UDP {
		id_slice := strings.Split(edge.ID, global.EDGE_CONNECTOR)
		if len(id_slice) != 3 {
			global.ERManager.ErrorTransmit("graph", "error", errors.Errorf("can not generate mirror id of edge: %s, failed to add edge.", edge.ID), false, false)
			return
		}

		id_slice[0], id_slice[2] = id_slice[2], id_slice[0]
		mirror_id := strings.Join(id_slice, global.EDGE_CONNECTOR)

		e.Lock.Lock()
		if _, ok := e.Lookup.Load(mirror_id); !ok {
			e.Lookup.Store(edge.ID, edge)
			e.Edges = append(e.Edges, edge)

			src_edges_any, ok := e.Node_Edges_map.Load(edge.Src)
			if ok {
				src_edges := src_edges_any.([]string)
				e.Node_Edges_map.Store(edge.Src, append(src_edges, edge.ID))
			} else {
				e.Node_Edges_map.Store(edge.Src, []string{edge.ID})
			}
			dst_edges_any, ok := e.Node_Edges_map.Load(edge.Dst)
			if ok {
				dst_edges := dst_edges_any.([]string)
				e.Node_Edges_map.Store(edge.Dst, append(dst_edges, edge.ID))
			} else {
				e.Node_Edges_map.Store(edge.Dst, []string{edge.ID})
			}
		}
		e.Lock.Unlock()

		return
	}

	e.Lock.Lock()
	if _, ok := e.Lookup.LoadOrStore(edge.ID, edge); !ok {
		e.Edges = append(e.Edges, edge)

		src_edges_any, ok := e.Node_Edges_map.Load(edge.Src)
		if ok {
			src_edges := src_edges_any.([]string)
			e.Node_Edges_map.Store(edge.Src, append(src_edges, edge.ID))
		} else {
			e.Node_Edges_map.Store(edge.Src, []string{edge.ID})
		}
		dst_edges_any, ok := e.Node_Edges_map.Load(edge.Dst)
		if ok {
			dst_edges := dst_edges_any.([]string)
			e.Node_Edges_map.Store(edge.Dst, append(dst_edges, edge.ID))
		} else {
			e.Node_Edges_map.Store(edge.Dst, []string{edge.ID})
		}
	}
	e.Lock.Unlock()
}

func (e *Edges) Remove(id string) error {
	for i := 0; i < len(e.Edges); i++ {
		if e.Edges[i].ID != id {
			continue
		}
		// 从e.edges中移除边
		e.Edges = append(e.Edges[:i], e.Edges[i+1:]...)
		// 从e.lookup中移除边
		if _, ok := e.Lookup.LoadAndDelete(id); !ok {
			return errors.Errorf("edge %+v not fount in lookup sync.map", id)
		}
		// 从e.node_edges_map中移除边
		if _, ok := e.Node_Edges_map.LoadAndDelete(id); !ok {
			return errors.Errorf("edge %+v not fount in node_edges_map sync.map", id)
		}
		return nil
	}

	return errors.Errorf("edge %+v not fount in slice", id)
}
