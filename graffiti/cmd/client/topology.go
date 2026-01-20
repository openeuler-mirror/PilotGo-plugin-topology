package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	endpoints "atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/endpoints"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/graph"
	shttp "atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/http"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/messages"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/service"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/websocket"
)

var (
	gremlinQuery      string
	filename          string
	publisherEndpoint string
)

// TopologyCmd skydive topology root command
var TopologyCmd = &cobra.Command{
	Use:          "topology",
	Short:        "Request on topology",
	Long:         "Request on topology",
	SilenceUsage: false,
}

// TopologyRequest skydive topology query command
var TopologyRequest = &cobra.Command{
	Use:   "query",
	Short: "query topology [deprecated: use 'client query' instead]",
	Long:  "query topology [deprecated: use 'client query' instead]",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr, "The 'client topology query' command is deprecated. Please use 'client query' instead")
		QueryCmd.Run(cmd, []string{gremlinQuery})
	},
}

// TopologyImport skydive topology import command
var TopologyImport = &cobra.Command{
	Use:   "import",
	Short: "import topology",
	Long:  "import topology",
	Run: func(cmd *cobra.Command, args []string) {
		url, err := shttp.MakeURL("ws", hubService.Addr, hubService.Port, publisherEndpoint, tlsConfig != nil)
		if err != nil {
			exitOnError(err)
		}
		opts := websocket.ClientOpts{AuthOpts: &AuthenticationOpts, Headers: http.Header{}, TLSConfig: tlsConfig}
		opts.Headers.Add("X-Persistence-Policy", string(endpoints.Persistent))
		client := websocket.NewClient(Host, service.Type("CLI"), url, opts)

		if err := client.Connect(context.Background()); err != nil {
			exitOnError(err)
		}

		go client.Run()
		defer func() {
			client.Flush()
			client.StopAndWait()
		}()

		content, err := os.ReadFile(filename)
		if err != nil {
			exitOnError(err)
		}

		els := []*graph.Elements{}
		if err := json.Unmarshal(content, &els); err != nil {
			exitOnError(err)
		}

		if len(els) != 1 {
			exitOnError(errors.New("Invalid graph format"))
		}

		for _, node := range els[0].Nodes {
			msg := messages.NewStructMessage(messages.NodeAddedMsgType, node)
			if err := client.SendMessage(msg); err != nil {
				exitOnError(fmt.Errorf("Failed to send message: %s", err))
			}
		}

		for _, edge := range els[0].Edges {
			msg := messages.NewStructMessage(messages.EdgeAddedMsgType, edge)
			if err := client.SendMessage(msg); err != nil {
				exitOnError(fmt.Errorf("Failed to send message: %s", err))
			}
		}
	},
}

// TopologyExport skydive topology export command
var TopologyExport = &cobra.Command{
	Use:   "export",
	Short: "export topology",
	Long:  "export topology",
	Run: func(cmd *cobra.Command, args []string) {
		QueryCmd.Run(cmd, []string{"G"})
	},
}

func init() {
	TopologyCmd.AddCommand(TopologyExport)

	TopologyImport.Flags().StringVarP(&filename, "file", "", "graph.json", "Input file")
	TopologyImport.Flags().StringVarP(&publisherEndpoint, "endpoint", "", "/ws/publisher", "Publisher WebSocket endpoint")
	TopologyCmd.AddCommand(TopologyImport)

	TopologyRequest.Flags().StringVarP(&gremlinQuery, "gremlin", "", "G", "Gremlin Query")
	TopologyRequest.Flags().StringVarP(&outputFormat, "format", "", "json", "Output format (json, dot or pcap)")
	TopologyCmd.AddCommand(TopologyRequest)
}
