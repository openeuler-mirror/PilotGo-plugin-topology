package handler

import (
	"net/http"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/service"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func SingleHostHandle(ctx *gin.Context) {
	agentmanager.Topo.UpdateMachineList()

	uuid := ctx.Param("uuid")
	nodes, edges, collect_errlist, process_errlist := service.SingleHostService(uuid)

	if len(collect_errlist) != 0 || len(process_errlist) != 0 {
		for i, cerr := range collect_errlist {
			collect_errlist[i] = errors.Wrap(cerr, "**warn**4") // err top
			agentmanager.Topo.ErrCh <- collect_errlist[i]
		}

		for i, perr := range process_errlist {
			process_errlist[i] = errors.Wrap(perr, "**warn**10") // err top
			agentmanager.Topo.ErrCh <- process_errlist[i]
		}
	}

	if len(nodes) == 0 || len(edges) == 0 {
		err := errors.New("nodes list is null or edges list is null **warn**0") // err top
		agentmanager.Topo.ErrCh <- err

		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":  0,
		"error": nil,
		"data": map[string]interface{}{
			"nodes": nodes,
			"edges": edges,
		},
	})
}

func SingleHostTreeHandle(ctx *gin.Context) {
	agentmanager.Topo.UpdateMachineList()

	uuid := ctx.Param("uuid")
	nodes, collect_errlist, process_errlist := service.SingleHostTreeService(uuid)

	if len(collect_errlist) != 0 || len(process_errlist) != 0 {
		for i, cerr := range collect_errlist {
			collect_errlist[i] = errors.Wrap(cerr, "**warn**4") // err top
			agentmanager.Topo.ErrCh <- collect_errlist[i]
		}

		for i, perr := range process_errlist {
			process_errlist[i] = errors.Wrap(perr, "**warn**10") // err top
			agentmanager.Topo.ErrCh <- process_errlist[i]
		}
	}

	if nodes == nil {
		err := errors.New("node tree is null **warn**0") // err top
		agentmanager.Topo.ErrCh <- err

		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":  0,
		"error": nil,
		"data": map[string]interface{}{
			"tree": nodes,
		},
	})
}

func MultiHostHandle(ctx *gin.Context) {
	agentmanager.Topo.UpdateMachineList()
	
	nodes, edges, collect_errlist, process_errlist := service.MultiHostService()

	if len(collect_errlist) != 0 || len(process_errlist) != 0 {
		for i, cerr := range collect_errlist {
			collect_errlist[i] = errors.Wrap(cerr, "**warn**4") // err top
			agentmanager.Topo.ErrCh <- collect_errlist[i]
		}

		for i, perr := range process_errlist {
			process_errlist[i] = errors.Wrap(perr, "**warn**10") // err top
			agentmanager.Topo.ErrCh <- process_errlist[i]
		}
	}

	if len(nodes) == 0 || len(edges) == 0 {
		err := errors.New("nodes list is null or edges list is null **warn**0") // err top
		agentmanager.Topo.ErrCh <- err

		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":  0,
		"error": nil,
		"data": map[string]interface{}{
			"nodes": nodes,
			"edges": edges,
		},
	})
}

func AgentListHandle(ctx *gin.Context) {
	agentmanager.Topo.UpdateMachineList()

	agentmap := make(map[string]string)
	agentmanager.Topo.AgentMap.Range(func(key, value interface{}) bool {
		agent := value.(*agentmanager.Agent_m)
		if agent.Host_2 != nil {
			agentmap[agent.UUID] = agent.IP + ":" + agent.Port
		}

		return true
	})

	ctx.JSON(http.StatusOK, gin.H{
		"code":  0,
		"error": nil,
		"data": map[string]interface{}{
			"agentlist": agentmap,
		},
	})
}
