package injector

import (
	"github.com/skydive-project/skydive/cmd/client"
)

var (
	encapType string
	ifName    string
)

func init() {
	InjectPacketCmd.Flags().StringVarP(&ifName, "if-name", "", "", "interface to inject packets into")
	InjectPacketCmd.Flags().StringVarP(&encapType, "encap-type", "", "ether", "encapsulation type")
	client.AddInjectPacketInjectFlags(InjectPacketCmd)
}
