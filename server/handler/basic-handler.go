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
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func HeartbeatHandle(ctx *gin.Context) {
	// agent发送的心跳参数为uuid和ip:port，写入redis的数据为 (heartbeat-uuid: {addr: "10.44.55.66:9992", time: "2023-12-22T17:09:23+08:00"})
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

		ctx.JSON(http.StatusOK, gin.H{
			"code":  -1,
			"error": fmt.Sprintf("%+v", fmt.Errorf("unknown agent's heartbeat: %s, %s", uuid, addr)),
			"data":  nil,
		})
		return
	}

	err := dao.Global_redis.Set(key, value)
	if err != nil {
		err = errors.Wrap(err, " **warn**2") // err top
		agentmanager.Topo.ErrCh <- err

		ctx.JSON(http.StatusOK, gin.H{
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
