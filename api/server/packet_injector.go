package server

import (
	"errors"
	"fmt"

	"github.com/skydive-project/skydive/api/types"
)

func (pi *PacketInjectorAPI) validateRequest(ppr *types.PacketInjection) error {
	pi.Graph.RLock()
	defer pi.Graph.RUnlock()

	srcNode := pi.getNode(ppr.Src)
	dstNode := pi.getNode(ppr.Dst)

	if srcNode == nil {
		return fmt.Errorf("Not able to find a source node for '%s'", ppr.Src)
	}

	ipField := "IPV4"
	if ppr.Type == types.PITypeICMP6 || ppr.Type == types.PITypeTCP6 || ppr.Type == types.PITypeUDP6 {
		ipField = "IPV6"
	}

	if len(ppr.Pcap) == 0 {
		ips, _ := srcNode.GetFieldStringList(ipField)
		if len(ips) == 0 && ppr.SrcIP == "" {
			return errors.New("No source IP in node")
		}

		if dstNode == nil && ppr.DstIP == "" {
			return errors.New("No destination node and IP")
		}

		dstIP := ppr.DstIP
		if dstIP == "" {
			ips, _ := dstNode.GetFieldStringList(ipField)
			if len(ips) == 0 {
				return errors.New("No destination IP in node")
			}
			dstIP = ips[0]
		}
	}
	return nil
}
