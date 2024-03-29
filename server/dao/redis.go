package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/utils"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"gitee.com/openeuler/PilotGo/sdk/utils/httputils"
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
		err = errors.Errorf("redis connection timeout: %s **errstackfatal**2", err.Error()) // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, true)
	}

	return r
}

func (r *RedisClient) Set(key string, value interface{}) error {
	if agentmanager.Topo != nil && Global_redis != nil {
		bytes, _ := json.Marshal(value)
		err := Global_redis.Client.Set(agentmanager.Topo.Tctx, key, string(bytes), 0).Err()
		if err != nil {
			err = errors.Errorf("failed to set key-value: %s **errstack**2", err.Error())
			return err
		}

		return nil
	}

	err := errors.New("global_redis is nil **errstack**11")
	return err
}

func (r *RedisClient) Get(key string, obj interface{}) (interface{}, error) {
	if agentmanager.Topo != nil && Global_redis != nil {
		data, err := Global_redis.Client.Get(agentmanager.Topo.Tctx, key).Result()
		if err != nil {
			err = errors.Errorf("failed to get value: %s, %s **errstack**2", key, err.Error())
			return nil, err
		}
		json.Unmarshal([]byte(data), obj)
		return obj, nil
	}

	return nil, errors.New("global_redis is nil **errstack**11")
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

	err := errors.New("global_redis is nil **errstack**11")
	return nil, err
}

func (r *RedisClient) Delete(key string) error {
	if agentmanager.Topo != nil && Global_redis != nil {
		err := Global_redis.Client.Del(agentmanager.Topo.Tctx, key).Err()
		if err != nil {
			err = errors.Errorf("failed to del key-value: %s **errstack**2", err.Error())
			return err
		}
		return nil
	}

	err := errors.New("global_redis is nil **errstack**11")
	return err
}

// 基于batch中的机器列表和PAgentMap更新TAgentMap中运行状态的agent
func (r *RedisClient) UpdateTopoRunningAgentList(uuids []string, updateonce bool) int {
	var running_agent_num int32
	var once sync.Once
	var wg sync.WaitGroup
	var abort_reason []string

	if agentmanager.Topo == nil {
		err := errors.New("agentmanager.Topo is not initialized!") // err top
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	if Global_redis == nil {
		return -1
	}

	// 重置TAgentMap
	agentmanager.Topo.TAgentMap.Range(func(key, value interface{}) bool {
		agentmanager.Topo.TAgentMap.Delete(key)
		return true
	})

	// 阻塞
	for {
		agent_keys, err := r.Scan("heartbeat-topoagent*")
		if err != nil {
			err = errors.Wrap(err, "**errstack**2") // err top
			agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
			continue
		}

		if len(agent_keys) != 0 {
			for _, agentkey := range agent_keys {
				wg.Add(1)
				go func(key string) {
					defer wg.Done()
					v, err := r.Get(key, &meta.AgentHeartbeat{})
					if err != nil {
						err = errors.Wrap(err, "**errstack**2") // err top
						agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
						return
					}

					agentvalue := v.(*meta.AgentHeartbeat)

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

					agentp := agentmanager.Topo.GetAgent_P(agentvalue.UUID)
					if agentp == nil {
						abort_reason = append(abort_reason, fmt.Sprintf("%s:未被pilotgo纳管", agentvalue.UUID))
						return
					}

					if ok, err := utils.IsIPandPORTValid(agentp.IP, agentmanager.Topo.AgentPort); !ok {
						err := errors.Errorf("%s:%s is unreachable (%s) %s **warn**1", agentp.IP, agentmanager.Topo.AgentPort, err.Error(), agentp.UUID) // err top
						agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
						abort_reason = append(abort_reason, fmt.Sprintf("%s:ip||port不可达", agentvalue.UUID))
						return
					}
					agentmanager.Topo.AddAgent_T(agentp)

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

		// ttcode
		if len(abort_reason) != 0 {
			logger.Debug(">>>>>>>>>>>>获取agent状态信息")
			for _, r := range abort_reason {
				logger.Debug(r)
			}
			logger.Debug(">>>>>>>>>>>>")
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

	if Global_redis == nil {
		err := errors.New("redis client not init **errstack**1")
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
		return
	}

	activeHeartbeatDetection := func(agent *agentmanager.Agent_m) {
		url := "http://" + agent.IP + ":" + conf.Config().Topo.Agent_port + "/plugin/topology/api/health"
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
				err = errors.Errorf("%+v **errstack**2", err.Error()) // err top
				agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
				return
			}

			key := "heartbeat-topoagent-" + agent.UUID
			value := meta.AgentHeartbeat{
				UUID:              agent.UUID,
				Addr:              agent.IP + ":" + conf.Config().Topo.Agent_port,
				HeartbeatInterval: resp_body.Data.Interval,
				Time:              time.Now(),
			}

			err = Global_redis.Set(key, value)
			if err != nil {
				err = errors.Wrap(err, " **errstack**2") // err top
				agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
				return
			}
		}
	}

	if len(uuids) == 0 {
		agentmanager.Topo.PAgentMap.Range(func(key, value interface{}) bool {
			agent := value.(*agentmanager.Agent_m)
			wg.Add(1)
			go func(a *agentmanager.Agent_m) {
				defer wg.Done()

				if ok, _ := utils.IsIPandPORTValid(a.IP, agentmanager.Topo.AgentPort); !ok {
					// err := errors.Errorf("%s:%s is unreachable (%s) %s **warn**1", a.IP, agentmanager.Topo.AgentPort, err.Error(), a.UUID) // err top
					// agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
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

			agent := agentmanager.Topo.GetAgent_P(_uuid)
			if agent == nil {
				return
			}

			if ok, _ := utils.IsIPandPORTValid(agent.IP, agentmanager.Topo.AgentPort); !ok {
				// err := errors.Errorf("%s:%s is unreachable (%s) %s **warn**1", agent.IP, agentmanager.Topo.AgentPort, err.Error(), agent.UUID) // err top
				// agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
				return
			}

			activeHeartbeatDetection(agent)
		}(uuid)
	}

	wg.Wait()
}
