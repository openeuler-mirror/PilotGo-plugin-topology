package handler

import (
	"fmt"
	"net/http"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/dao"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type AgentHeartbeat struct {
	Addr string
	Time time.Time
}

func HeartbeatHandle(ctx *gin.Context) {
	// agent发送的心跳参数为uuid和ip:port，写入redis的数据为 (heartbeat-uuid: {addr: "10.44.55.66:9992", time: "2023-12-22T17:09:23+08:00"})
	key := "heartbeat-topoagent-" + ctx.Query("uuid")
	value := AgentHeartbeat{
		Addr: ctx.Query("agentaddr"),
		Time: time.Now(),
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
