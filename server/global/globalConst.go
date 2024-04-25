package global

var (
	NODE_TYPES   []string
	EDGE_TYPES   []string
	DEFAULT_TAGS []string
)

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

const (
	NODE_CONNECTOR = "_"
	EDGE_CONNECTOR = "__"
	STR_CONNECTOR  = "___"
)

const (
	INNER_LAYOUT_1 = "5"
	INNER_LAYOUT_2 = "4"
	INNER_LAYOUT_3 = "2"
	INNER_LAYOUT_4 = "3"
	INNER_LAYOUT_5 = "1"
)

// 事件类型
const (
	EVENT_TYPE_SYSTEM        = "system"
	EVENT_TYPE_SECURITY      = "security"
	EVENT_TYPE_PERFORMANCE   = "performance"
	EVENT_TYPE_CONFIGURATION = "configuration"
	EVENT_TYPE_LOG           = "log"
)

// 事件级别
const (
	EVENT_LEVEL_FATAL = "fatal"
	EVENT_LEVEL_ERROR = "error"
	EVENT_LEVEL_WARN  = "warn"
	EVENT_LEVEL_INFO  = "info"
)

func init() {
	NODE_TYPES = []string{NODE_HOST, NODE_PROCESS, NODE_THREAD, NODE_NET, NODE_APP, NODE_RESOURCE}
	EDGE_TYPES = []string{EDGE_BELONG, EDGE_SERVER, EDGE_CLIENT, EDGE_TCP, EDGE_UDP}

	DEFAULT_TAGS = []string{
		"redis-server",
		"mysqld",
		"arangodb",
		"elasticsearch",
		"neo4j",
		"prometheus",
		"node-exporter",
		"grafana-server",
		"nginx-server",
		"kafka",
	}
}
