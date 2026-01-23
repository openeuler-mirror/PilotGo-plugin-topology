package agent

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/skydive-project/skydive/agent"
	"github.com/skydive-project/skydive/config"
	"github.com/skydive-project/skydive/graffiti/logging"
	"github.com/skydive-project/skydive/version"

	"github.com/spf13/cobra"
)

// AgentCmd describe the skydive agent root command
var AgentCmd = &cobra.Command{
	Use:          "agent",
	Short:        "Skydive agent",
	Long:         "Skydive agent",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		config.Set("logging.id", "agent")
		logging.GetLogger().Noticef("Skydive Agent %s starting...", version.Version)

		agent, err := agent.NewAgent()
		if err != nil {
			logging.GetLogger().Errorf("Can't start Skydive agent: %v", err)
			os.Exit(1)
		}

		agent.Start()

		logging.GetLogger().Notice("Skydive Agent started")
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch

		agent.Stop()

		logging.GetLogger().Notice("Skydive Agent stopped.")
	},
}

func init() {
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	AgentCmd.Flags().String("host-id", host, "ID used to reference the agent, defaults to hostname")
	config.BindPFlag("host_id", AgentCmd.Flags().Lookup("host-id"))

	AgentCmd.Flags().String("listen", "127.0.0.1:8081", "address and port for the agent API")
	config.BindPFlag("agent.listen", AgentCmd.Flags().Lookup("listen"))
}
