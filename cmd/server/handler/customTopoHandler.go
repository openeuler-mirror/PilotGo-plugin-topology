package handler

import (
	"strconv"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/db/mysqlmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/graph"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/resourcemanage"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/service/custom"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/service/webclient"
	"gitee.com/openeuler/PilotGo/sdk/response"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func CustomTopoListHandle(ctx *gin.Context) {
	query := &response.PaginationQ{}
	err := ctx.ShouldBindQuery(query)
	if err != nil {
		err = errors.New("failed to load parameters in url")
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	if query.PageSize == 0 && query.Page == 0 {
		err := errors.New("query topo configuration list failed: page size and page can not be zero")
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	tcs, total, err, exit := custom.CustomTopoListService(query)
	if err != nil {
		if exit {
			err = errors.Wrap(err, " ")
			response.Fail(ctx, nil, errors.Cause(err).Error())
			resourcemanage.ERManager.ErrorTransmit("error", err, true, true)
			return
		}
		err = errors.Wrap(err, " ")
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	response.DataPagination(ctx, tcs, total, query)
}

func CreateCustomTopoHandle(ctx *gin.Context) {
	var tc *mysqlmanager.Topo_configuration = new(mysqlmanager.Topo_configuration)

	if err := ctx.ShouldBindJSON(tc); err != nil {
		err = errors.Wrap(err, " ")
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	if tc.Name == "" && tc.BatchId == 0 && len(tc.NodeRules) == 0 && len(tc.TagRules) == 0 {
		err := errors.New("create topo configuration failed: topo configuration required")
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	tcdb_id, err, exit := custom.CreateCustomTopoService(tc)
	if err != nil {
		if exit {
			err = errors.Wrap(err, " ")
			response.Fail(ctx, nil, errors.Cause(err).Error())
			resourcemanage.ERManager.ErrorTransmit("error", err, true, true)
			return
		}
		err = errors.Wrap(err, " ")
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
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
		err = errors.Wrap(err, " ")
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	if req_body.TC.Name == "" && req_body.TC.BatchId == 0 && len(req_body.TC.NodeRules) == 0 && len(req_body.TC.TagRules) == 0 {
		err := errors.New("update topo configuration failed: topo configuration required")
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	tcdb_id, err, exit := custom.UpdateCustomTopoService(req_body.TC, *req_body.ID)
	if err != nil {
		if exit {
			err = errors.Wrap(err, " ")
			response.Fail(ctx, nil, errors.Cause(err).Error())
			resourcemanage.ERManager.ErrorTransmit("error", err, true, true)
			return
		}
		err = errors.Wrap(err, " ")
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	response.Success(ctx, tcdb_id, "successfully updated action")
}

func RunCustomTopoHandle(ctx *gin.Context) {
	// TODO: 执行业务之前先判断batch集群中的机器是否部署且运行topo-agent

	tcid_str := ctx.Query("id")
	webclient_id := ctx.Query("clientId")

	doneChan := make(chan *graph.TopoDataBuffer, 1)

	go func() {
		var custom_topodata *graph.TopoDataBuffer = new(graph.TopoDataBuffer)

		if tcid_str == "" {
			err := errors.New("id is nil")
			resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
			doneChan <- custom_topodata
			response.Fail(ctx, nil, errors.Cause(err).Error())
			return
		}
		custom_topodata.TopoConfId = tcid_str

		tcid_int, err := strconv.Atoi(tcid_str)
		if err != nil {
			err = errors.Wrap(err, " ")
			resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
			doneChan <- custom_topodata
			response.Fail(ctx, nil, errors.Cause(err).Error())
			return
		}

		exit := false
		custom_topodata.Nodes, custom_topodata.Edges, custom_topodata.Combos, err, exit = custom.RunCustomTopoService(uint(tcid_int))
		if err != nil {
			if exit {
				err = errors.Wrap(err, " ")
				doneChan <- custom_topodata
				response.Fail(ctx, nil, errors.Cause(err).Error())
				resourcemanage.ERManager.ErrorTransmit("error", err, true, true)
				return
			}
			err = errors.Wrap(err, " ")
			resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
			doneChan <- custom_topodata
			response.Fail(ctx, nil, errors.Cause(err).Error())
			return
		}
		if len(custom_topodata.Nodes.Nodes) == 0 || len(custom_topodata.Edges.Edges) == 0 {
			err := errors.New("nodes list is null or edges list is null")
			resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
			doneChan <- custom_topodata
			response.Fail(ctx, nil, errors.Cause(err).Error())
			return
		}

		webclient.WebClientsManager.UpdateClientTopoDataBuffer(webclient_id, custom_topodata)
		
		doneChan <- webclient.WebClientsManager.Get(webclient_id)
	}()

	select {
	case <-ctx.Request.Context().Done():
		return
	case data := <-doneChan:
		if data == nil {
			err := errors.Errorf("topodatabuff is nill, client: %s, %s", ctx.Request.RemoteAddr, webclient_id)
			resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
			response.Fail(ctx, nil, err.Error())
			return
		}
		if len(data.Combos) != 0 && len(data.Edges.Edges) != 0 && len(data.Nodes.Nodes) != 0 {
			response.Success(ctx, map[string]interface{}{
				"nodes":  data.Nodes.Nodes,
				"edges":  data.Edges.Edges,
				"combos": data.Combos,
			}, "")
		}
	}
}

func DeleteCustomTopoHandle(ctx *gin.Context) {
	req_body := struct {
		IDs []uint `json:"ids"`
	}{}

	if err := ctx.ShouldBindJSON(&req_body); err != nil {
		err = errors.New(err.Error())
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	if err, exit := custom.DeleteCustomTopoService(req_body.IDs); err != nil {
		if exit {
			err = errors.Wrap(err, " ")
			response.Fail(ctx, nil, errors.Cause(err).Error())
			resourcemanage.ERManager.ErrorTransmit("error", err, true, true)
			return
		}
		err = errors.Wrap(err, " ")
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}

	response.Success(ctx, nil, "successfully deleted action")
}
