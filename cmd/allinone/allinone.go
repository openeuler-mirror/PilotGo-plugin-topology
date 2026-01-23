package allinone

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/kardianos/osext"
	"github.com/spf13/cobra"

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

	},
}
