package server

import (
	"fmt"

	"github.com/skydive-project/skydive/api/types"
)

func (pi *PacketInjectorAPI) validateRequest(ppr *types.PacketInjection) error {
	return fmt.Errorf("packet injection disabled because skydive was not built with \"packetinject\" build tag")
}
