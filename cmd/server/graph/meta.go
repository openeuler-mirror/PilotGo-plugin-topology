/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package graph

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

type TopoDataBuffer struct {
	TopoConfId string
	Nodes      *Nodes
	Edges      *Edges
	Combos     []map[string]string
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

	Connections   []Netconnection      `json:"connections"`
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

func StructToMap(obj interface{}) map[string]string {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}

	if objValue.Kind() != reflect.Struct {
		return nil
	}

	objType := objValue.Type()
	fieldCount := objType.NumField()

	m := make(map[string]string)
	for i := 0; i < fieldCount; i++ {
		field := objType.Field(i)
		fieldValue := objValue.Field(i)

		switch fieldValue.Kind() {
		case reflect.String:
			m[field.Name] = fieldValue.Interface().(string)
		case reflect.Uint64:
			fieldvalue_uint64 := fieldValue.Interface().(uint64)
			m[field.Name] = strconv.Itoa(int(fieldvalue_uint64))
		case reflect.Float64:
			fieldvalue_float64 := fieldValue.Interface().(float64)
			m[field.Name] = strconv.FormatFloat(fieldvalue_float64, 'f', -1, 64)
		}
	}

	return m
}

func HostToMap(host *Host, a_i_map *map[string][]string) *map[string]string {
	host_metrics := StructToMap(host)

	interfaces_string := []string{}
	for key, value := range *a_i_map {
		interfaces_string = append(interfaces_string, key+":"+strings.Join(value, " "))
	}

	if _, ok := host_metrics["interfaces"]; ok {
		host_metrics["interfaces"] = strings.Join(interfaces_string, ";")
	}

	return &host_metrics
}

func ProcessToMap(process *Process) *map[string]string {
	uids_string := []string{}
	for _, u := range process.Uids {
		uids_string = append(uids_string, strconv.Itoa(int(u)))
	}

	gids_string := []string{}
	for _, g := range process.Gids {
		gids_string = append(gids_string, strconv.Itoa(int(g)))
	}

	openfiles_string := []string{}
	for _, of := range process.OpenFiles {
		openfiles_string = append(openfiles_string, strconv.Itoa(int(of.Fd))+":"+of.Path)
	}

	cpid_string := []string{}
	for _, cid := range process.Cpid {
		cpid_string = append(cpid_string, strconv.Itoa(int(cid)))
	}

	return &map[string]string{
		"Pid":                         strconv.Itoa(int(process.Pid)),
		"Ppid":                        strconv.Itoa(int(process.Ppid)),
		"Cpid":                        strings.Join(cpid_string, " "),
		"Uids":                        strings.Join(uids_string, " "),
		"Gids":                        strings.Join(gids_string, " "),
		"Status":                      process.Status,
		"CreateTime":                  strconv.Itoa(int(process.CreateTime)),
		"Cwd":                         process.Cwd,
		"ExePath":                     process.ExePath,
		"Cmdline":                     process.Cmdline,
		"Nice":                        strconv.Itoa(int(process.Nice)),
		"IOnice":                      strconv.Itoa(int(process.IOnice)),
		"DISK-rc":                     strconv.Itoa(int(process.IOCounters.ReadCount)),
		"DISK-rb":                     strconv.Itoa(int(process.IOCounters.ReadBytes)),
		"DISK-wc":                     strconv.Itoa(int(process.IOCounters.WriteCount)),
		"DISK-wb":                     strconv.Itoa(int(process.IOCounters.WriteBytes)),
		"fd":                          strings.Join(openfiles_string, " "),
		"NumCtxSwitches-v":            strconv.Itoa(int(process.NumCtxSwitches.Voluntary)),
		"NumCtxSwitches-inv":          strconv.Itoa(int(process.NumCtxSwitches.Involuntary)),
		"PageFaults-MinorFaults":      strconv.Itoa(int(process.PageFaults.MinorFaults)),
		"PageFaults-MajorFaults":      strconv.Itoa(int(process.PageFaults.MajorFaults)),
		"PageFaults-ChildMinorFaults": strconv.Itoa(int(process.PageFaults.ChildMinorFaults)),
		"PageFaults-ChildMajorFaults": strconv.Itoa(int(process.PageFaults.ChildMajorFaults)),
		"CPUPercent":                  strconv.FormatFloat(process.CPUPercent, 'f', -1, 64),
		"MemoryPercent":               strconv.FormatFloat(process.MemoryPercent, 'f', -1, 64),
		"MemoryInfo":                  process.MemoryInfo.String(),
	}
}

func ThreadToMap(thread *Thread) *map[string]string {
	return &map[string]string{
		"Tid":       strconv.Itoa(int(thread.Tid)),
		"Tgid":      strconv.Itoa(int(thread.Tgid)),
		"CPU":       thread.CPU,
		"User":      strconv.FormatFloat(thread.User, 'f', -1, 64),
		"System":    strconv.FormatFloat(thread.System, 'f', -1, 64),
		"Idle":      strconv.FormatFloat(thread.Idle, 'f', -1, 64),
		"Nice":      strconv.FormatFloat(thread.Nice, 'f', -1, 64),
		"Iowait":    strconv.FormatFloat(thread.Iowait, 'f', -1, 64),
		"Irq":       strconv.FormatFloat(thread.Irq, 'f', -1, 64),
		"Softirq":   strconv.FormatFloat(thread.Softirq, 'f', -1, 64),
		"Steal":     strconv.FormatFloat(thread.Steal, 'f', -1, 64),
		"Guest":     strconv.FormatFloat(thread.Guest, 'f', -1, 64),
		"GuestNice": strconv.FormatFloat(thread.GuestNice, 'f', -1, 64),
	}
}

// net节点的metrics字段 临时定义
func NetToMap(net *Netconnection) *map[string]string {
	uids_string := []string{}
	for _, uid := range net.Uids {
		uids_string = append(uids_string, strconv.Itoa(int(uid)))
	}

	return &map[string]string{
		"Fd":     strconv.Itoa(int(net.Fd)),
		"Family": strconv.Itoa(int(net.Family)),
		"Type":   strconv.Itoa(int(net.Type)),
		"Laddr":  net.Laddr,
		"Raddr":  net.Raddr,
		"Status": net.Status,
		"Uids":   strings.Join(uids_string, " "),
		"Pid":    strconv.Itoa(int(net.Pid)),
	}
}

// func NetToMap(net *net.IOCountersStat, a_i_map *map[string][]string) *map[string]string {
// 	addrs := []string{}
// 	for key, value := range *a_i_map {
// 		if net.Name == key {
// 			addrs = value
// 		}
// 	}

// 	return &map[string]string{
// 		"Name":        net.Name,
// 		"addrs":       addrs[0],
// 		"BytesSent":   strconv.Itoa(int(net.BytesSent)),
// 		"BytesRecv":   strconv.Itoa(int(net.BytesRecv)),
// 		"PacketsSent": strconv.Itoa(int(net.PacketsSent)),
// 		"PacketsRecv": strconv.Itoa(int(net.PacketsRecv)),
// 		"Errin":       strconv.Itoa(int(net.Errin)),
// 		"Errout":      strconv.Itoa(int(net.Errout)),
// 		"Dropin":      strconv.Itoa(int(net.Dropin)),
// 		"Dropout":     strconv.Itoa(int(net.Dropout)),
// 		"Fifoin":      strconv.Itoa(int(net.Fifoin)),
// 		"Fifoout":     strconv.Itoa(int(net.Fifoout)),
// 	}
// }

func DiskToMap(disk *Disk) *map[string]string {
	disk_map := make(map[string]string)
	partition_map := StructToMap(disk.Partition)
	iocounter_map := StructToMap(disk.IOcounter)
	usage_map := StructToMap(disk.Usage)

	for k, v := range partition_map {
		disk_map[k] = v
	}

	for k, v := range iocounter_map {
		if k != "Name" {
			disk_map[k] = v
		}
	}

	for k, v := range usage_map {
		if k != "Path" && k != "Fstype" {
			disk_map[k] = v
		}
	}

	return &disk_map
}

func CpuToMap(cpu *Cpu) *map[string]string {
	cpu_map := make(map[string]string)
	info_map := StructToMap(cpu.Info)
	time_map := StructToMap(cpu.Time)

	for k, v := range info_map {
		if k != "Flags" {
			cpu_map[k] = v
		}
	}

	for k, v := range time_map {
		if k != "CPU" {
			cpu_map[k] = v
		}
	}

	return &cpu_map
}

func InterfaceToMap(iface *NetIOcounter) *map[string]string {
	iface_map := make(map[string]string)
	old_map := StructToMap(iface)

	for k, v := range old_map {
		if k != "Name" {
			iface_map[k] = v
		}
	}

	return &iface_map
}
