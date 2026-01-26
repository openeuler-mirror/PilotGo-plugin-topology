package client

import (
	"fmt"
	"os"

	api "github.com/skydive-project/skydive/api/types"
	"github.com/skydive-project/skydive/graffiti/graph"
	"github.com/skydive-project/skydive/graffiti/logging"
	"github.com/skydive-project/skydive/validator"

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

// EdgeRuleCreate skydive edge create command
var EdgeRuleCreate = &cobra.Command{
	Use:          "create",
	Short:        "create",
	Long:         "create",
	SilenceUsage: false,
	PreRun: func(cmd *cobra.Command, args []string) {
		if relationType == "" {
			logging.GetLogger().Error("--relationtype is a required parameter")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		m, err := graph.DefToMetadata(metadata, graph.Metadata{})
		if err != nil {
			exitOnError(err)
		}
		m["RelationType"] = relationType

		edgeRule := &api.EdgeRule{
			Name:        name,
			Description: description,
			Src:         src,
			Dst:         dst,
			Metadata:    m,
		}

		if err = validator.Validate("edgerule", edgeRule); err != nil {
			exitOnError(fmt.Errorf("Error while validating edge rule: %s", err))
		}

		if err = CrudClient.Create("edgerule", &edgeRule, nil); err != nil {
			exitOnError(err)
		}

		printJSON(edgeRule)
	},
}

// EdgeRuleList skydive edge list command
var EdgeRuleList = &cobra.Command{
	Use:          "list",
	Short:        "list",
	Long:         "list",
	SilenceUsage: false,

	Run: func(cmd *cobra.Command, args []string) {
		var edges map[string]api.EdgeRule
		if err := CrudClient.List("edgerule", &edges); err != nil {
			exitOnError(err)
		}
		printJSON(edges)
	},
}

func init() {
	EdgeRuleCmd.AddCommand(EdgeRuleCreate)
	EdgeRuleCmd.AddCommand(EdgeRuleList)
	EdgeRuleCmd.AddCommand(EdgeRuleGet)
	EdgeRuleCmd.AddCommand(EdgeRuleDelete)

	addCreateEdgeRuleFlags(EdgeRuleCreate)
}
