package client

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	gclient "github.com/skydive-project/skydive/graffiti/cmd/client"
	"github.com/skydive-project/skydive/graffiti/http"
	"github.com/skydive-project/skydive/graffiti/logging"
)

var (
	host         string
	analyzerAddr string

	// ClientCmd describes the skydive client root command
	ClientCmd = gclient.ClientCmd
	// CrudClient holds the client API CRUD client
	CrudClient *http.CrudClient
)

func exitOnError(err error) {
	logging.GetLogger().Error(err)
	os.Exit(1)
}

func printJSON(obj interface{}) {
	s, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		logging.GetLogger().Error(err)
		os.Exit(1)
	}
	fmt.Println(string(s))
}

// RegisterClientCommands registers the 'client' CLI subcommands
func RegisterClientCommands(cmd *cobra.Command) {
	cmd.AddCommand(CaptureCmd)
	cmd.AddCommand(EdgeRuleCmd)
	cmd.AddCommand(NodeRuleCmd)
	cmd.AddCommand(PacketInjectorCmd)
	cmd.AddCommand(PcapCmd)
	cmd.AddCommand(StatusCmd)
}
