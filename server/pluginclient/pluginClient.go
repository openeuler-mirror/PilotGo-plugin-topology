package pluginclient

import (
	"fmt"
	"os"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/conf"
	"gitee.com/openeuler/PilotGo/sdk/common"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"gitee.com/openeuler/PilotGo/sdk/plugin/client"
	"github.com/pkg/errors"
)

var Global_Client *client.Client

func InitPluginClient() {
	if conf.Global_Config != nil && conf.Global_Config.Topo.Https_enabled {
		PluginInfo.Url = fmt.Sprintf("https://%s", conf.Global_Config.Topo.Addr_target)
	} else if conf.Global_Config != nil && !conf.Global_Config.Topo.Https_enabled {
		PluginInfo.Url = fmt.Sprintf("http://%s", conf.Global_Config.Topo.Addr_target)
	} else {
		err := errors.New("Global_Config is nil")
		logger.Fatal("%+v", err)
	}

	Global_Client = client.DefaultClient(PluginInfo)

	GetExtentions()

	GetTags()

	go uploadResource()
}

func uploadResource() {
	for !Global_Client.IsBind() {
		time.Sleep(100 * time.Millisecond)
	}

	dirPath := conf.Global_Config.Topo.Path
	filename_list := []string{}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		logger.Error("fail to read files: %s", err.Error())
		return
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filename_list = append(filename_list, file.Name())
	}

	for _, filename := range filename_list {
		err := Global_Client.FileUpload(dirPath, filename)
		if err != nil {
			logger.Error("fail to upload file: %s", err.Error())
		}
	}
}

// 注册插件扩展点
func GetExtentions() {
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
	me1 := &common.MachineExtention{
		Type:       common.ExtentionMachine,
		Name:       "部署topo-collect",
		URL:        "/plugin/topology/api/deploy_collect_endpoint",
		Permission: "plugin.topology.agent/install",
	}
	me2 := &common.MachineExtention{
		Type:       common.ExtentionMachine,
		Name:       "停用topo-collect",
		URL:        "/plugin/topology/api/collect_endpoint?action=stop",
		Permission: "plugin.topology.agent/stop",
	}
	me3 := &common.MachineExtention{
		Type:       common.ExtentionMachine,
		Name:       "卸载topo-collect",
		URL:        "/plugin/topology/api/collect_endpoint?action=remove",
		Permission: "plugin.topology.agent/remove",
	}
	ex = append(ex, pe1, pe2, me1, me2, me3)
	Global_Client.RegisterExtention(ex)
}

func GetTags() {
	tag_cb := func(uuids []string) []common.Tag {
		var tags []common.Tag
		for _, uuid := range uuids {
			tag := common.Tag{
				UUID: uuid,
				Type: common.TypeOk,
				Data: "topo-collect",
			}
			tags = append(tags, tag)
		}
		return tags
	}
	Global_Client.OnGetTags(tag_cb)
}
