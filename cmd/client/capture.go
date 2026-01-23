package client

import (
	"errors"
	"fmt"
	"time"

	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/http"
	"github.com/go-playground/validator/v10"
	"github.com/skydive-project/skydive/flow"
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

// CaptureCreate skydive capture creates command
var CaptureCreate = &cobra.Command{
	Use:   "create",
	Short: "Create capture",
	Long:  "Create capture",
	PreRun: func(cmd *cobra.Command, args []string) {
		if nodeTID != "" {
			if captureQuery != "" {
				exitOnError(errors.New("Options --node and --gremlin are exclusive"))
			}
			captureQuery = fmt.Sprintf("g.V().Has('TID', '%s')", nodeTID)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		var layers flow.ExtraLayers
		if err := layers.Parse(extraLayers...); err != nil {
			exitOnError(err)
		}

		capture := api.NewCapture(captureQuery, bpfFilter)
		capture.Name = captureName
		capture.Description = captureDescription
		capture.Type = captureType
		capture.Port = port
		capture.SamplingRate = samplingRate
		capture.PollingInterval = pollingInterval
		capture.HeaderSize = headerSize
		capture.ExtraTCPMetric = extraTCPMetric
		capture.IPDefrag = ipDefrag
		capture.ReassembleTCP = reassembleTCP
		capture.LayerKeyMode = layerKeyMode
		capture.RawPacketLimit = rawPacketLimit
		capture.ExtraLayers = int(layers)
		capture.Target = target
		capture.TargetType = targetType

		if err := validator.Validate("capture", capture); err != nil {
			exitOnError(err)
		}

		var createOpts *http.CreateOptions
		if captureTTL != 0 {
			createOpts = &http.CreateOptions{
				TTL: time.Duration(captureTTL) * time.Millisecond,
			}
		}

		if err := CrudClient.Create("capture", &capture, createOpts); err != nil {
			exitOnError(err)
		}
		printJSON(&capture)
	},
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
