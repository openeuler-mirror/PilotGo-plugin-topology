package agentmanager

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"gitee.com/openeuler/PilotGo/sdk/plugin/client"
	"gitee.com/openeuler/PilotGo/sdk/utils/httputils"
	"github.com/mitchellh/mapstructure"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var Topo *Topoclient

type Topoclient struct {
	Sdkmethod *client.Client
	AgentMap  sync.Map
	Errmu     sync.Locker
	ErrCond   *sync.Cond
	ErrCh     chan error
	Out       io.Writer
}

func (t *Topoclient) InitMachineList() {
	for {
		if Topo.Sdkmethod.Server() != "" {
			break
		}

		logger.Debug("waiting for handshake with pilotgo server...")
		time.Sleep(2 * time.Second)
	}

	url := "http://" + Topo.Sdkmethod.Server() + "/api/v1/pluginapi/machine_list"

	resp, err := httputils.Get(url, nil)
	if err != nil {
		err = errors.Errorf("%s (url: %s) **fatal**2", err.Error(), url) // err top
		t.ErrCh <- err
		t.Errmu.Lock()
		t.ErrCond.Wait()
		t.Errmu.Unlock()
		close(t.ErrCh)
		os.Exit(1)
	}

	statuscode := resp.StatusCode
	if statuscode != 200 {
		err := errors.Errorf("http返回状态码异常: %d, %s **fatal**2", statuscode, url) // err top
		t.ErrCh <- err
		t.Errmu.Lock()
		t.ErrCond.Wait()
		t.Errmu.Unlock()
		close(t.ErrCh)
		os.Exit(1)
	}

	result := &struct {
		Code int         `json:"code"`
		Data interface{} `json:"data"`
	}{}

	err = json.Unmarshal(resp.Body, result)
	if err != nil {
		err = errors.Errorf("%s **fatal**2", err.Error()) // err top
		t.ErrCh <- err
		t.Errmu.Lock()
		t.ErrCond.Wait()
		t.Errmu.Unlock()
		close(t.ErrCh)
		os.Exit(1)
	}

	for _, m := range result.Data.([]interface{}) {
		p := &Agent_m{}
		mapstructure.Decode(m, p)
		p.TAState = 0
		Topo.AddAgent(p)
	}
}

func (t *Topoclient) UpdateMachineList() {
	url := "http://" + Topo.Sdkmethod.Server() + "/api/v1/pluginapi/machine_list"

	resp, err := httputils.Get(url, nil)
	if err != nil {
		err = errors.Errorf("%s **fatal**2", err.Error()) // err top
		t.ErrCh <- err
		t.Errmu.Lock()
		t.ErrCond.Wait()
		t.Errmu.Unlock()
		close(t.ErrCh)
		os.Exit(1)
	}

	statuscode := resp.StatusCode
	if statuscode != 200 {
		err = errors.Errorf("http返回状态码异常: %d **fatal**2", statuscode) // err top
		t.ErrCh <- err
		t.Errmu.Lock()
		t.ErrCond.Wait()
		t.Errmu.Unlock()
		close(t.ErrCh)
		os.Exit(1)
	}

	result := &struct {
		Code int         `json:"code"`
		Data interface{} `json:"data"`
	}{}

	err = json.Unmarshal(resp.Body, result)
	if err != nil {
		err = errors.Errorf("%s **fatal**2", err.Error()) // err top
		t.ErrCh <- err
		t.Errmu.Lock()
		t.ErrCond.Wait()
		t.Errmu.Unlock()
		close(t.ErrCh)
		os.Exit(1)
	}

	for _, m := range result.Data.([]interface{}) {
		p := &Agent_m{}
		mapstructure.Decode(m, p)
		p.TAState = 0

		agent := t.GetAgent(p.UUID)
		if agent == nil {
			Topo.AddAgent(p)
			return
		}

		agent.IP = p.IP
		agent.ID = p.ID
		agent.UUID = p.UUID
		agent.Port = p.Port
		agent.Departid = p.Departid
		agent.Departname = p.Departname
		agent.State = p.State
		agent.TAState = p.TAState
		Topo.AddAgent(agent)
	}
}

func (t *Topoclient) InitLogger() {
	err := logger.Init(conf.Config().Logopts)
	if err != nil {
		err = errors.Errorf("%s **fatal**2", err.Error()) // err top
		t.ErrCh <- err
		t.Errmu.Lock()
		t.ErrCond.Wait()
		t.Errmu.Unlock()
		close(t.ErrCh)
		os.Exit(1)
	}
}

func (t *Topoclient) InitPluginClient() {
	var errcondmu sync.Mutex
	PluginInfo.Url = "http://" + conf.Config().Topo.Server_addr + "/plugin/topology"
	PluginClient := client.DefaultClient(PluginInfo)

	Topo = &Topoclient{
		Sdkmethod: PluginClient,
		Errmu:     &errcondmu,
		ErrCond:   sync.NewCond(&errcondmu),
		ErrCh:     make(chan error, 10),
	}
}

func (t *Topoclient) InitErrorControl(errch <-chan error, emu sync.Locker, econd *sync.Cond) {
	switch conf.Global_config.Logopts.Driver {
	case "stdout":
		t.Out = os.Stdout
	case "file":
		logfile, err := os.OpenFile(conf.Global_config.Logopts.Path, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
		t.Out = logfile
	}

	go func(ch <-chan error, mu sync.Locker, ec *sync.Cond) {
		for {
			err, ok := <-ch
			if !ok {
				break
			}

			if err != nil {
				errarr := strings.Split(err.Error(), "**")
				switch errarr[1] {
				case "warn":
					fmt.Fprintf(t.Out, "%+v\n", err)
					// errors.EORE(err)
				case "fatal":
					mu.Lock()
					fmt.Fprintf(t.Out, "%+v\n", err)
					// errors.EORE(err)
					ec.Broadcast()
					mu.Unlock()
				default:
					fmt.Printf("only support warn and fatal error type: %+v\n", err)
					os.Exit(1)
				}
			}
		}
	}(errch, emu, econd)
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

func (t *Topoclient) SignalMonitoring(driver neo4j.Driver) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for {
		s := <-c
		switch s {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			driver.Close()
			fmt.Printf("close the connection to neo4j\n")
			os.Exit(-1)
		default:
			fmt.Printf("unknown signal: %s\n", s.String())
		}
	}
}
