package completion

import (
	"fmt"

	"github.com/spf13/cobra"
)

// BashCompletion skydive root command
var BashCompletion = &cobra.Command{
	Use:          "bash-completion",
	Short:        "Generate bash completion helper",
	Long:         "Generate bash completion helper (skydive-bash-completion.sh)",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Root().GenBashCompletionFile("skydive-bash-completion.sh")
		fmt.Println("skydive-bash-completion.sh has been generated")
	},
}
