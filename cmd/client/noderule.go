package client

import (
	"os"

	"github.com/skydive-project/skydive/graffiti/logging"
	"github.com/spf13/cobra"
)

var (
	name        string
	description string
	metadata    string
	nodeName    string
	nodeType    string
	query       string
	action      string
)

// NodeRuleCmd skydive node rule root command
var NodeRuleCmd = &cobra.Command{
	Use:          "node-rule",
	Short:        "node-rule",
	Long:         "node-rule",
	SilenceUsage: false,
}

// NodeRuleGet skydive node get command
var NodeRuleGet = &cobra.Command{
	Use:          "get",
	Short:        "get",
	Long:         "get",
	SilenceUsage: false,

	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Usage()
			os.Exit(1)
		}
	},

	Run: func(cmd *cobra.Command, args []string) {
		var node api.NodeRule
		if err := CrudClient.Get("noderule", args[0], &node); err != nil {
			logging.GetLogger().Error(err.Error())
			os.Exit(1)
		}
		printJSON(&node)
	},
}
