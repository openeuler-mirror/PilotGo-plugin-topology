package dao

import (
	"context"
	"encoding/json"
	"fmt"
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
	if agentmanager.Topo != nil && Global_redis != nil {
		bytes, _ := json.Marshal(value)
		err := Global_redis.Client.Set(agentmanager.Topo.Tctx, key, string(bytes), 0).Err()
		if err != nil {
			err = errors.Errorf("failed to set key-value: %s **2", err.Error())
			return err
		}

		return nil
	}

	err := errors.New("global_redis is nil **warn**11")
	return err
}

func (r *RedisClient) Get(key string, obj interface{}) (interface{}, error) {
	if agentmanager.Topo != nil && Global_redis != nil {
		data, err := Global_redis.Client.Get(agentmanager.Topo.Tctx, key).Result()
		if err != nil {
			err = errors.Errorf("failed to get value: %s, %s **2", key, err.Error())
			return nil, err
		}
		json.Unmarshal([]byte(data), obj)
		return obj, nil
	}

	err := errors.New("global_redis is nil **warn**11")
	return nil, err
}

func (r *RedisClient) Scan(key string) ([]string, error) {
	keys := []string{}

	if agentmanager.Topo != nil && Global_redis != nil {
		iterator := Global_redis.Client.Scan(agentmanager.Topo.Tctx, 0, key, 0).Iterator()
		for iterator.Next(agentmanager.Topo.Tctx) {
			key := iterator.Val()
			keys = append(keys, key)
		}

		return keys, nil
	}

	err := errors.New("global_redis is nil **warn**11")
	return nil, err
}

func (r *RedisClient) Delete(key string) error {
	if agentmanager.Topo != nil && Global_redis != nil {
		err := Global_redis.Client.Del(agentmanager.Topo.Tctx, key).Err()
		if err != nil {
			err = errors.Errorf("failed to del key-value: %s **2", err.Error())
			return err
		}
		return nil
	}

	err := errors.New("global_redis is nil **warn**11")
	return err
}

// 更新运行状态agent的列表
func (r *RedisClient) UpdateTopoRunningAgentList() int {
	var runningAgentNum int
	var once sync.Once

	if agentmanager.Topo == nil {
		err := errors.New("agentmanager.Topo is not initialized!") // err top
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	agentmanager.Topo.TAgentMap.Range(func(key, value interface{}) bool {
		agentmanager.Topo.TAgentMap.Delete(key)
		return true
	})

	for {
		agent_keys, err := r.Scan("heartbeat-topoagent*")
		if err != nil {
			err = errors.Wrap(err, "**warn**2") // err top
			agentmanager.Topo.ErrCh <- err
			continue
		}

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
					if agentp := agentmanager.Topo.GetAgent_P(agentvalue.UUID); agentp != nil {
						agentmanager.Topo.AddAgent_T(agentp)
						runningAgentNum += 1
					}
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
