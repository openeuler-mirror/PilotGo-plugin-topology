package global

import (
	"context"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/resourcemanage"
	"github.com/go-redis/redis/v8"
	influx "github.com/influxdata/influxdb-client-go/v2"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var (
	NODE_TYPES   []string
	EDGE_TYPES   []string
	DEFAULT_TAGS []string
)

var (
	RootContext = context.Background()
)

var ERManager *resourcemanage.ErrorReleaseManagement

var (
	Global_graph_database string

	Global_neo4j_driver  neo4j.Driver
	Global_redis_client  *redis.Client
	Global_influx_client influx.Client
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
	NODE_CONNECTOR = "_._"
	EDGE_CONNECTOR = "_._._"
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
	EVENT_TYPE_0  = "kernel"
	EVENT_TYPE_1  = "user_level"
	EVENT_TYPE_2  = "mail"
	EVENT_TYPE_3  = "system_daemons"
	EVENT_TYPE_4  = "security_authorizaion"
	EVENT_TYPE_5  = "by_log_service"
	EVENT_TYPE_7  = "network_news"
	EVENT_TYPE_9  = "clock_daemon"
	EVENT_TYPE_11 = "ftp_daemon"
	EVENT_TYPE_12 = "ntp_daemon"
	EVENT_TYPE_13 = "log_audit"
	EVENT_TYPE_14 = "log_alert"
	EVENT_TYPE_16 = "reserve"
	EVENT_TYPE_17 = "reserve"
	EVENT_TYPE_18 = "reserve"
)

// 事件级别
const (
	EVENT_LEVEL_0 = "emergency"
	EVENT_LEVEL_1 = "alert"
	EVENT_LEVEL_2 = "critical"
	EVENT_LEVEL_3 = "error"
	EVENT_LEVEL_4 = "warn"
	EVENT_LEVEL_5 = "notice"
	EVENT_LEVEL_6 = "info"
	EVENT_LEVEL_7 = "debug"
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
