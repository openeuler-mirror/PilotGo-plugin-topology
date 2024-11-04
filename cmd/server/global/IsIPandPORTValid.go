package global

import (
	"fmt"
	"net"
	"time"
)

const (
	req_timeout = 1000 * time.Millisecond
)

// 检测IP是否可达
func IsIPandPORTValid(ip, port string) (bool, error) {
	addr, err := net.ResolveIPAddr("ip", ip)
	if err != nil {
		return false, err
	}

	// 设置连接超时时间
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", addr.String(), port), req_timeout)
	if err != nil {
		return false, err
	}

	conn.Close()
	return true, nil
}
