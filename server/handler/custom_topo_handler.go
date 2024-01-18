package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/service"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func CustomTopoListHandle(ctx *gin.Context) {
	custom_map := make(map[string]interface{})

	ctx.JSON(http.StatusOK, gin.H{
		"code":  0,
		"error": nil,
		"data":  custom_map,
	})
}

func CreateCustomTopoHandle(ctx *gin.Context) {
	var topoconfig *meta.Topo_configuration = new(meta.Topo_configuration)
	if err := ctx.ShouldBindJSON(topoconfig); err != nil {
		err = errors.Wrap(err, "**warn**1") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		return
	}

	machines_bytes, machines_err := json.Marshal(topoconfig.Machines)
	noderules_bytes, noderules_err := json.Marshal(topoconfig.NodeRules)
	tagrules_bytes, tagrules_err := json.Marshal(topoconfig.TagRules)
	if machines_err != nil || noderules_err != nil || tagrules_err != nil {
		err := errors.Errorf("json marshal error: machines(%s) noderules(%s) tagrules)%s **warn**4", machines_err, noderules_err, tagrules_err) // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		return
	}

	topoconfig.Machines = string(machines_bytes)
	topoconfig.NodeRules = string(noderules_bytes)
	topoconfig.TagRules = string(tagrules_bytes)

	if err := dao.Global_mysql.AddTopoConfiguration(topoconfig); err != nil {
		err = errors.Wrap(err, "**warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":  0,
		"error": nil,
		"data":  nil,
	})
}

func UpdateCustomTopoHandle(ctx *gin.Context) {

}

func UseCustomTopoHandle(ctx *gin.Context) {
	tcid_str := ctx.Query("id")
	if tcid_str == "" {
		err := errors.New("id is nil **warn**1") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":  -1,
			"error": fmt.Errorf("id is nil"),
			"data":  nil,
		})
		return
	}

	tcid_int, err := strconv.Atoi(tcid_str)
	if err != nil {
		err = errors.Wrap(err, "**warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		return
	}

	nodes, edges, combos, err := service.CustomTopoService(uint(tcid_int))
	if err != nil {
		err = errors.Wrap(err, " **warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		return
	}

	if len(nodes) == 0 || len(edges) == 0 {
		err := errors.New("nodes list is null or edges list is null **warn**0") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		ctx.JSON(http.StatusInternalServerError, gin.H{
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
			"nodes":  nodes,
			"edges":  edges,
			"combos": combos,
		},
	})
}

func DeleteCustomTopoHandle(ctx *gin.Context) {

}
