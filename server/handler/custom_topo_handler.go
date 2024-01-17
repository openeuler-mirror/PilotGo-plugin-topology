package handler

import (
	"encoding/json"
	"net/http"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func CreateCustomTopoHandle(ctx *gin.Context) {
	var topoconfig *meta.TopoConfiguration = new(meta.TopoConfiguration)
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
		"data":  nil,
	})
}

func UpdateCustomTopoHandle(ctx *gin.Context) {

}

func UseCustomTopoHandle(ctx *gin.Context) {

}

func DeleteCustomTopoHandle(ctx *gin.Context) {

}
