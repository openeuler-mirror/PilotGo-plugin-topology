/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Fri Nov 8 09:13:05 2024 +0800
 */
package global

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

type Data_collector interface {
	Collect_process_instant_data() error
	Collect_host_data() error
	Collect_netconnection_all_data() error
	Collect_addrInterfaceMap_data() error
}

type Host struct {
	Hostname             string `json:"hostname"`
	Uptime               uint64 `json:"uptime"`
	BootTime             uint64 `json:"bootTime"`
	Procs                uint64 `json:"procs"`           // number of processes
	OS                   string `json:"os"`              // ex: freebsd, linux
	Platform             string `json:"platform"`        // ex: ubuntu, linuxmint
	PlatformFamily       string `json:"platformFamily"`  // ex: debian, rhel
	PlatformVersion      string `json:"platformVersion"` // version of the complete OS
	KernelVersion        string `json:"kernelVersion"`   // version of the OS kernel (if available)
	KernelArch           string `json:"kernelArch"`      // native cpu architecture queried at runtime, as returned by `uname -m` or empty string in case of error
	VirtualizationSystem string `json:"virtualizationSystem"`
	VirtualizationRole   string `json:"virtualizationRole"` // guest or host
	MachineUUID          string `json:"MachineUUID"`        // ex: pilotgo agent uuid
}

type Process struct {
	Pid     int32    `json:"pid"`
	Ppid    int32    `json:"ppid"`
	Cpid    []int32  `json:"cpid"`
	Tids    []int32  `json:"tid"`
	Threads []Thread `json:"threads"`
	Uids    []int32  `json:"uids"`
	Gids    []int32  `json:"gids"`

	Username   string `json:"username"`
	Status     string `json:"status"`
	CreateTime int64  `json:"createtime"`
	ExePath    string `json:"exepath"`
	ExeName    string `json:"exename"`
	Cmdline    string `json:"cmdline"`
	Cwd        string `json:"cwd"`

	Nice   int32 `json:"nice"`
	IOnice int32 `json:"ionice"`

	Connections   []Netconnection `json:"connections"`
	NetIOCounters []net.IOCountersStat `json:"netiocounters"`

	IOCounters process.IOCountersStat `json:"iocounters"`

	OpenFiles []process.OpenFilesStat `json:"openfiles"`
	NumFDs    int32                   `json:"numfds"`

	NumCtxSwitches process.NumCtxSwitchesStat `json:"numctxswitches"`
	PageFaults     process.PageFaultsStat     `json:"pagefaults"`
	MemoryInfo     process.MemoryInfoStat     `json:"memoryinfo"`
	CPUPercent     float64                    `json:"cpupercent"`
	MemoryPercent  float64                    `json:"memorypercent"`
}

type Thread struct {
	Tid       int32   `json:"tid"`
	Tgid      int32   `json:"tgid"`
	CPU       string  `json:"cpu"`
	User      float64 `json:"user"`
	System    float64 `json:"system"`
	Idle      float64 `json:"idle"`
	Nice      float64 `json:"nice"`
	Iowait    float64 `json:"iowait"`
	Irq       float64 `json:"irq"`
	Softirq   float64 `json:"softirq"`
	Steal     float64 `json:"steal"`
	Guest     float64 `json:"guest"`
	GuestNice float64 `json:"guestNice"`
}

type Netconnection struct {
	Fd     uint32  `json:"fd"`
	Family uint32  `json:"family"`
	Type   uint32  `json:"type"`
	Laddr  string  `json:"laddr"`
	Raddr  string  `json:"raddr"`
	Status string  `json:"status"`
	Uids   []int32 `json:"uids"`
	Pid    int32   `json:"pid"`
}

type NetIOcounter struct {
	Name        string `json:"name"`        
	BytesSent   uint64 `json:"bytesSent"` 
	BytesRecv   uint64 `json:"bytesRecv"`   
	PacketsSent uint64 `json:"packetsSent"` 
	PacketsRecv uint64 `json:"packetsRecv"` 
	Errin       uint64 `json:"errin"`      
	Errout      uint64 `json:"errout"`   
	Dropin      uint64 `json:"dropin"`   
	Dropout     uint64 `json:"dropout"`   
	Fifoin      uint64 `json:"fifoin"`     
	Fifoout     uint64 `json:"fifoout"`  
}

type Disk struct {
	Partition disk.PartitionStat  `json:"partition"`
	IOcounter disk.IOCountersStat `json:"iocounter"`
	Usage     disk.UsageStat      `json:"usage"`
}

type Cpu struct {
	Info cpu.InfoStat  `json:"info"`
	Time cpu.TimesStat `json:"time"`
}
