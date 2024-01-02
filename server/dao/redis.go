package dao

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type RedisClient struct {
	Addr     string
	Password string
	DB       int
	Client   redis.Client
}

var Global_redis *RedisClient

func RedisInit(url, pass string, db int, dialTimeout time.Duration) *RedisClient {
	r := &RedisClient{
		Addr:     url,
		Password: pass,
		DB:       db,
	}

	r.Client = *redis.NewClient(&redis.Options{
		Addr:     r.Addr,
		Password: r.Password,
		DB:       r.DB,
	})

	// 使用超时上下文，验证redis
	timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), dialTimeout)
	defer cancelFunc()
	_, err := r.Client.Ping(timeoutCtx).Result()
	if err != nil {
		err = errors.Errorf("redis connection timeout: %s **fatal**2", err.Error()) // err top
		agentmanager.Topo.ErrCh <- err
		agentmanager.Topo.Errmu.Lock()
		agentmanager.Topo.ErrCond.Wait()
		agentmanager.Topo.Errmu.Unlock()
		close(agentmanager.Topo.ErrCh)
		os.Exit(1)
	}

	return r
}

func (r *RedisClient) Set(key string, value interface{}) error {
	var ctx = context.Background()

	bytes, _ := json.Marshal(value)
	err := r.Client.Set(ctx, key, string(bytes), 0).Err()
	if err != nil {
		err = errors.Errorf("failed to set key-value: %s **2", err.Error())
		return err
	}

	return nil
}

func (r *RedisClient) Get(key string, obj interface{}) (interface{}, error) {
	var ctx = context.Background()

	data, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		err = errors.Errorf("failed to get value: %s, %s **2", key, err.Error())
		return nil, err
	}
	json.Unmarshal([]byte(data), obj)
	return obj, nil
}
func (r *RedisClient) Scan(key string) []string {
	var ctx = context.Background()
	keys := []string{}

	iterator := r.Client.Scan(ctx, 0, key, 0).Iterator()
	for iterator.Next(ctx) {
		key := iterator.Val()
		keys = append(keys, key)
	}

	return keys
}

func (r *RedisClient) Delete(key string) error {
	var ctx = context.Background()

	err := r.Client.Del(ctx, key).Err()
	if err != nil {
		err = errors.Errorf("failed to del key-value: %s **2", err.Error())
		return err
	}
	return nil
}

// 更新运行状态agent的列表
func (r *RedisClient) UpdateTopoRunningAgentList() int {
	var runningAgentNum int
	var once sync.Once

	agentmanager.Topo.TAgentMap.Range(func(key, value interface{}) bool {
		agentmanager.Topo.TAgentMap.Delete(key)
		return true
	})

	for {
		agent_keys := r.Scan("heartbeat-topoagent*")
		if len(agent_keys) != 0 {
			for _, agentkey := range agent_keys {
				v, err := r.Get(agentkey, &meta.AgentHeartbeat{})
				if err != nil {
					err = errors.Wrap(err, "**warn**2") // err top
					agentmanager.Topo.ErrCh <- err
					continue
				}

				agentvalue := v.(*meta.AgentHeartbeat)

				if time.Since(agentvalue.Time) < 1*time.Second+time.Duration(agentvalue.HeartbeatInterval)*time.Second {
					agentmanager.Topo.AddAgent_T(agentmanager.Topo.GetAgent_P(agentvalue.UUID))
					runningAgentNum += 1
				}
			}

			if runningAgentNum > 0 {
				break
			}
		}

		once.Do(func() {
			logger.Warn("no running agent......")
		})

		time.Sleep(1 * time.Second)
	}

	logger.Info("running agent number: %d", runningAgentNum)

	return runningAgentNum
}
