package handler

import (
	"strconv"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/service"
	"gitee.com/openeuler/PilotGo/sdk/response"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func CustomTopoListHandle(ctx *gin.Context) {
	query := &response.PaginationQ{}
	err := ctx.ShouldBindQuery(query)
	if err != nil {
		err = errors.New("failed to load parameters in url **warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	tcs, total, err := service.CustomTopoListService(query)
	if err != nil {
		err = errors.Wrap(err, "**warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	response.DataPagination(ctx, tcs, total, query)
}

func CreateCustomTopoHandle(ctx *gin.Context) {
	var tc *meta.Topo_configuration = new(meta.Topo_configuration)

	if err := ctx.ShouldBindJSON(tc); err != nil {
		err = errors.Wrap(err, "**warn**1") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	tcdb_id, err := service.CreateCustomTopoService(tc)
	if err != nil {
		err = errors.Wrap(err, "**warn**1") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	response.Success(ctx, tcdb_id, "successfully created action")
}

func UpdateCustomTopoHandle(ctx *gin.Context) {
	tcid_str := ctx.Query("id")
	if tcid_str == "" {
		err := errors.New("id is nil **warn**1") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	tcid_int, err := strconv.Atoi(tcid_str)
	if err != nil {
		err = errors.Wrap(err, "**warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	var tc *meta.Topo_configuration = new(meta.Topo_configuration)
	if err := ctx.ShouldBindJSON(tc); err != nil {
		err = errors.Wrap(err, "**warn**1") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	tcdb, err := dao.Global_mysql.TopoConfigurationToDB(tc)
	if err != nil {
		err = errors.Wrap(err, "**warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	if err := dao.Global_mysql.DeleteTopoConfiguration(uint(tcid_int)); err != nil {
		err = errors.Wrap(err, "**warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	tcdb.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
	tcdb_id, err := dao.Global_mysql.AddTopoConfiguration(tcdb)
	if err != nil {
		err = errors.Wrap(err, "**warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	response.Success(ctx, tcdb_id, "successfully updated action")
}

func RunCustomTopoHandle(ctx *gin.Context) {
	// TODO: 执行业务之前先判断batch集群中的机器是否部署且运行topo-agent

	tcid_str := ctx.Query("id")
	if tcid_str == "" {
		err := errors.New("id is nil **warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	tcid_int, err := strconv.Atoi(tcid_str)
	if err != nil {
		err = errors.Wrap(err, "**warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	nodes, edges, combos, err := service.RunCustomTopoService(uint(tcid_int))
	if err != nil {
		err = errors.Wrap(err, " **warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	if len(nodes) == 0 || len(edges) == 0 {
		err := errors.New("nodes list is null or edges list is null **warn**0") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	response.Success(ctx, map[string]interface{}{
		"nodes":  nodes,
		"edges":  edges,
		"combos": combos,
	}, "")
}

func DeleteCustomTopoHandle(ctx *gin.Context) {
	tcid_str := ctx.Query("id")
	if tcid_str == "" {
		err := errors.New("id is nil **warn**1") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	tcid_int, err := strconv.Atoi(tcid_str)
	if err != nil {
		err = errors.Wrap(err, "**warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	if err := dao.Global_mysql.DeleteTopoConfiguration(uint(tcid_int)); err != nil {
		err = errors.Wrap(err, "**warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	response.Success(ctx, nil, "successfully deleted action")
}