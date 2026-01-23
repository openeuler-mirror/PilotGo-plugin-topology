package allinone

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"atomgit.com/openeuler/PilotGo-plugin-topology/cmd"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/http"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/logging"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/service"
	"github.com/avast/retry-go"
	"github.com/kardianos/osext"
	"github.com/spf13/cobra"

	"github.com/skydive-project/skydive/analyzer"
	"github.com/skydive-project/skydive/config"
)

var (
	analyzerUID uint32
)

// AllInOneCmd describes the allinone root command
var AllInOneCmd = &cobra.Command{
	Use:          "allinone",
	Short:        "Skydive All-In-One",
	Long:         "Skydive All-In-One (Analyzer + Agent)",
	SilenceUsage: true,
	Run: func(c *cobra.Command, args []string) {
		if uid := os.Geteuid(); uid != 0 {
			fmt.Fprintln(os.Stderr, "All-In-One mode has to be run as root")
			os.Exit(1)
		}

		skydivePath, _ := osext.Executable()
		logFile := config.GetString("logging.file.path")
		extension := filepath.Ext(logFile)
		logFile = strings.TrimSuffix(logFile, extension)
		os.Setenv("SKYDIVE_LOGGING_FILE_PATH", logFile+"-analyzer"+extension)

		analyzerAttr := &os.ProcAttr{
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
			Env:   os.Environ(),
		}

		if analyzerUID != 0 {
			analyzerAttr.Sys = &syscall.SysProcAttr{
				Credential: &syscall.Credential{
					Uid: analyzerUID,
				},
			}
		}

		args = []string{"skydive"}
		for _, cfgFile := range cmd.CfgFiles {
			args = append(args, "-c")
			args = append(args, cfgFile)
		}

		if cmd.CfgBackend != "" {
			args = append(args, "-b")
			args = append(args, cmd.CfgBackend)
		}

		analyzerArgs := make([]string, len(args))
		copy(analyzerArgs, args)

		if len(cmd.CfgFiles) != 0 {
			if err := cmdconfig.LoadConfiguration(cmd.CfgBackend, cmd.CfgFiles); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to initialize config: %s", err.Error())
				os.Exit(1)
			}
		}

		analyzerProcess, err := os.StartProcess(skydivePath, append(analyzerArgs, "analyzer"), analyzerAttr)
		if err != nil {
			logging.GetLogger().Errorf("Can't start Skydive analyzer: %v", err)
			os.Exit(1)
		}

		authOptions := analyzer.ClusterAuthenticationOpts()
		svcAddr, _ := service.AddressFromString(config.GetString("analyzer.listen"))
		tlsConfig, err := config.GetTLSClientConfig(true)
		if err != nil {
			logging.GetLogger().Errorf("Can't start Skydive analyzer: %v", err)
			os.Exit(1)
		}
		addr := svcAddr.Addr
		if addr == "0.0.0.0" {
			addr = "127.0.0.1"
		}

		url, err := config.GetURL("http", addr, svcAddr.Port, "")
		if err != nil {
			logging.GetLogger().Errorf("Invalid service address: %s", err)
			os.Exit(1)
		}
		restClient := http.NewRestClient(url, authOptions, tlsConfig)

		err = retry.Do(func() error {
			_, err := restClient.Request("GET", "/api", nil, nil)
			return err
		})
	},
}
