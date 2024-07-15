package handler

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/pluginclient"
	"gitee.com/openeuler/PilotGo/sdk/common"
	"gitee.com/openeuler/PilotGo/sdk/response"
	"github.com/gin-gonic/gin"
)

func DeployCollectEndpointHandle(ctx *gin.Context) {
	uuids := &struct {
		MachineUUIDs []string `json:"uuids"`
	}{}
	if err := ctx.ShouldBind(uuids); err != nil {
		err = errors.Errorf("%s **errstack**0", err.Error())
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		response.Fail(ctx, nil, "parameter error")
		return
	}

	file, err := os.Open(strings.TrimSuffix(conf.Global_Config.Topo.Path, "/") + "/deploy-collect-endpoint.sh")
	if err != nil {
		err = errors.Errorf("%s **errstack**0", err.Error())
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		response.Fail(ctx, nil, "open file error: "+errors.Cause(err).Error())
		return
	}
	defer file.Close()
	script_bytes, err := io.ReadAll(file)
	if err != nil {
		err = errors.Errorf("%s **errstack**0", err.Error())
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		response.Fail(ctx, nil, "read file error: "+errors.Cause(err).Error())
		return
	}

	batch := &common.Batch{
		MachineUUIDs: uuids.MachineUUIDs,
	}
	cmdresults, err := pluginclient.Global_Client.RunScript(batch, string(script_bytes), []string{
		"--workdir=/root/test-topo",
		fmt.Sprintf("--pilotgoserver=%s", pluginclient.Global_Client.Server()),
		fmt.Sprintf("--toposerver=%s", strings.Split(pluginclient.Global_Client.PluginInfo.Url, "//")[1]),
		"--fleet=10.41.161.101:8220",
	})
	if err != nil {
		err = errors.Errorf("%s **errstack**0", err.Error())
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
		response.Fail(ctx, nil, errors.Cause(err).Error())
		return
	}
	for _, res := range cmdresults {
		err := errors.Errorf("collect endpoint deploy: [retcode:%d][uuid:%s][ip:%s][stdout:%s][stderr:%s] **warn**0", res.RetCode, res.MachineUUID, res.MachineIP, res.Stdout, res.Stderr)
		errormanager.ErrorTransmit(pluginclient.Global_Context, err, false)
	}

	response.Success(ctx, nil, "采集端部署完成")
}
