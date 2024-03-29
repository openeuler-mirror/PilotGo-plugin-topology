package handler

import (
	"strconv"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
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
		err = errors.New("failed to load parameters in url **errstack**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		response.Fail(ctx, nil, err.Error())
		return
	}

	tcs, total, err := service.CustomTopoListService(query)
	if err != nil {
		err = errors.Wrap(err, "**errstack**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		response.Fail(ctx, nil, err.Error())
		return
	}

	response.DataPagination(ctx, tcs, total, query)
}

func CreateCustomTopoHandle(ctx *gin.Context) {
	var tc *meta.Topo_configuration = new(meta.Topo_configuration)

	if err := ctx.ShouldBindJSON(tc); err != nil {
		err = errors.Wrap(err, "**errstack**1") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		response.Fail(ctx, nil, err.Error())
		return
	}

	tcdb_id, err := service.CreateCustomTopoService(tc)
	if err != nil {
		err = errors.Wrap(err, "**errstack**1") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		response.Fail(ctx, nil, err.Error())
		return
	}

	response.Success(ctx, tcdb_id, "successfully created action")
}

func UpdateCustomTopoHandle(ctx *gin.Context) {
	// var tc *meta.Topo_configuration = new(meta.Topo_configuration)
	req_body := struct {
		TC *meta.Topo_configuration `json:"topo_configuration"`
		ID *uint                    `json:"id"`
	}{}

	// fmt.Printf("%+v\n", ctx.Request.Body)
	if err := ctx.ShouldBindJSON(&req_body); err != nil {
		err = errors.Wrap(err, "**errstack**1") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		response.Fail(ctx, nil, err.Error())
		return
	}

	tcdb_id, err := service.UpdateCustomTopoService(req_body.TC, *req_body.ID)
	if err != nil {
		err = errors.Wrap(err, "**errstack**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		response.Fail(ctx, nil, err.Error())
		return
	}

	response.Success(ctx, tcdb_id, "successfully updated action")
}

func RunCustomTopoHandle(ctx *gin.Context) {
	// TODO: 执行业务之前先判断batch集群中的机器是否部署且运行topo-agent
	type alldata struct {
		Nodes  []*meta.Node
		Edges  []*meta.Edge
		Combos []map[string]string
	}
	doneChan := make(chan *alldata, 1)

	go func() {
		var topodata *alldata = new(alldata)

		tcid_str := ctx.Query("id")
		if tcid_str == "" {
			err := errors.New("id is nil **errstack**2") // err top
			agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
			response.Fail(ctx, nil, err.Error())
			return
		}

		tcid_int, err := strconv.Atoi(tcid_str)
		if err != nil {
			err = errors.Wrap(err, "**errstack**2") // err top
			agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
			response.Fail(ctx, nil, err.Error())
			return
		}

		topodata.Nodes, topodata.Edges, topodata.Combos, err = service.RunCustomTopoService(uint(tcid_int))
		if err != nil {
			err = errors.Wrap(err, " **errstack**2") // err top
			agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
			response.Fail(ctx, nil, err.Error())
			return
		}

		if len(topodata.Nodes) == 0 || len(topodata.Edges) == 0 {
			err := errors.New("nodes list is null or edges list is null **errstack**0") // err top
			agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
			response.Fail(ctx, nil, err.Error())
			return
		}

		doneChan <- topodata
	}()

	select {
	case <-ctx.Request.Context().Done():
		return
	case res := <-doneChan:
		response.Success(ctx, map[string]interface{}{
			"nodes":  res.Nodes,
			"edges":  res.Edges,
			"combos": res.Combos,
		}, "")
	}
}

func DeleteCustomTopoHandle(ctx *gin.Context) {
	req_body := struct {
		IDs []uint `json:"ids"`
	}{}

	if err := ctx.ShouldBindJSON(&req_body); err != nil {
		err = errors.New(err.Error() + "**errstack**1") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		response.Fail(ctx, nil, err.Error())
		return
	}

	if err := service.DeleteCustomTopoService(req_body.IDs); err != nil {
		err = errors.Wrap(err, "**errstack**1") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		response.Fail(ctx, nil, err.Error())
		return
	}

	response.Success(ctx, nil, "successfully deleted action")
}
