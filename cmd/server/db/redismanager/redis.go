package redismanager

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/global"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"gitee.com/openeuler/PilotGo/sdk/utils/httputils"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

var Global_Redis *RedisClient

type RedisClient struct {
	Addr     string
	Password string
	DB       int
	Client   redis.Client
}

func RedisInit(url, pass string, db int, dialTimeout time.Duration) *RedisClient {
	r := &RedisClient{
		Addr:     url,
		Password: pass,
		DB:       db,
	}

	var cfg *redis.Options
	if conf.Global_Config.Redis.UseTLS {
		cfg = &redis.Options{
			Addr:     r.Addr,
			Password: r.Password,
			DB:       r.DB,
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	} else {
		cfg = &redis.Options{
			Addr:     r.Addr,
			Password: r.Password,
			DB:       r.DB,
		}
	}

	global.Global_redis_client = redis.NewClient(cfg)
	r.Client = *global.Global_redis_client

	// 使用超时上下文，验证redis
	timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), dialTimeout)
	defer cancelFunc()
	_, err := r.Client.Ping(timeoutCtx).Result()
	if err != nil {
		err = errors.Errorf("redis connection timeout: %s", err.Error())
		global.ERManager.ErrorTransmit("error", err, true, true)
	}

	return r
}

func (r *RedisClient) Set(key string, value interface{}) error {
	if key == "" {
		return errors.New("key is empty")
	}

	bytes, _ := json.Marshal(value)
	err := r.Client.Set(global.RootContext, key, string(bytes), 0).Err()
	if err != nil {
		err = errors.Errorf("failed to set key-value: %s", err.Error())
		return err
	}

	return nil
}

func (r *RedisClient) Get(key string, obj interface{}) (interface{}, error) {
	if key == "" {
		return nil, errors.New("key is empty")
	}

	data, err := r.Client.Get(global.RootContext, key).Result()
	if err != nil {
		err = errors.Errorf("failed to get value: %s, %s", key, err.Error())
		return nil, err
	}
	json.Unmarshal([]byte(data), obj)
	return obj, nil
}

func (r *RedisClient) Scan(key string) ([]string, error) {
	keys := []string{}

	if key == "" {
		return nil, errors.New("key is empty")
	}

	iterator := r.Client.Scan(global.RootContext, 0, key, 0).Iterator()
	for iterator.Next(global.RootContext) {
		key := iterator.Val()
		keys = append(keys, key)
	}

	return keys, nil
}

func (r *RedisClient) Delete(key string) error {
	if key == "" {
		return errors.New("key is empty")
	}

	err := r.Client.Del(global.RootContext, key).Err()
	if err != nil {
		err = errors.Errorf("failed to del key-value: %s", err.Error())
		return err
	}
	return nil
}

// 基于batch中的机器列表和PAgentMap更新TAgentMap中运行状态的agent
func (r *RedisClient) UpdateTopoRunningAgentList(uuids []string, updateonce bool) int {
	var running_agent_num int32
	var once sync.Once
	var wg sync.WaitGroup
	var abort_reason []string

	if agentmanager.Global_AgentManager == nil {
		err := errors.New("Global_AgentManager is nil")
		global.ERManager.ErrorTransmit("error", err, true, true)
		return -1
	}

	// 重置TAgentMap
	agentmanager.Global_AgentManager.TAgentMap.Range(func(key, value interface{}) bool {
		agentmanager.Global_AgentManager.TAgentMap.Delete(key)
		return true
	})

	// 阻塞
	for {
		agent_keys, err := r.Scan("heartbeat-topoagent*")
		if err != nil {
			err = errors.Wrap(err, " ")
			global.ERManager.ErrorTransmit("error", err, false, true)
			continue
		}

		if len(agent_keys) != 0 {
			for _, agentkey := range agent_keys {
				wg.Add(1)
				go func(key string) {
					defer wg.Done()
					v, err := r.Get(key, &AgentHeartbeat{})
					if err != nil {
						err = errors.Wrap(err, " ")
						global.ERManager.ErrorTransmit("error", err, false, true)
						return
					}

					agentvalue := v.(*AgentHeartbeat)

					if len(uuids) != 0 {
						inbatch := false
						for _, uuid := range uuids {
							if agentvalue.UUID == uuid {
								inbatch = true
								break
							}
						}

						if !inbatch {
							abort_reason = append(abort_reason, fmt.Sprintf("%s:未在批次范围内", agentvalue.UUID))
							return
						}
					}

					// 当agent满足：①心跳要求；②包含于pilotgo纳管的机器范围内；③包含于uuids批次机器内；④agent的IP:PORT可连通等条件时，则加入TAgentMap
					if elapse := time.Since(agentvalue.Time); elapse >= 1*time.Second+time.Duration(agentvalue.HeartbeatInterval)*time.Second {
						abort_reason = append(abort_reason, fmt.Sprintf("%s:心跳超时", agentvalue.UUID))
						return
					}

					agentp := agentmanager.Global_AgentManager.GetAgent_P(agentvalue.UUID)
					if agentp == nil {
						abort_reason = append(abort_reason, fmt.Sprintf("%s:未被pilotgo纳管", agentvalue.UUID))
						return
					}

					if ok, err := global.IsIPandPORTValid(agentp.IP, agentmanager.Global_AgentManager.AgentPort); !ok {
						err := errors.Errorf("%s:%s is unreachable (%s) %s", agentp.IP, agentmanager.Global_AgentManager.AgentPort, err.Error(), agentp.UUID)
						global.ERManager.ErrorTransmit("warn", err, false, false)
						abort_reason = append(abort_reason, fmt.Sprintf("%s:ip||port不可达", agentvalue.UUID))
						return
					}
					agentmanager.Global_AgentManager.AddAgent_T(agentp)

					atomic.AddInt32(&running_agent_num, int32(1))
				}(agentkey)

			}

			wg.Wait()

			if running_agent_num > 0 {
				break
			}
		}

		once.Do(func() {
			logger.Warn("no running agent......")
		})

		if len(abort_reason) != 0 {
			logger.Warn("========agent status========")
			for _, r := range abort_reason {
				logger.Warn(r)
			}
			logger.Warn("============================")
		}

		if updateonce {
			return 0
		}

		time.Sleep(1 * time.Second)
	}

	logger.Info("running agent number: %d", running_agent_num)
	return int(running_agent_num)
}

// server端对agent端的主动健康监测，更新agent心跳信息
func (r *RedisClient) ActiveHeartbeatDetection(uuids []string) {
	var wg sync.WaitGroup

	activeHeartbeatDetection := func(agent *agentmanager.Agent) {
		url := "http://" + agent.IP + ":" + conf.Global_Config.Topo.Agent_port + "/plugin/topology/api/health"
		if resp, err := httputils.Get(url, nil); err == nil && resp != nil && resp.StatusCode == 200 {
			type agentinfo struct {
				Interval int `json:"interval"`
			}

			resp_body := struct {
				Code int       `json:"code"`
				Data agentinfo `json:"data"`
				Msg  string    `json:"msg"`
			}{}

			err = json.Unmarshal(resp.Body, &resp_body)
			if err != nil {
				err = errors.Errorf(err.Error())
				global.ERManager.ErrorTransmit("error", err, false, true)
				return
			}

			key := "heartbeat-topoagent-" + agent.UUID
			value := AgentHeartbeat{
				UUID:              agent.UUID,
				Addr:              agent.IP + ":" + conf.Global_Config.Topo.Agent_port,
				HeartbeatInterval: resp_body.Data.Interval,
				Time:              time.Now(),
			}

			err = r.Set(key, value)
			if err != nil {
				err = errors.Wrap(err, " ")
				global.ERManager.ErrorTransmit("error", err, false, true)
				return
			}
		}
	}

	if agentmanager.Global_AgentManager == nil {
		err := errors.New("Global_AgentManager is nil")
		global.ERManager.ErrorTransmit("error", err, true, true)
		return
	}

	if len(uuids) == 0 {
		agentmanager.Global_AgentManager.PAgentMap.Range(func(key, value interface{}) bool {
			agent := value.(*agentmanager.Agent)
			wg.Add(1)
			go func(a *agentmanager.Agent) {
				defer wg.Done()

				if ok, _ := global.IsIPandPORTValid(a.IP, agentmanager.Global_AgentManager.AgentPort); !ok {
					// err := errors.Errorf("%s:%s is unreachable (%s) %s", a.IP, agentmanager.Topo.AgentPort, err.Error(), a.UUID)
					// resourcemanage.ErrorTransmit("warn", err, false, false)
					return
				}

				activeHeartbeatDetection(agent)
			}(agent)
			return true
		})

		wg.Wait()

		return
	}

	for _, uuid := range uuids {
		wg.Add(1)
		go func(_uuid string) {
			defer wg.Done()

			agent := agentmanager.Global_AgentManager.GetAgent_P(_uuid)
			if agent == nil {
				return
			}

			if ok, _ := global.IsIPandPORTValid(agent.IP, agentmanager.Global_AgentManager.AgentPort); !ok {
				// err := errors.Errorf("%s:%s is unreachable (%s) %s", agent.IP, agentmanager.Topo.AgentPort, err.Error(), agent.UUID)
				// resourcemanage.ErrorTransmit("warn", err, false, false)
				return
			}

			activeHeartbeatDetection(agent)
		}(uuid)
	}

	wg.Wait()
}
