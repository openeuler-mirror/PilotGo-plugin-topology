package handler

import (
	"strconv"
	"strings"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/db/mysqlmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/graph"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/service"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/service/custom"
	"gitee.com/openeuler/PilotGo/sdk/response"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func CustomTopoListHandle(ctx *gin.Context) {
	query := &response.PaginationQ{}
	err := ctx.ShouldBindQuery(query)
	if err != nil {
		err = errors.New("failed to load parameters in url **errstack**2") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	if query.PageSize == 0 && query.Page == 0 {
		err := errors.New("query topo configuration list failed: page size and page can not be zero **errstack**1") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	tcs, total, err := custom.CustomTopoListService(query)
	if err != nil {
		switch strings.Split(errors.Cause(err).Error(), "**")[1] {
		case "errstack":
			err = errors.Wrap(err, " **errstack**2") // err top
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
			response.Fail(ctx, nil, errors.Cause(err).Error())
			return
		case "errstackfatal":
			err = errors.Wrap(err, " **errstackfatal**2") // err top
			response.Fail(ctx, nil, errors.Cause(err).Error())
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
			return
		}
	}

	response.DataPagination(ctx, tcs, total, query)
}

func CreateCustomTopoHandle(ctx *gin.Context) {
	var tc *mysqlmanager.Topo_configuration = new(mysqlmanager.Topo_configuration)

	if err := ctx.ShouldBindJSON(tc); err != nil {
		err = errors.Wrap(err, " **errstack**1") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	if tc.Name == "" && tc.BatchId == 0 && len(tc.NodeRules) == 0 && len(tc.TagRules) == 0 {
		err := errors.New("create topo configuration failed: topo configuration required **errstack**1") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	tcdb_id, err := custom.CreateCustomTopoService(tc)
	if err != nil {
		switch strings.Split(errors.Cause(err).Error(), "**")[1] {
		case "errstack":
			err = errors.Wrap(err, "**errstack**1") // err top
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
			response.Fail(ctx, nil, errors.Cause(err).Error())
			return
		case "errstackfatal":
			err = errors.Wrap(err, " **errstackfatal**2") // err top
			response.Fail(ctx, nil, errors.Cause(err).Error())
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
			return
		}
	}

	response.Success(ctx, tcdb_id, "successfully created action")
}

func UpdateCustomTopoHandle(ctx *gin.Context) {
	// var tc *meta.Topo_configuration = new(meta.Topo_configuration)
	req_body := struct {
		TC *mysqlmanager.Topo_configuration `json:"topo_configuration"`
		ID *uint                            `json:"id"`
	}{}

	// fmt.Printf("%+v\n", ctx.Request.Body)
	if err := ctx.ShouldBindJSON(&req_body); err != nil {
		err = errors.Wrap(err, "**errstack**1") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	if req_body.TC.Name == "" && req_body.TC.BatchId == 0 && len(req_body.TC.NodeRules) == 0 && len(req_body.TC.TagRules) == 0 {
		err := errors.New("update topo configuration failed: topo configuration required **errstack**1") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	tcdb_id, err := custom.UpdateCustomTopoService(req_body.TC, *req_body.ID)
	if err != nil {
		switch strings.Split(errors.Cause(err).Error(), "**")[1] {
		case "errstack":
			err = errors.Wrap(err, "**errstack**2") // err top
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
			response.Fail(ctx, nil, errors.Cause(err).Error())
			return
		case "errstackfatal":
			err = errors.Wrap(err, " **errstackfatal**2") // err top
			response.Fail(ctx, nil, errors.Cause(err).Error())
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
			return
		}
	}

	response.Success(ctx, tcdb_id, "successfully updated action")
}

func RunCustomTopoHandle(ctx *gin.Context) {
	// TODO: 执行业务之前先判断batch集群中的机器是否部署且运行topo-agent

	doneChan := make(chan *graph.TopoDataBuffer, 1)

	go func() {
		var custom_topodata *graph.TopoDataBuffer = new(graph.TopoDataBuffer)

		tcid_str := ctx.Query("id")
		if tcid_str == "" {
			err := errors.New("id is nil **errstack**2") // err top
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
			doneChan <- custom_topodata
			response.Fail(ctx, nil, errors.Cause(err).Error())
			return
		}
		custom_topodata.TopoConfId = tcid_str

		tcid_int, err := strconv.Atoi(tcid_str)
		if err != nil {
			err = errors.Wrap(err, "**errstack**2") // err top
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
			doneChan <- custom_topodata
			response.Fail(ctx, nil, errors.Cause(err).Error())
			return
		}

		custom_topodata.Nodes, custom_topodata.Edges, custom_topodata.Combos, err = custom.RunCustomTopoService(uint(tcid_int))
		if err != nil {
			if len(strings.Split(errors.Cause(err).Error(), "**")) == 0 {
				err = errors.Errorf("wrong err format: %s **warn**0", errors.Cause(err).Error())
				errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
				doneChan <- custom_topodata
				response.Fail(ctx, nil, errors.Cause(err).Error())
				return
			}
			switch strings.Split(errors.Cause(err).Error(), "**")[1] {
			case "errstack":
				err = errors.Wrap(err, " **errstack**2")
				errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
				doneChan <- custom_topodata
				response.Fail(ctx, nil, errors.Cause(err).Error())
				return
			case "errstackfatal":
				err = errors.Wrap(err, " **errstackfatal**2")
				doneChan <- custom_topodata
				response.Fail(ctx, nil, errors.Cause(err).Error())
				errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
				return
			}
		}
		if len(custom_topodata.Nodes.Nodes) == 0 || len(custom_topodata.Edges.Edges) == 0 {
			err := errors.New("nodes list is null or edges list is null **errstack**0")
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
			doneChan <- custom_topodata
			response.Fail(ctx, nil, errors.Cause(err).Error())
			return
		}

		service.UpdateGlobalTopoDataBuffer(custom_topodata)
		doneChan <- graph.Global_TopoDataBuffer
	}()

	select {
	case <-ctx.Request.Context().Done():
		return
	case res := <-doneChan:
		if len(res.Combos) != 0 && len(res.Edges.Edges) != 0 && len(res.Nodes.Nodes) != 0 {
			response.Success(ctx, map[string]interface{}{
				"nodes":  res.Nodes.Nodes,
				"edges":  res.Edges.Edges,
				"combos": res.Combos,
			}, "")
		}
	}
}

func DeleteCustomTopoHandle(ctx *gin.Context) {
	req_body := struct {
		IDs []uint `json:"ids"`
	}{}

	if err := ctx.ShouldBindJSON(&req_body); err != nil {
		err = errors.New(err.Error() + "**errstack**1") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	if err := custom.DeleteCustomTopoService(req_body.IDs); err != nil {
		switch strings.Split(errors.Cause(err).Error(), "**")[1] {
		case "errstack":
			err = errors.Wrap(err, "**errstack**1") // err top
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
			response.Fail(ctx, nil, errors.Cause(err).Error())
			return
		case "errstackfatal":
			err = errors.Wrap(err, " **errstackfatal**2") // err top
			response.Fail(ctx, nil, errors.Cause(err).Error())
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
			return
		}
	}

	response.Success(ctx, nil, "successfully deleted action")
}
