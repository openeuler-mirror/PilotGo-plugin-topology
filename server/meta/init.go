// 禁止调用utils包
package meta

var NODE_TYPES []string
var EDGE_TYPES []string

const (
	NODE_HOST     = "host"
	NODE_PROCESS  = "process"
	NODE_THREAD   = "thread"
	NODE_NET      = "net"
	NODE_APP      = "app"
	NODE_RESOURCE = "resource"
)

const (
	EDGE_BELONG = "belong"
	EDGE_SERVER = "server"
	EDGE_CLIENT = "client"
	EDGE_TCP    = "tcp"
	EDGE_UDP    = "udp"
)

func init() {
	NODE_TYPES = []string{NODE_HOST, NODE_PROCESS, NODE_THREAD, NODE_NET, NODE_APP, NODE_RESOURCE}
	EDGE_TYPES = []string{EDGE_BELONG, EDGE_SERVER, EDGE_CLIENT, EDGE_TCP, EDGE_UDP}
}