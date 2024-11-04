package handler

import (
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/resourcemanage"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/service/public"
	"gitee.com/openeuler/PilotGo/sdk/response"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// func SingleHostHandle(ctx *gin.Context) {
// 	uuid := ctx.Param("uuid")
// 	nodes, edges, collect_errlist, process_errlist := service.SingleHostService(uuid)

// 	if len(collect_errlist) != 0 || len(process_errlist) != 0 {
// 		for i, cerr := range collect_errlist {
// 			collect_errlist[i] = errors.Wrap(cerr, " ")
//          resourcemanage.ERManager.ErrorTransmit("error", collect_errlist[i], false, true)
// 		}

// 		for i, perr := range process_errlist {
// 			process_errlist[i] = errors.Wrap(perr, " ")
//          resourcemanage.ERManager.ErrorTransmit("error", process_errlist[i], false, true)
// 		}
// 	}

// 	if len(nodes) == 0 || len(edges) == 0 {
// 		err := errors.New("nodes list is null or edges list is null")
//      resourcemanage.ERManager.ErrorTransmit("error", err, false, true)

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
	nodes, err, exit := public.SingleHostTreeService(uuid)
	if err != nil {
		if exit {
			err = errors.Wrap(err, " ")
			response.Fail(ctx, nil, errors.Cause(err).Error())
			resourcemanage.ERManager.ErrorTransmit("error", err, true, true)
			return
		}
		err = errors.Wrap(err, " ")
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
		response.Fail(ctx, nil, err.Error())
		return
	}

	if nodes == nil {
		err := errors.New("node tree is null")
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)

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
	nodes, edges, combos, err, exit := public.MultiHostService()
	if err != nil {
		if exit {
			err = errors.Wrap(err, " ")
			response.Fail(ctx, nil, errors.Cause(err).Error())
			resourcemanage.ERManager.ErrorTransmit("error", err, true, true)
			return
		}
		err = errors.Wrap(err, " ")
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)
		response.Fail(ctx, nil, err.Error())
		return
	}

	if len(nodes) == 0 || len(edges) == 0 {
		err := errors.New("nodes list is null or edges list is null")
		resourcemanage.ERManager.ErrorTransmit("error", err, false, true)

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
