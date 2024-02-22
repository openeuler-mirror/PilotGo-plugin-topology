package agentmanager

import (
	"context"
	"fmt"
	"os"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"gitee.com/openeuler/PilotGo/sdk/plugin/client"
)

const Version = "1.0.1"

var PluginInfo = &client.PluginInfo{
	Name:        "topology",
	Version:     Version,
	Description: "System application architecture detection.",
	Author:      "wangjunqi",
	Email:       "wangjunqi@kylinos.cn",
	Url:         "http://10.1.10.131:9991/plugin/topology",
	PluginType:  "iframe",
}

func WaitingForHandshake() {
	i := 0
	loop := []string{`*.....`, `.*....`, `..*...`, `...*..`, `....*.`, `.....*`}
	for {
		if Topo != nil && Topo.Sdkmethod != nil && Topo.Sdkmethod.Server() != "" {
			break
		}
		fmt.Printf("\r Waiting for handshake with pilotgo server%s", loop[i])
		if i < 5 {
			i++
		} else {
			i = 0
		}
		time.Sleep(1 * time.Second)
	}
}

func ErrorTransmit(ctx context.Context, err error, ch chan *meta.Topoerror, exit_after_print bool) {
	if exit_after_print {
		cctx, cancelF := context.WithCancel(ctx)
		ch <- &meta.Topoerror{
			Err:    err,
			Cancel: cancelF,
		}
		<-cctx.Done()
		close(ch)
		os.Exit(1)
	}

	ch <- &meta.Topoerror{
		Err: err,
	}
}
