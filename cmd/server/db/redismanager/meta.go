package redismanager

import "time"

type AgentHeartbeat struct {
	UUID              string
	Addr              string
	HeartbeatInterval int
	Time              time.Time
}