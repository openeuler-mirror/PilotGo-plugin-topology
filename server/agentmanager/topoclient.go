package agentmanager

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo/sdk/common"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"gitee.com/openeuler/PilotGo/sdk/plugin/client"
	"gitee.com/openeuler/PilotGo/sdk/utils/httputils"
	"github.com/go-redis/redis/v8"
	"github.com/mitchellh/mapstructure"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var Topo *Topoclient

type Topoclient struct {
	Sdkmethod *client.Client

	PAgentMap sync.Map
	TAgentMap sync.Map

	mu   sync.Locker // 暂时闲置
	cond *sync.Cond  // 暂时闲置

	ErrCh chan *meta.Topoerror

	Out io.Writer

	Tctx context.Context
}

func InitPluginClient() {
	var errcondmu sync.Mutex
	// PluginInfo.Url = "http://" + conf.Config().Topo.Server_addr + "/plugin/topology"
	PluginInfo.Url = "http://" + conf.Config().Topo.Server_addr
	PluginClient := client.DefaultClient(PluginInfo)

	// 注册插件扩展点
	var ex []common.Extention
	pe1 := &common.PageExtention{
		Type:       common.ExtentionPage,
		Name:       "配置列表",
		URL:        "/topoList",
		Permission: "plugin.topology.page/menu",
	}
	pe2 := &common.PageExtention{
		Type:       common.ExtentionPage,
		Name:       "创建配置",
		URL:        "/customTopo",
		Permission: "plugin.topology.page/menu",
	}
	ex = append(ex, pe1, pe2)
	PluginClient.RegisterExtention(ex)

	Topo = &Topoclient{
		Sdkmethod: PluginClient,
		mu:        &errcondmu,
		cond:      sync.NewCond(&errcondmu),
		ErrCh:     make(chan *meta.Topoerror, 10),
		Tctx:      context.Background(),
	}
}

func (t *Topoclient) InitMachineList() {
	Wait4TopoServerReady()
	t.Sdkmethod.Wait4Bind()

	// url := "http://" + t.Sdkmethod.Server() + "/api/v1/pluginapi/machine_list"

	// resp, err := httputils.Get(url, &httputils.Params{
	// 	Cookie: map[string]string{
	// 		client.TokenCookie: "",
	// 	}})
	// if err != nil {
	// 	err = errors.Errorf("err-> %s (url-> %s) **fatal**2", err.Error(), url) // err top
	// 	ErrorTransmit(t.Tctx, err, t.ErrCh, true)
	// }

	// statuscode := resp.StatusCode
	// if statuscode != 200 {
	// 	msg := ""
	// 	resp_body := struct {
	// 		Code int    `json:"code"`
	// 		Msg  string `json:"msg"`
	// 	}{}
	// 	if len(resp.Body) != 0 {
	// 		json.Unmarshal(resp.Body, &resp_body)
	// 		msg = resp_body.Msg
	// 	}
	// 	err := errors.Errorf("http返回状态码异常: %d, %s, %s **fatal**2", statuscode, msg, url) // err top
	// 	ErrorTransmit(t.Tctx, err, t.ErrCh, true)
	// }

	// result := &struct {
	// 	Code int         `json:"code"`
	// 	Data interface{} `json:"data"`
	// }{}

	// err = json.Unmarshal(resp.Body, result)
	// if err != nil {
	// 	err = errors.Errorf("%s **fatal**2", err.Error()) // err top
	// 	ErrorTransmit(t.Tctx, err, t.ErrCh, true)
	// }

	// for _, m := range result.Data.([]interface{}) {
	// 	p := &Agent_m{}
	// 	mapstructure.Decode(m, p)
	// 	p.TAState = 0
	// 	t.AddAgent_P(p)
	// }

	machine_list, err := t.Sdkmethod.MachineList()
	if err != nil {
		err = errors.Errorf("%s **fatal**2", err.Error()) // err top
		ErrorTransmit(t.Tctx, err, t.ErrCh, true)
	}

	for _, m := range machine_list {
		p := &Agent_m{}
		p.UUID = m.UUID
		p.Departname = m.Department
		p.IP = m.IP
		p.TAState = 0
		t.AddAgent_P(p)
	}
}

func (t *Topoclient) GetBatchList() ([]*common.BatchList, error) {
	var batch_list []*common.BatchList = make([]*common.BatchList, 0)

	url := "http://" + t.Sdkmethod.Server() + "/api/v1/pluginapi/batch_list"

	resp, err := httputils.Get(url, nil)
	if err != nil {
		return nil, errors.Errorf("err-> %s (url-> %s) **errstackfatal**2", err.Error(), url)
	}

	statuscode := resp.StatusCode
	if statuscode != 200 {
		return nil, errors.Errorf("http返回状态码异常: %d, %s **errstackfatal**2", statuscode, url)
	}

	result := &struct {
		Code int         `json:"code"`
		Data interface{} `json:"data"`
		Msg  string      `json:"msg"`
	}{}

	err = json.Unmarshal(resp.Body, result)
	if err != nil {
		return nil, errors.Errorf("%s **errstackfatal**2", err.Error())
	}

	for _, m := range result.Data.([]interface{}) {
		p := &common.BatchList{}
		mapstructure.Decode(m, p)
		batch_list = append(batch_list, p)
	}

	return batch_list, nil
}

func (t *Topoclient) GetBatchMachineList(batchid uint) ([]string, error) {
	var machine_list []string = make([]string, 0)

	url := "http://" + t.Sdkmethod.Server() + "/api/v1/pluginapi/batch_uuid?batchId=" + strconv.Itoa(int(batchid))

	resp, err := httputils.Get(url, nil)
	if err != nil {
		return nil, errors.Errorf("err-> %s (url-> %s) **errstackfatal**2", err.Error(), url)
	}

	statuscode := resp.StatusCode
	if statuscode != 200 {
		return nil, errors.Errorf("http返回状态码异常: %d, %s **errstackfatal**2", statuscode, url)
	}

	result := &struct {
		Code int         `json:"code"`
		Data interface{} `json:"data"`
		Msg  string      `json:"msg"`
	}{}

	err = json.Unmarshal(resp.Body, result)
	if err != nil {
		return nil, errors.Errorf("%s **errstackfatal**2", err.Error())
	}

	for _, m := range result.Data.([]interface{}) {
		machine_list = append(machine_list, m.(string))
	}

	return machine_list, nil
}

func (t *Topoclient) UpdateMachineList() {
	// var ip_port_pilotgo_server string

	// if Topo != nil && Topo.Sdkmethod != nil {
	// 	ip_port_pilotgo_server = Topo.Sdkmethod.Server()
	// }

	// url := "http://" + ip_port_pilotgo_server + "/api/v1/pluginapi/machine_list"

	// resp, err := httputils.Get(url, nil)
	// if err != nil {
	// 	err = errors.Errorf("%s **fatal**2", err.Error()) // err top
	// 	ErrorTransmit(t.Tctx, err, t.ErrCh, true)
	// }

	// statuscode := resp.StatusCode
	// if statuscode != 200 {
	// 	err = errors.Errorf("http返回状态码异常: %d **fatal**2", statuscode) // err top
	// 	ErrorTransmit(t.Tctx, err, t.ErrCh, true)
	// }

	// result := &struct {
	// 	Code int         `json:"code"`
	// 	Data interface{} `json:"data"`
	// }{}

	// err = json.Unmarshal(resp.Body, result)
	// if err != nil {
	// 	err = errors.Errorf("%s **fatal**2", err.Error()) // err top
	// 	ErrorTransmit(t.Tctx, err, t.ErrCh, true)
	// }

	// if Topo != nil {
	// 	Topo.PAgentMap.Range(func(key, value interface{}) bool {
	// 		t.DeleteAgent_P(key.(string))
	// 		return true
	// 	})
	// } else {
	// 	err := errors.New("agentmanager.Topo is nil, can not clear Topo.PAgentMap **fatal**6") // err top
	// 	ErrorTransmit(t.Tctx, err, t.ErrCh, true)
	// }

	// for _, m := range result.Data.([]interface{}) {
	// 	p := &Agent_m{}
	// 	mapstructure.Decode(m, p)
	// 	p.TAState = 0
	// 	t.AddAgent_P(p)
	// }

	machine_list, err := t.Sdkmethod.MachineList()
	if err != nil {
		err = errors.Errorf("%s **fatal**2", err.Error()) // err top
		ErrorTransmit(t.Tctx, err, t.ErrCh, true)
	}

	if Topo != nil {
		Topo.PAgentMap.Range(func(key, value interface{}) bool {
			t.DeleteAgent_P(key.(string))
			return true
		})
	} else {
		err := errors.New("agentmanager.Topo is nil, can not clear Topo.PAgentMap **fatal**6") // err top
		ErrorTransmit(t.Tctx, err, t.ErrCh, true)
	}

	for _, m := range machine_list {
		p := &Agent_m{}
		p.UUID = m.UUID
		p.Departname = m.Department
		p.IP = m.IP
		p.TAState = 0
		t.AddAgent_P(p)
	}
}

func (t *Topoclient) InitLogger() {
	err := logger.Init(conf.Config().Logopts)
	if err != nil {
		err = errors.Errorf("%s **errstackfatal**2", err.Error()) // err top
		ErrorTransmit(t.Tctx, err, t.ErrCh, true)
	}
}

func (t *Topoclient) InitErrorControl(errch <-chan *meta.Topoerror) {
	switch conf.Config().Logopts.Driver {
	case "stdout":
		t.Out = os.Stdout
	case "file":
		logfile, err := os.OpenFile(conf.Global_config.Logopts.Path, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
		t.Out = logfile
	}

	go func(ch <-chan *meta.Topoerror) {
		for {
			topoerr, ok := <-ch
			if !ok {
				break
			}

			if topoerr.Err != nil {
				errarr := strings.Split(errors.Cause(topoerr.Err).Error(), "**")
				switch errarr[1] {
				case "debug": // 只打印最底层error的message，不展开错误链的调用栈
					logger.Debug("%+v\n", errors.Cause(topoerr.Err).Error())
				case "warn": // 只打印最底层error的message，不展开错误链的调用栈
					logger.Warn("%+v\n", errors.Cause(topoerr.Err).Error())
				case "errstack": // 打印错误链的调用栈
					fmt.Fprintf(t.Out, "%+v\n", topoerr.Err)
					// errors.EORE(err)
				case "errstackfatal": // 打印错误链的调用栈，并结束程序
					fmt.Fprintf(t.Out, "%+v\n", topoerr.Err)
					// errors.EORE(err)
					topoerr.Cancel()
				default:
					fmt.Printf("only support \"info warn fatal\" error type: %+v\n", topoerr.Err)
					os.Exit(1)
				}
			}
		}
	}(errch)
}

func (t *Topoclient) InitConfig() {
	flag.StringVar(&conf.Config_dir, "conf", "/opt/PilotGo/plugin/topology/server", "topo-server configuration directory")
	flag.Parse()

	bytes, err := os.ReadFile(conf.Config_file())
	if err != nil {
		err = errors.Errorf("open file failed: %s, %s", conf.Config_file(), err.Error()) // err top
		fmt.Printf("%+v\n", err)
		os.Exit(-1)
	}

	err = yaml.Unmarshal(bytes, &conf.Global_config)
	if err != nil {
		err = errors.Errorf("yaml unmarshal failed: %s", err.Error()) // err top
		fmt.Printf("%+v\n", err)
		os.Exit(-1)
	}
}

func (t *Topoclient) SignalMonitoring(neo4jclient neo4j.Driver, redisclient redis.Client) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for {
		s := <-c
		switch s {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			neo4jclient.Close()
			fmt.Println()
			logger.Info("close the connection to neo4j\n")
			redisclient.Close()
			logger.Info("close the connection to redis\n")
			os.Exit(-1)
		default:
			logger.Warn("unknown signal-> %s\n", s.String())
		}
	}
}
