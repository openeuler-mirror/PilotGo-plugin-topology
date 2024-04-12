package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db/graphmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db/redismanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/pluginclient"
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
	value := redismanager.AgentHeartbeat{
		UUID:              uuid,
		Addr:              addr,
		HeartbeatInterval: heartbeatinterval,
		Time:              time.Now(),
	}

	if agentmanager.Global_AgentManager == nil {
		err := errors.New("Global_AgentManager is nil **errstackfatal**0") // err top
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
		return
	}

	if agentmanager.Global_AgentManager.GetAgent_P(uuid) == nil {
		err := errors.Errorf("unknown agent's heartbeat: %s, %s **warn**1", uuid, addr) // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		return
	}

	if redismanager.Global_Redis == nil {
		err := errors.New("Global_Redis is nil **errstackfatal**0") // err top
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":  -1,
			"error": err.Error(),
			"data":  nil,
		})
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
		return
	}

	err := redismanager.Global_Redis.Set(key, value)
	if err != nil {
		err = errors.Wrap(err, " **errstack**2") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)

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
	if graphmanager.Global_GraphDB == nil {
		err := errors.New("Global_GraphDB is nil **errstackfatal**0") // err top
		response.Fail(ctx, nil, err.Error())
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
		return
	}

	times, err := graphmanager.Global_GraphDB.Timestamps_query()
	if err != nil {
		err = errors.Wrap(err, " **errstack**2")
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)

		response.Fail(ctx, nil, err.Error())
		return
	}

	response.Success(ctx, times, "")
}

func AgentListHandle(ctx *gin.Context) {
	agentmap := make(map[string]string)

	if agentmanager.Global_AgentManager == nil {
		err := errors.New("Global_AgentManager is nil **errstackfatal**0") // err top
		response.Fail(ctx, nil, err.Error())
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
		return
	}

	agentmanager.Global_AgentManager.TAgentMap.Range(func(key, value interface{}) bool {
		agent := value.(*agentmanager.Agent)
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
	if pluginclient.Global_Client == nil {
		err := errors.New("Global_Client is nil **errstackfatal**2") // err top
		response.Fail(ctx, nil, err.Error())
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
		return
	}

	batchlist, err := pluginclient.Global_Client.BatchList()
	if err != nil {
		err = errors.Errorf("%+v **errstack**2", err.Error()) // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		response.Fail(ctx, nil, err.Error())
		return
	}

	response.Success(ctx, batchlist, "successfully get batch list")
}

func BatchMachineListHandle(ctx *gin.Context) {
	var machines []map[string]string = make([]map[string]string, 0)

	BatchId := ctx.Query("batchId")
	if BatchId == "" {
		response.Fail(ctx, nil, "batchId is empty")
		return
	}

	if pluginclient.Global_Client == nil {
		err := errors.New("Global_Client is nil **errstackfatal**2") // err top
		response.Fail(ctx, nil, err.Error())
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
		return
	}

	machine_uuids, err := pluginclient.Global_Client.BatchUUIDList(BatchId)
	if err != nil {
		err = errors.Errorf("%+v **errstack**2", err.Error()) // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		response.Fail(ctx, nil, err.Error())
		return
	}

	if agentmanager.Global_AgentManager == nil {
		err := errors.New("Global_AgentManager is nil **errstackfatal**0") // err top
		response.Fail(ctx, nil, err.Error())
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
		return
	}

	agentmanager.Global_AgentManager.PAgentMap.Range(func(key, value interface{}) bool {
		uuid := key.(string)
		agent := value.(*agentmanager.Agent)
		for _, _uuid := range machine_uuids {
			if uuid == _uuid {
				machines = append(machines, map[string]string{
					"uuid": uuid,
					"ip":   agent.IP,
				})
			}
		}

		return true
	})

	response.Success(ctx, machines, "successfully get batch list")
}
