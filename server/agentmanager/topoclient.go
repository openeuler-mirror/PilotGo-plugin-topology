package agentmanager

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/pluginclient"
	"gitee.com/openeuler/PilotGo/sdk/common"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"gitee.com/openeuler/PilotGo/sdk/plugin/client"
	"github.com/go-redis/redis/v8"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var Topo *Topoclient

type Topoclient struct {
	PAgentMap sync.Map
	TAgentMap sync.Map

	ErrCh chan *meta.Topoerror

	Out io.Writer

	AgentPort string
}

func InitPluginClient() {
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
		ErrCh:     make(chan *meta.Topoerror, 10),
		AgentPort: conf.Config().Topo.Agent_port,
	}
}

// 初始化PAgentMap中的agent
func (t *Topoclient) InitMachineList() {
	Wait4TopoServerReady()
	pluginclient.GlobalClient.Wait4Bind()

	machine_list, err := pluginclient.GlobalClient.MachineList()
	if err != nil {
		err = errors.Errorf("%s **errstackfatal**2", err.Error()) // err top
		ErrorTransmit(pluginclient.GlobalContext, err, t.ErrCh, true)
	}

	for _, m := range machine_list {
		p := &Agent{}
		p.UUID = m.UUID
		p.Departname = m.Department
		p.IP = m.IP
		p.TAState = 0
		t.AddAgent_P(p)
	}
}

// 更新PAgentMap中的agent
func (t *Topoclient) UpdateMachineList() {
	machine_list, err := pluginclient.GlobalClient.MachineList()
	if err != nil {
		err = errors.Errorf("%s **errstackfatal**2", err.Error()) // err top
		ErrorTransmit(pluginclient.GlobalContext, err, t.ErrCh, true)
	}

	if Topo != nil {
		Topo.PAgentMap.Range(func(key, value interface{}) bool {
			t.DeleteAgent_P(key.(string))
			return true
		})
	} else {
		err := errors.New("agentmanager.Topo is nil, can not clear Topo.PAgentMap **errstackfatal**6") // err top
		ErrorTransmit(pluginclient.GlobalContext, err, t.ErrCh, true)
	}

	for _, m := range machine_list {
		p := &Agent{}
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
		ErrorTransmit(pluginclient.GlobalContext, err, t.ErrCh, true)
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
				if len(errarr) < 2 {
					logger.Error("topoerror type required in root error (err: %+v)", topoerr.Err)
					os.Exit(1)
				}

				switch errarr[1] {
				case "debug": // 只打印最底层error的message，不展开错误链的调用栈
					logger.Debug("%+v\n", strings.Split(errors.Cause(topoerr.Err).Error(), "**")[0])
				case "warn": // 只打印最底层error的message，不展开错误链的调用栈
					logger.Warn("%+v\n", strings.Split(errors.Cause(topoerr.Err).Error(), "**")[0])
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
