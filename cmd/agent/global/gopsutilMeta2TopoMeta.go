/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Fri Nov 8 09:13:05 2024 +0800
 */
package global

import (
	"strconv"

	"github.com/shirou/gopsutil/v4/net"
)

func GopsutilNetMeta2TopoNetMeta(gopsnets []net.ConnectionStat) []Netconnection {
	toponets := []Netconnection{}

	for _, c := range gopsnets {
		if c.Status == "NONE" {
			continue
		}
		if c.Laddr.Port == 22 || c.Raddr.Port == 22 {
			continue
		}
		c1 := &Netconnection{}
		c1.Fd = c.Fd
		c1.Family = c.Family
		c1.Type = c.Type
		c1.Laddr = c.Laddr.IP + ":" + strconv.Itoa(int(c.Laddr.Port))
		c1.Raddr = c.Raddr.IP + ":" + strconv.Itoa(int(c.Raddr.Port))
		c1.Status = c.Status
		c1.Uids = c.Uids
		c1.Pid = c.Pid
		toponets = append(toponets, *c1)
	}

	return toponets
}
