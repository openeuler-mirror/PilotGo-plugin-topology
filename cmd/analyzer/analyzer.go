package analyzer

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/skydive-project/skydive/analyzer"
	"github.com/skydive-project/skydive/config"
	"github.com/skydive-project/skydive/graffiti/logging"
	"github.com/skydive-project/skydive/version"

	"github.com/spf13/cobra"
)

// AnalyzerCmd describes the skydive analyzer root command
var AnalyzerCmd = &cobra.Command{
	Use:          "analyzer",
	Short:        "Skydive analyzer",
	Long:         "Skydive analyzer",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		config.Set("logging.id", "analyzer")
		logging.GetLogger().Noticef("Skydive Analyzer %s starting...", version.Version)

		server, err := analyzer.NewServerFromConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create analyzer: %s", err)
			os.Exit(1)
		}

		if err := server.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start analyzer: %s", err)
			os.Exit(1)
		}

		logging.GetLogger().Notice("Skydive Analyzer started !")
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch

		server.Stop()

		logging.GetLogger().Notice("Skydive Analyzer stopped.")
	},
}

func init() {
	AnalyzerCmd.Flags().String("listen", "127.0.0.1:8082", "address and port for the analyzer API")
	config.BindPFlag("analyzer.listen", AnalyzerCmd.Flags().Lookup("listen"))
}
