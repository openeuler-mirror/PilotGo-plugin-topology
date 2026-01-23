package client

import (
	"github.com/spf13/cobra"
)

var (
	bpfFilter          string
	captureName        string
	captureDescription string
	captureType        string
	captureTTL         uint64
	captureQuery       string
	nodeTID            string
	port               int
	samplingRate       uint32
	pollingInterval    uint32
	headerSize         int
	rawPacketLimit     int
	extraTCPMetric     bool
	ipDefrag           bool
	reassembleTCP      bool
	layerKeyMode       string
	extraLayers        []string
	target             string
	targetType         string
)

// CaptureCmd skydive capture root command
var CaptureCmd = &cobra.Command{
	Use:          "capture",
	Short:        "Manage captures",
	Long:         "Manage captures",
	SilenceUsage: false,
}

// CaptureList skydive capture list command
var CaptureList = &cobra.Command{
	Use:   "list",
	Short: "List captures",
	Long:  "List captures",
	Run: func(cmd *cobra.Command, args []string) {
		var captures map[string]api.Capture
		if err := CrudClient.List("capture", &captures); err != nil {
			exitOnError(err)
		}
		printJSON(captures)
	},
}

func init() {
	CaptureCmd.AddCommand(CaptureList)
	CaptureCmd.AddCommand(CaptureCreate)
	CaptureCmd.AddCommand(CaptureGet)
	CaptureCmd.AddCommand(CaptureDelete)

	addCaptureFlags(CaptureCreate)
}
