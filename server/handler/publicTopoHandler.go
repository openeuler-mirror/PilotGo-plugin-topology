package handler

import (
	"strings"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/service/public"
	"gitee.com/openeuler/PilotGo/sdk/response"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// func SingleHostHandle(ctx *gin.Context) {
// 	uuid := ctx.Param("uuid")
// 	nodes, edges, collect_errlist, process_errlist := service.SingleHostService(uuid)

// 	if len(collect_errlist) != 0 || len(process_errlist) != 0 {
// 		for i, cerr := range collect_errlist {
// 			collect_errlist[i] = errors.Wrap(cerr, "**errstack**4") // err top
// 			agentmanager.Topo.ErrCh <- collect_errlist[i]
// 		}

// 		for i, perr := range process_errlist {
// 			process_errlist[i] = errors.Wrap(perr, "**errstack**10") // err top
// 			agentmanager.Topo.ErrCh <- process_errlist[i]
// 		}
// 	}

// 	if len(nodes) == 0 || len(edges) == 0 {
// 		err := errors.New("nodes list is null or edges list is null **errstack**0") // err top
// 		agentmanager.Topo.ErrCh <- err

// 		ctx.JSON(http.StatusBadRequest, gin.H{
// 			"code":  -1,
// 			"error": err.Error(),
// 			"data":  nil,
// 		})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"code":  0,
// 		"error": nil,
// 		"data": map[string]interface{}{
// 			"nodes": nodes,
// 			"edges": edges,
// 		},
// 	})
// }

func SingleHostTreeHandle(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	nodes, err := public.SingleHostTreeService(uuid)
	if err != nil {
		switch strings.Split(errors.Cause(err).Error(), "**")[1] {
		case "errstack":
			err = errors.Wrap(err, " **errstack**2") // err top
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
			response.Fail(ctx, nil, err.Error())
			return
		case "errstackfatal":
			err = errors.Wrap(err, " **errstackfatal**2") // err top
			response.Fail(ctx, nil, errors.Cause(err).Error())
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
			return
		}

	}

	if nodes == nil {
		err := errors.New("node tree is null **errstack**0") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	// ttcode 导出图数据用于前端调试
	// logfile, _ := os.OpenFile("/root/single.json", os.O_WRONLY|os.O_CREATE, 0666)
	// encoder := json.NewEncoder(logfile)
	// encoder.SetIndent("", " ")
	// encoder.Encode(gin.H{
	// 	"code":  0,
	// 	"error": nil,
	// 	"data": map[string]interface{}{
	// 		"tree": nodes,
	// 	},
	// })
	// signal.Close()
	// os.Exit(1)

	response.Success(ctx, map[string]interface{}{
		"tree": nodes,
	}, "")
}

func MultiHostHandle(ctx *gin.Context) {
	nodes, edges, combos, err := public.MultiHostService()
	if err != nil {
		switch strings.Split(errors.Cause(err).Error(), "**")[1] {
		case "errstack":
			err = errors.Wrap(err, " **errstack**2") // err top
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
			response.Fail(ctx, nil, err.Error())
			return
		case "errstackfatal":
			err = errors.Wrap(err, " **errstack**2") // err top
			response.Fail(ctx, nil, errors.Cause(err).Error())
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
			return
		}

	}

	if len(nodes) == 0 || len(edges) == 0 {
		err := errors.New("nodes list is null or edges list is null **errstack**0") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	// ttcode 导出图数据用于前端调试
	// logfile, _ := os.OpenFile("/root/cluster.json", os.O_WRONLY|os.O_CREATE, 0666)
	// encoder := json.NewEncoder(logfile)
	// encoder.SetIndent("", " ")
	// encoder.Encode(gin.H{
	// 	"code":  0,
	// 	"error": nil,
	// 	"data": map[string]interface{}{
	// 		"nodes":  nodes,
	// 		"edges":  edges,
	// 		"combos": combos,
	// 	},
	// })
	// signal.Close()
	// os.Exit(1)

	response.Success(ctx, map[string]interface{}{
		"nodes":  nodes,
		"edges":  edges,
		"combos": combos,
	}, "")
}
