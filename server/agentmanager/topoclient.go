package agentmanager

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"gitee.com/openeuler/PilotGo/sdk/plugin/client"
	"gitee.com/openeuler/PilotGo/sdk/utils/httputils"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var Topo *Topoclient

type Topoclient struct {
	Sdkmethod *client.Client
	AgentMap  sync.Map
	ErrGroup  *sync.WaitGroup
	ErrCh     chan error
	Out       io.Writer
}

func (t *Topoclient) InitMachineList() {
	url := Topo.Sdkmethod.Server + "/api/v1/pluginapi/machine_list"

	resp, err := httputils.Get(url, nil)
	if err != nil {
		err = errors.Errorf("%s **fatal**2", err.Error()) // err top
		t.ErrCh <- err
		t.ErrGroup.Add(1)
		t.ErrGroup.Wait()
		close(t.ErrCh)
		os.Exit(1)
	}

	statuscode := resp.StatusCode
	if statuscode != 200 {
		err = errors.Errorf("http返回状态码异常: %d; %s **fatal**2", statuscode, url) // err top
		t.ErrCh <- err
		t.ErrGroup.Add(1)
		t.ErrGroup.Wait()
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
		t.ErrGroup.Add(1)
		t.ErrGroup.Wait()
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
	url := Topo.Sdkmethod.Server + "/api/v1/pluginapi/machine_list"

	resp, err := httputils.Get(url, nil)
	if err != nil {
		err = errors.Errorf("%s **fatal**2", err.Error()) // err top
		t.ErrCh <- err
		t.ErrGroup.Add(1)
		t.ErrGroup.Wait()
		close(t.ErrCh)
		os.Exit(1)
	}

	statuscode := resp.StatusCode
	if statuscode != 200 {
		err = errors.Errorf("http返回状态码异常: %d **fatal**2", statuscode) // err top
		t.ErrCh <- err
		t.ErrGroup.Add(1)
		t.ErrGroup.Wait()
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
		t.ErrGroup.Add(1)
		t.ErrGroup.Wait()
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
		t.ErrGroup.Add(1)
		t.ErrGroup.Wait()
		close(t.ErrCh)
		os.Exit(1)
	}
}

func (t *Topoclient) InitPluginClient() {
	PluginInfo.Url = "http://" + conf.Config().Topo.Server_addr + "/plugin/topology"
	PluginClient := client.DefaultClient(PluginInfo)
	PluginClient.Server = "http://" + conf.Config().PilotGo.Addr
	Topo = &Topoclient{
		Sdkmethod: PluginClient,
		ErrGroup:  &sync.WaitGroup{},
		ErrCh:     make(chan error, 10),
	}
}

func (t *Topoclient) InitErrorControl(errch <-chan error, errgroup *sync.WaitGroup) {
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

	go func(ch <-chan error, group *sync.WaitGroup) {
		for {
			err, ok := <-errch
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
					fmt.Fprintf(t.Out, "%+v\n", err)
					// errors.EORE(err)
					errgroup.Done()
				default:
					fmt.Printf("only support warn and fatal error type: %+v\n", err)
					os.Exit(1)
				}
			}
		}
	}(errch, errgroup)
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
