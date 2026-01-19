package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	api "atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/api/server"
	etcdclient "atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/etcd/client"
	etcdserver "atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/etcd/server"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	shttp "atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/http"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/hub"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/logging"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/service"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/websocket"
)

const (
	defaultQueueSize = 10000
)

var (
	hubListen        string
	embeddedEtcd     bool
	etcdServers      []string
	writeCompression bool
	queueSize        int
	pingDelay        int
	pongTimeout      int
)

// HubCmd describes the graffiti hub command
var HubCmd = &cobra.Command{
	Use:          "hub",
	Short:        "Graffiti hub",
	Long:         "Graffiti hub",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		logging.GetLogger().Noticef("Graffiti hub starting...")

		hostname, err := os.Hostname()
		if err != nil {
			logging.GetLogger().Errorf("Failed to get hostname: %s", err)
			os.Exit(1)
		}

		authBackend := shttp.NewNoAuthenticationBackend()

		if len(etcdServers) == 0 {
			logging.GetLogger().Error("No Etcd server provided")
			os.Exit(1)
		}

		etcdClientOpts := etcdclient.Opts{
			Servers: etcdServers,
			Timeout: 5 * time.Second,
		}

		etcdClient, err := etcdclient.NewClient(hostname, etcdClientOpts)
		if err != nil {
			logging.GetLogger().Error(err)
			os.Exit(1)
		}

		origin := "graffiti-hub"
		cached, err := graph.NewCachedBackend(nil, etcdClient, hostname, service.Type(origin))
		if err != nil {
			logging.GetLogger().Error(err)
			os.Exit(1)
		}

		g := graph.NewGraph(hostname, cached, origin)

		hubOpts := hub.Opts{
			Hostname: hostname,
			WebsocketOpts: websocket.ServerOpts{
				WriteCompression: writeCompression,
				QueueSize:        queueSize,
				PingDelay:        time.Second * time.Duration(pingDelay),
				PongTimeout:      time.Second * time.Duration(pongTimeout),
			},
			APIAuthBackend:     authBackend,
			ClusterAuthBackend: authBackend,
			EtcdClient:         etcdClient,
		}

		var etcdServer *etcdserver.EmbeddedServer
		if embeddedEtcd {
			hubOpts.EtcdServerOpts = &etcdserver.EmbeddedServerOpts{
				Name:    "localhost",
				Listen:  "127.0.0.1:12379",
				DataDir: "/tmp/etcd",
			}
		}

		hub, err := hub.NewHub(hostname, service.Type("Hub"), hubListen, g, cached, "/ws/pod", hubOpts)
		if err != nil {
			logging.GetLogger().Error(err)
			os.Exit(1)
		}

		api.RegisterStatusAPI(hub.HTTPServer(), hub, authBackend)

		hub.Start()

		logging.GetLogger().Notice("Graffiti hub started !")
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch

		hub.Stop()
		if etcdServer != nil {
			etcdServer.Stop()
		}

		logging.GetLogger().Notice("Graffiti hub stopped.")
	},
}

func init() {
	defaultEtcdAddr := fmt.Sprintf("http://%s:%d", etcdclient.DefaultServer, etcdclient.DefaultPort)
	HubCmd.Flags().StringVarP(&hubListen, "listen", "l", "127.0.0.1:8082", "address and port for the hub server")
	HubCmd.Flags().IntVar(&queueSize, "queue-size", 10000, "websocket queue size")
	HubCmd.Flags().IntVar(&pingDelay, "ping-delay", 2, "websocket ping delay")
	HubCmd.Flags().IntVar(&pongTimeout, "pong-timeout", 10, "websocket pong timeout")
	HubCmd.Flags().BoolVar(&embeddedEtcd, "embedded-etcd", false, "run embedded etcd server")
	HubCmd.Flags().StringArrayVar(&etcdServers, "etcd-servers", []string{defaultEtcdAddr}, "etcd servers")
}
