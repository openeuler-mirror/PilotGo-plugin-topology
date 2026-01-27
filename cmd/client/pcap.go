package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/skydive-project/skydive/graffiti/logging"

	"github.com/spf13/cobra"
)

var (
	pcapTrace string
)

// PcapCmd skydive pcap root command
var PcapCmd = &cobra.Command{
	Use:   "pcap",
	Short: "Import flows from PCAP file",
	Long:  "Import flows from PCAP file",
	PreRun: func(cmd *cobra.Command, args []string) {
		if pcapTrace == "" {
			logging.GetLogger().Error("You need to specify a PCAP file")
			cmd.Usage()
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		file, err := os.Open(pcapTrace)
		if err != nil {
			exitOnError(err)
		}
		defer file.Close()

		resp, err := CrudClient.Request("POST", "pcap", file, nil)
		if err != nil {
			exitOnError(err)
		}

		if resp.StatusCode == http.StatusAccepted {
			fmt.Printf("%s was successfully imported\n", pcapTrace)
		} else {
			content, _ := ioutil.ReadAll(resp.Body)
			exitOnError(fmt.Errorf("Failed to import %s: %s", pcapTrace, string(content)))
		}
	},
}

func init() {
	PcapCmd.Flags().StringVarP(&pcapTrace, "trace", "t", "", "PCAP trace file to read")
}
