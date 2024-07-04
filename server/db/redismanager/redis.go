package redismanager

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/global"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"
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
	r.Client = *redis.NewClient(cfg)

	// 使用超时上下文，验证redis
	timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), dialTimeout)
	defer cancelFunc()
	_, err := r.Client.Ping(timeoutCtx).Result()
	if err != nil {
		err = errors.Errorf("redis connection timeout: %s **errstackfatal**2", err.Error()) // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
	}

	return r
}

func (r *RedisClient) Set(key string, value interface{}) error {
	if key == "" {
		return errors.New("key is empty **errstack**1")
	}

	bytes, _ := json.Marshal(value)
	err := r.Client.Set(pluginclient.Global_Context, key, string(bytes), 0).Err()
	if err != nil {
		err = errors.Errorf("failed to set key-value: %s **errstack**2", err.Error())
		return err
	}

	return nil
}

func (r *RedisClient) Get(key string, obj interface{}) (interface{}, error) {
	if key == "" {
		return nil, errors.New("key is empty **errstack**1")
	}

	data, err := r.Client.Get(pluginclient.Global_Context, key).Result()
	if err != nil {
		err = errors.Errorf("failed to get value: %s, %s **errstack**2", key, err.Error())
		return nil, err
	}
	json.Unmarshal([]byte(data), obj)
	return obj, nil
}

func (r *RedisClient) Scan(key string) ([]string, error) {
	keys := []string{}

	if key == "" {
		return nil, errors.New("key is empty **errstack**1")
	}

	iterator := r.Client.Scan(pluginclient.Global_Context, 0, key, 0).Iterator()
	for iterator.Next(pluginclient.Global_Context) {
		key := iterator.Val()
		keys = append(keys, key)
	}

	return keys, nil
}

func (r *RedisClient) Delete(key string) error {
	if key == "" {
		return errors.New("key is empty **errstack**1")
	}

	err := r.Client.Del(pluginclient.Global_Context, key).Err()
	if err != nil {
		err = errors.Errorf("failed to del key-value: %s **errstack**2", err.Error())
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
		err := errors.New("Global_AgentManager is nil **errstackfatal**0") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
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
			err = errors.Wrap(err, "**errstack**2") // err top
			errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
			continue
		}

		if len(agent_keys) != 0 {
			for _, agentkey := range agent_keys {
				wg.Add(1)
				go func(key string) {
					defer wg.Done()
					v, err := r.Get(key, &AgentHeartbeat{})
					if err != nil {
						err = errors.Wrap(err, "**errstack**2") // err top
						errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
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
						err := errors.Errorf("%s:%s is unreachable (%s) %s **warn**1", agentp.IP, agentmanager.Global_AgentManager.AgentPort, err.Error(), agentp.UUID) // err top
						errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
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
				err = errors.Errorf("%+v **errstack**2", err.Error()) // err top
				errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
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
				err = errors.Wrap(err, " **errstack**2") // err top
				errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
				return
			}
		}
	}

	if agentmanager.Global_AgentManager == nil {
		err := errors.New("Global_AgentManager is nil **errstackfatal**0") // err top
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, true)
		return
	}

	if len(uuids) == 0 {
		agentmanager.Global_AgentManager.PAgentMap.Range(func(key, value interface{}) bool {
			agent := value.(*agentmanager.Agent)
			wg.Add(1)
			go func(a *agentmanager.Agent) {
				defer wg.Done()

				if ok, _ := global.IsIPandPORTValid(a.IP, agentmanager.Global_AgentManager.AgentPort); !ok {
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

			agent := agentmanager.Global_AgentManager.GetAgent_P(_uuid)
			if agent == nil {
				return
			}

			if ok, _ := global.IsIPandPORTValid(agent.IP, agentmanager.Global_AgentManager.AgentPort); !ok {
				// err := errors.Errorf("%s:%s is unreachable (%s) %s **warn**1", agent.IP, agentmanager.Topo.AgentPort, err.Error(), agent.UUID) // err top
				// agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, false)
				return
			}

			activeHeartbeatDetection(agent)
		}(uuid)
	}

	wg.Wait()
}
