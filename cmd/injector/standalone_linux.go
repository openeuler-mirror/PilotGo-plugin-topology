package injector

import (
	"fmt"
	"os"
	"time"

	"github.com/skydive-project/go-debouncer"
	"github.com/skydive-project/skydive/cmd/client"
	"github.com/spf13/cobra"
)

var (
	encapType string
	ifName    string
)

// InjectPacketCmd injects packets without using any Skydive analyzer/agent
var InjectPacketCmd = &cobra.Command{
	Use:          "inject-packet",
	Short:        "inject packet",
	Long:         "inject packet",
	SilenceUsage: false,
	Run: func(cmd *cobra.Command, args []string) {
		pir, err := client.GetPacketInjectRequest()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			cmd.Usage()
			os.Exit(1)
		}

		if ifName == "" {
			if ifName = getNameFromMac(pir.SrcMAC); ifName == "" {
				fmt.Fprintf(os.Stderr, "Could not find interface with MAC %s", pir.SrcMAC)
				os.Exit(1)
			}
		}

		rawSocket, err := rawsocket.NewRawSocket(ifName, rawsocket.OnlyIPPackets)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open socket for interface %s, err: %s\n", ifName, err)
			os.Exit(1)
		}
		defer rawSocket.Close()

		var packetForger pi.PacketForger
		if len(pir.Pcap) > 0 {
			packetForger, err = pi.NewPcapPacketGenerator(pir)
		} else {
			packetForger, err = pi.NewForgedPacketGenerator(pir, encapType)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create packet generator for interface %s, err: %s\n", ifName, err)
			os.Exit(1)
		}
		defer packetForger.Close()

		lastPacketCount := 0
		showPacketCount := func() {
			if lastPacketCount != 0 {
				fmt.Fprintf(os.Stderr, "%d packet(s) injected on interface %s\n", lastPacketCount, ifName)
			}
			lastPacketCount = 0
		}

		debouncer := debouncer.New(time.Second, showPacketCount)
		debouncer.Start()
		defer debouncer.Stop()
	},
}

func init() {
	InjectPacketCmd.Flags().StringVarP(&ifName, "if-name", "", "", "interface to inject packets into")
	InjectPacketCmd.Flags().StringVarP(&encapType, "encap-type", "", "ether", "encapsulation type")
	client.AddInjectPacketInjectFlags(InjectPacketCmd)
}
