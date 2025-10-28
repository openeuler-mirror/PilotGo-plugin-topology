package graph

import (
	"fmt"

	"github.com/skydive-project/skydive/graffiti/filters"
	"github.com/skydive-project/skydive/graffiti/service"
	ws "github.com/skydive-project/skydive/graffiti/websocket"
)

// ClientOrigin return a string identifying a client using its service type and host id
func ClientOrigin(c ws.Speaker) string {
	return string(c.GetServiceType()) + "." + c.GetRemoteHost()
}

// DelSubGraphOfOrigin deletes all the nodes with a specified origin
func DelSubGraphOfOrigin(g *Graph, origin string) {
	filter := filters.NewTermStringFilter("@Origin", origin)
	g.DelNodes(NewElementFilter(filter))
}

// Origin returns string representation of origin components
func Origin(hostID string, kind service.Type) string {
	return fmt.Sprintf("%s.%s", kind, hostID)
}
