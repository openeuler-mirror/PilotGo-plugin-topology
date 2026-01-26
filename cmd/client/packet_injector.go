package client

import (
	"github.com/spf13/cobra"
)

var (
	srcNode          string
	dstNode          string
	srcIP            string
	srcMAC           string
	srcPort          uint16
	dstPort          uint16
	dstIP            string
	dstMAC           string
	packetType       string
	payload          string
	pcap             string
	id               uint16
	count            uint64
	interval         uint64
	mode             string
	incrementPayload int64
	ttl              uint8
)

// AddInjectPacketInjectFlags add the command line flags for a packet injection
func AddInjectPacketInjectFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&srcIP, "src-ip", "", "", "source node IP")
	cmd.Flags().StringVarP(&dstIP, "dst-ip", "", "", "destination node IP")
	cmd.Flags().StringVarP(&srcMAC, "src-mac", "", "", "source node MAC")
	cmd.Flags().StringVarP(&dstMAC, "dst-mac", "", "", "destination node MAC")
	cmd.Flags().Uint16VarP(&srcPort, "src-port", "", 0, "source port for TCP packet")
	cmd.Flags().Uint16VarP(&dstPort, "dst-port", "", 0, "destination port for TCP packet")
	cmd.Flags().StringVarP(&packetType, "type", "", "icmp4", "packet type: icmp4, icmp6, tcp4, tcp6, udp4 and udp6")
	cmd.Flags().StringVarP(&payload, "payload", "", "", "payload")
	cmd.Flags().StringVar(&pcap, "pcap", "", "PCAP file")
	cmd.Flags().Uint16VarP(&id, "id", "", 0, "ICMP identification")
	cmd.Flags().StringVarP(&mode, "mode", "", "unique", "specify mode of packet generation, `unique` or `random`")
	cmd.Flags().Int64VarP(&incrementPayload, "inc-payload", "", 0, "increase payload size each packet")
	cmd.Flags().Uint64VarP(&count, "count", "", 1, "number of packets to be generated")
	cmd.Flags().Uint64VarP(&interval, "interval", "", 0, "wait interval milliseconds between sending each packet")
	cmd.Flags().Uint8VarP(&ttl, "ttl", "", 64, "IP time-to-live header")
}

// PacketInjectorCmd skydive inject-packet root command
var PacketInjectorCmd = &cobra.Command{
	Use:          "injection",
	Short:        "Inject packets",
	Long:         "Inject packets",
	Aliases:      []string{"inject-packet"},
	SilenceUsage: false,
}
