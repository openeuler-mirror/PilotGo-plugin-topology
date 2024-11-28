/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
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
