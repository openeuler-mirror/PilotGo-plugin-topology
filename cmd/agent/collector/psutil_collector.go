/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package collector

import (
	"encoding/json"
	"strconv"
	"sync"

	"github.com/pkg/errors"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/global"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/process"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/net"
)

type PsutilCollector struct {
	Host_1             *global.Host
	Processes_1        []*global.Process
	Netconnections_1   []*global.Netconnection
	NetIOcounters_1    []*global.NetIOcounter
	AddrInterfaceMap_1 map[string][]string
	Disks_1            []*global.Disk
	Cpus_1             []*global.Cpu
}

func CreatePsutilCollector() *PsutilCollector {
	return &PsutilCollector{}
}

func (pc *PsutilCollector) Collect_host_data() error {
	hostinit, err := host.Info()
	if err != nil {
		return errors.New(err.Error())
	}

	m_u_bytes, err := global.FileReadBytes(global.Agentuuid_filepath)
	if err != nil {
		return errors.Wrap(err, " ")
	}
	type machineuuid struct {
		Agentuuid string `json:"agent_uuid"`
	}
	m_u_struct := &machineuuid{}
	json.Unmarshal(m_u_bytes, m_u_struct)

	pc.Host_1 = &global.Host{
		Hostname:             hostinit.Hostname,
		Uptime:               hostinit.Uptime,
		BootTime:             hostinit.BootTime,
		Procs:                hostinit.Procs,
		OS:                   hostinit.OS,
		Platform:             hostinit.Platform,
		PlatformFamily:       hostinit.PlatformFamily,
		PlatformVersion:      hostinit.PlatformVersion,
		KernelVersion:        hostinit.KernelVersion,
		KernelArch:           hostinit.KernelArch,
		VirtualizationSystem: hostinit.VirtualizationSystem,
		VirtualizationRole:   hostinit.VirtualizationRole,
		MachineUUID:          m_u_struct.Agentuuid,
	}
	return nil
}

func (pc *PsutilCollector) Collect_process_instant_data() error {
	Echo_process_err := func(method string, err error, processid int32) {
		if err != nil {
			// _, filepath, line, _ := runtime.Caller(1)
			// fmt.Printf("file: %s, line: %d, func: %s, processid: %d, err: %s\n", filepath, line-1, method, processid, err.Error())
			// err = errors.Errorf("%s: %s, %d", err.Error(), method, processid)

			// fmt.Printf("%v: %s, %d\n", err.Error(), method, processid)
			return
		}
	}

	var wg sync.WaitGroup
	var lock *sync.Mutex = new(sync.Mutex)

	processes_0, err := process.Processes()
	if err != nil {
		return errors.Errorf("failed to get processes: %s", err)
	}

	for _, p0 := range processes_0 {
		wg.Add(1)
		go func(_p0 *process.Process, _lock *sync.Mutex) {
			defer wg.Done()
			p1 := &global.Process{}

			p1.Pid = _p0.Pid

			p1.Ppid, err = _p0.Ppid()
			Echo_process_err("ppid", err, _p0.Pid)

			children, err := _p0.Children()
			Echo_process_err("children", err, _p0.Pid)
			if len(children) != 0 {
				for _, c := range children {
					p1.Cpid = append(p1.Cpid, c.Pid)
				}
			}

			thread, err := _p0.Threads()
			Echo_process_err("threads", err, _p0.Pid)
			if len(thread) != 0 {
				tgid, err := _p0.Tgid()
				Echo_process_err("tgid", err, _p0.Pid)

				for k, v := range thread {
					p1.Tids = append(p1.Tids, k)
					t := &global.Thread{
						Tid:       k,
						Tgid:      tgid,
						CPU:       v.CPU,
						User:      v.User,
						System:    v.System,
						Idle:      v.Idle,
						Nice:      v.Nice,
						Iowait:    v.Iowait,
						Irq:       v.Irq,
						Softirq:   v.Softirq,
						Steal:     v.Steal,
						Guest:     v.Guest,
						GuestNice: v.GuestNice,
					}
					p1.Threads = append(p1.Threads, *t)
				}
			}

			p1.Uids, err = _p0.Uids()
			Echo_process_err("uids", err, _p0.Pid)

			p1.Gids, err = _p0.Gids()
			Echo_process_err("gids", err, _p0.Pid)

			p1.Username, err = _p0.Username()
			Echo_process_err("username", err, _p0.Pid)

			p1.Status, err = _p0.Status()
			Echo_process_err("status", err, _p0.Pid)

			p1.CreateTime, err = _p0.CreateTime()
			Echo_process_err("createtime", err, _p0.Pid)

			p1.ExePath, err = _p0.Exe()
			Echo_process_err("exe", err, _p0.Pid)

			p1.ExeName, err = _p0.Name()
			Echo_process_err("name", err, _p0.Pid)

			p1.Cmdline, err = _p0.Cmdline()
			Echo_process_err("cmdline", err, _p0.Pid)

			p1.Cwd, err = _p0.Cwd()
			Echo_process_err("cwd", err, _p0.Pid)

			p1.Nice, err = _p0.Nice()
			Echo_process_err("nice", err, _p0.Pid)

			p1.IOnice, err = _p0.IOnice()
			Echo_process_err("ionice", err, _p0.Pid)

			connections, err := _p0.Connections()
			Echo_process_err("connections", err, _p0.Pid)
			p1.Connections = global.GopsutilNetMeta2TopoNetMeta(connections)

			p1.NetIOCounters, err = _p0.NetIOCounters(true)
			Echo_process_err("netiocounters", err, _p0.Pid)

			iocounters, err := _p0.IOCounters()
			Echo_process_err("iocounters", err, _p0.Pid)
			if iocounters != nil {
				p1.IOCounters = *iocounters
			}

			p1.OpenFiles, err = _p0.OpenFiles()
			Echo_process_err("openfiles", err, _p0.Pid)

			p1.NumFDs, err = _p0.NumFDs()
			Echo_process_err("numfds", err, _p0.Pid)

			numctxswitches, err := _p0.NumCtxSwitches()
			Echo_process_err("numctxswitches", err, _p0.Pid)
			if numctxswitches != nil {
				p1.NumCtxSwitches = *numctxswitches
			}

			pagefaults, err := _p0.PageFaults()
			Echo_process_err("pagefaults", err, _p0.Pid)
			if pagefaults != nil {
				p1.PageFaults = *pagefaults
			}

			memoryinfo, err := _p0.MemoryInfo()
			Echo_process_err("memoryinfo", err, _p0.Pid)
			if memoryinfo != nil {
				p1.MemoryInfo = *memoryinfo
			}

			p1.CPUPercent, err = _p0.CPUPercent()
			Echo_process_err("cpupercent", err, _p0.Pid)

			memorypercent, err := _p0.MemoryPercent()
			Echo_process_err("memorypercent", err, _p0.Pid)
			p1.MemoryPercent = float64(memorypercent)

			_lock.Lock()
			pc.Processes_1 = append(pc.Processes_1, p1)
			_lock.Unlock()
		}(p0, lock)
	}

	wg.Wait()

	return nil
}

func (pc *PsutilCollector) Collect_netconnection_all_data() error {
	connections, err := net.Connections("all")
	if err != nil {
		return errors.Errorf("failed to run net.connections: %s", err)
	}

	for _, c := range connections {
		c1 := &global.Netconnection{}
		// if c.Status == "NONE" {
		// 	continue
		// }

		if c.Laddr.Port == 22 || c.Raddr.Port == 22 {
			continue
		}

		if c.Status != "ESTABLISHED" {
			continue
		}

		c1.Fd = c.Fd
		c1.Family = c.Family
		c1.Type = c.Type
		c1.Laddr = c.Laddr.IP + ":" + strconv.Itoa(int(c.Laddr.Port))
		c1.Raddr = c.Raddr.IP + ":" + strconv.Itoa(int(c.Raddr.Port))
		c1.Status = c.Status
		c1.Uids = c.Uids
		c1.Pid = c.Pid
		pc.Netconnections_1 = append(pc.Netconnections_1, c1)
	}
	return nil
}

func (pc *PsutilCollector) Collect_addrInterfaceMap_data() error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return errors.Errorf("failed to run net.interfaces: %s", err)
	}

	addrinterfacemap := map[string][]string{}
	for _, iface := range interfaces {
		for _, addr := range iface.Addrs {
			addrinterfacemap[iface.Name] = append(addrinterfacemap[iface.Name], addr.Addr)
		}
	}

	pc.AddrInterfaceMap_1 = addrinterfacemap
	return nil
}

func (pc *PsutilCollector) Collect_interfaces_io_data() error {
	iocounters, err := net.IOCounters(true)
	if err != nil {
		return errors.Errorf("failed to collect interfaces io: %s", err.Error())
	}

	for _, iocounter := range iocounters {
		interfaceIO := &global.NetIOcounter{}

		interfaceIO.Name = iocounter.Name
		interfaceIO.BytesRecv = iocounter.BytesRecv
		interfaceIO.BytesSent = iocounter.BytesSent
		interfaceIO.Dropin = iocounter.Dropin
		interfaceIO.Dropout = iocounter.Dropout
		interfaceIO.Errin = iocounter.Errin
		interfaceIO.Errout = iocounter.Errout
		interfaceIO.Fifoin = iocounter.Fifoin
		interfaceIO.Fifoout = iocounter.Fifoout
		interfaceIO.PacketsRecv = iocounter.PacketsRecv
		interfaceIO.PacketsSent = iocounter.PacketsSent

		pc.NetIOcounters_1 = append(pc.NetIOcounters_1, interfaceIO)
	}
	return nil
}

func (pc *PsutilCollector) Collect_disk_data() error {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return errors.Errorf("failed to collect disk partitions: %s", err.Error())
	}

	for _, partition := range partitions {
		disk_entity := &global.Disk{}
		disk_entity.Partition = partition

		iocounter, err := disk.IOCounters([]string{disk_entity.Partition.Device}...)
		if err != nil {
			return errors.Errorf("failed to collect disk io: %s", err.Error())
		}

		disk_entity.IOcounter = iocounter[partition.Device]

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			return errors.Errorf("failed to collect disk usage: %s", err.Error())
		}

		disk_entity.Usage = *usage

		pc.Disks_1 = append(pc.Disks_1, disk_entity)
	}
	return nil
}

func (pc *PsutilCollector) Collect_cpu_data() error {
	cputimes, err := cpu.Times(true)
	if err != nil {
		return errors.Errorf("failed to collect cpu times: %s", err.Error())
	}

	for i, cputime := range cputimes {
		cpu_entity := &global.Cpu{}
		cpu_entity.Time = cputime

		cpuinfos, err := cpu.Info()
		if err != nil {
			return errors.Errorf("failed to collect cpu info: %s", err.Error())
		}
		cpu_entity.Info = cpuinfos[i]

		pc.Cpus_1 = append(pc.Cpus_1, cpu_entity)
	}
	return nil
}
