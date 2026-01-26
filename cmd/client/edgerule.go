package client

import (
	"github.com/spf13/cobra"
)

var (
	src          string
	dst          string
	relationType string
)

// EdgeRuleCmd skydive edge rule root command
var EdgeRuleCmd = &cobra.Command{
	Use:          "edge-rule",
	Short:        "edge-rule",
	Long:         "edge-rule",
	SilenceUsage: false,
}

func init() {
	EdgeRuleCmd.AddCommand(EdgeRuleCreate)
	EdgeRuleCmd.AddCommand(EdgeRuleList)
	EdgeRuleCmd.AddCommand(EdgeRuleGet)
	EdgeRuleCmd.AddCommand(EdgeRuleDelete)

	addCreateEdgeRuleFlags(EdgeRuleCreate)
}
