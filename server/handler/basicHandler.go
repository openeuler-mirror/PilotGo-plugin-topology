package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"gitee.com/openeuler/PilotGo/sdk/response"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func HeartbeatHandle(ctx *gin.Context) {
	// agent发送的心跳参数为uuid、ip:port、HeartbeatInterval、time，
	// 写入redis的数据为 (heartbeat-<uuid>: {"UUID": "f7504bef-76e9-446c-95ee-196878b398a1", "Addr": "10.44.55.66:9992", "HeartbeatInterval": 60, "Time": "2023-12-22T17:09:23+08:00"})
	uuid := ctx.Query("uuid")
	addr := ctx.Query("agentaddr")
	heartbeatinterval, _ := strconv.Atoi(ctx.Query("interval"))

	key := "heartbeat-topoagent-" + uuid
	value := meta.AgentHeartbeat{
		UUID:              uuid,
		Addr:              addr,
		HeartbeatInterval: heartbeatinterval,
		Time:              time.Now(),
	}

	if agentmanager.Topo.GetAgent_P(uuid) == nil {
		// err := errors.Errorf("unknown agent's heartbeat: %s, %s **warn**1", uuid, addr) // err top
		// agentmanager.Topo.ErrCh <- err

		logger.Warn("unknown agent's heartbeat: %s, %s", uuid, addr)

		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":  -1,
			"error": fmt.Sprintf("%+v", fmt.Errorf("unknown agent's heartbeat: %s, %s", uuid, addr)),
			"data":  nil,
		})
		return
	}

	err := dao.Global_redis.Set(key, value)
	if err != nil {
		err = errors.Wrap(err, " **warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":  -1,
			"error": fmt.Sprintf("%+v", err),
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

func TimestampsHandle(ctx *gin.Context) {
	times, err := dao.Global_GraphDB.Timestamps_query()
	if err != nil {
		err = errors.Wrap(err, " **warn**2")
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	response.Success(ctx, times, "")
}

func AgentListHandle(ctx *gin.Context) {
	agentmap := make(map[string]string)

	agentmanager.Topo.TAgentMap.Range(func(key, value interface{}) bool {
		agent := value.(*agentmanager.Agent_m)
		if agent.Host_2 != nil {
			agentmap[agent.UUID] = agent.IP + ":" + agent.Port
		}

		return true
	})

	response.Success(ctx, map[string]interface{}{
		"agentlist": agentmap,
	}, "")
}

func BatchListHandle(ctx *gin.Context) {
	batchlist, err := agentmanager.Topo.GetBatchList()
	if err != nil {
		err = errors.Wrap(err, "**warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	response.Success(ctx, batchlist, "successfully get batch list")
}

func BatchMachineListHandle(ctx *gin.Context) {
	BatchId := ctx.Query("batchId")

	machine_uuids, err := agentmanager.Topo.GetBatchMachineList(BatchId)
	if err != nil {
		err = errors.Wrap(err, "**warn**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	response.Success(ctx, machine_uuids, "successfully get batch list")
}