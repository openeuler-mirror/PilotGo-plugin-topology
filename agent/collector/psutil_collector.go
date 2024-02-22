package collector

import (
	"encoding/json"
	"fmt"
	"strconv"

	"gitee.com/openeuler/PilotGo-plugin-topology-agent/utils"
	"gitee.com/openeuler/PilotGo/sdk/logger"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/process"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/net"
)

var Psutildata *PsutilCollector

type PsutilCollector struct {
	Host_1             *utils.Host
	Processes_1        []*utils.Process
	Netconnections_1   []*utils.Netconnection
	NetIOcounters_1    []*utils.NetIOcounter
	AddrInterfaceMap_1 map[string][]string
	Disks_1            []*utils.Disk
	Cpus_1             []*utils.Cpu
}

func CreatePsutilCollector() *PsutilCollector {
	return &PsutilCollector{}
}

func (pc *PsutilCollector) Collect_host_data() error {
	hostinit, err := host.Info()
	if err != nil {
		// err = errors.New(err.Error())
		logger.Error(err.Error())
		return err
	}

	m_u_bytes, err := utils.FileReadBytes(utils.Agentuuid_filepath)
	if err != nil {
		// err = errors.New(err.Error())
		logger.Error(err.Error())
		return err
	}
	type machineuuid struct {
		Agentuuid string `json:"agent_uuid"`
	}
	m_u_struct := &machineuuid{}
	json.Unmarshal(m_u_bytes, m_u_struct)

	pc.Host_1 = &utils.Host{
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
			fmt.Printf("%v: %s, %d\n", err.Error(), method, processid)
		}
	}

	processes_0, err := process.Processes()
	if err != nil {
		// err = errors.Errorf("failed to get processes: %s", err)
		logger.Error("failed to get processes: %s", err)
		return err
	}

	for _, p0 := range processes_0 {
		p1 := &utils.Process{}

		p1.Pid = p0.Pid

		p1.Ppid, err = p0.Ppid()
		Echo_process_err("ppid", err, p0.Pid)

		children, err := p0.Children()
		Echo_process_err("children", err, p0.Pid)
		if len(children) != 0 {
			for _, c := range children {
				p1.Cpid = append(p1.Cpid, c.Pid)
			}
		}

		thread, err := p0.Threads()
		Echo_process_err("threads", err, p0.Pid)
		if len(thread) != 0 {
			tgid, err := p0.Tgid()
			Echo_process_err("tgid", err, p0.Pid)

			for k, v := range thread {
				p1.Tids = append(p1.Tids, k)
				t := &utils.Thread{
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

		p1.Uids, err = p0.Uids()
		Echo_process_err("uids", err, p0.Pid)

		p1.Gids, err = p0.Gids()
		Echo_process_err("gids", err, p0.Pid)

		p1.Username, err = p0.Username()
		Echo_process_err("username", err, p0.Pid)

		p1.Status, err = p0.Status()
		Echo_process_err("status", err, p0.Pid)

		p1.CreateTime, err = p0.CreateTime()
		Echo_process_err("createtime", err, p0.Pid)

		p1.ExePath, err = p0.Exe()
		Echo_process_err("exe", err, p0.Pid)

		p1.ExeName, err = p0.Name()
		Echo_process_err("name", err, p0.Pid)

		p1.Cmdline, err = p0.Cmdline()
		Echo_process_err("cmdline", err, p0.Pid)

		p1.Cwd, err = p0.Cwd()
		Echo_process_err("cwd", err, p0.Pid)

		p1.Nice, err = p0.Nice()
		Echo_process_err("nice", err, p0.Pid)

		p1.IOnice, err = p0.IOnice()
		Echo_process_err("ionice", err, p0.Pid)

		connections, err := p0.Connections()
		Echo_process_err("connections", err, p0.Pid)
		for _, c := range connections {
			if c.Status == "NONE" {
				continue
			}
			p1.Connections = append(p1.Connections, c)
		}

		p1.NetIOCounters, err = p0.NetIOCounters(true)
		Echo_process_err("netiocounters", err, p0.Pid)

		iocounters, err := p0.IOCounters()
		Echo_process_err("iocounters", err, p0.Pid)
		p1.IOCounters = *iocounters

		p1.OpenFiles, err = p0.OpenFiles()
		Echo_process_err("openfiles", err, p0.Pid)

		p1.NumFDs, err = p0.NumFDs()
		Echo_process_err("numfds", err, p0.Pid)

		numctxswitches, err := p0.NumCtxSwitches()
		Echo_process_err("numctxswitches", err, p0.Pid)
		p1.NumCtxSwitches = *numctxswitches

		pagefaults, err := p0.PageFaults()
		Echo_process_err("pagefaults", err, p0.Pid)
		p1.PageFaults = *pagefaults

		memoryinfo, err := p0.MemoryInfo()
		Echo_process_err("memoryinfo", err, p0.Pid)
		p1.MemoryInfo = *memoryinfo

		p1.CPUPercent, err = p0.CPUPercent()
		Echo_process_err("cpupercent", err, p0.Pid)

		memorypercent, err := p0.MemoryPercent()
		Echo_process_err("memorypercent", err, p0.Pid)
		p1.MemoryPercent = float64(memorypercent)

		pc.Processes_1 = append(pc.Processes_1, p1)
	}

	return nil
}

func (pc *PsutilCollector) Collect_netconnection_all_data() error {
	connections, err := net.Connections("all")
	if err != nil {
		// err = errors.Errorf("failed to run net.connections: %s", err)
		logger.Error("failed to run net.connections: %s", err)
		return err
	}

	for _, c := range connections {
		c1 := &utils.Netconnection{}
		if c.Status == "NONE" {
			continue
		}

		if c.Laddr.Port == 22 || c.Raddr.Port == 22 {
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
		// err = errors.Errorf("failed to run net.interfaces: %s", err)
		logger.Error("failed to run net.interfaces: %s", err)
		return err
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
		// err = errors.Errorf("failed to collect interfaces io: %s", err.Error())
		logger.Error("failed to collect interfaces io: %s", err.Error())
		return err
	}

	for _, iocounter := range iocounters {
		interfaceIO := &utils.NetIOcounter{}

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
		// err = errors.Errorf("failed to collect disk partitions: %s", err.Error())
		logger.Error("failed to collect disk partitions: %s", err.Error())
		return err
	}

	for _, partition := range partitions {
		disk_entity := &utils.Disk{}
		disk_entity.Partition = partition

		iocounter, err := disk.IOCounters([]string{disk_entity.Partition.Device}...)
		if err != nil {
			// err = errors.Errorf("failed to collect disk io: %s", err.Error())
			logger.Error("failed to collect disk io: %s", err.Error())
			return err
		}

		disk_entity.IOcounter = iocounter[partition.Device]

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			// err = errors.Errorf("failed to collect disk usage: %s", err.Error())
			logger.Error("failed to collect disk usage: %s", err.Error())
			return err
		}

		disk_entity.Usage = *usage

		pc.Disks_1 = append(pc.Disks_1, disk_entity)
	}

	return nil
}

func (pc *PsutilCollector) Collect_cpu_data() error {
	cputimes, err := cpu.Times(true)
	if err != nil {
		// err = errors.Errorf("failed to collect cpu times: %s", err.Error())
		logger.Error("failed to collect cpu times: %s", err.Error())
		return err
	}

	for i, cputime := range cputimes {
		cpu_entity := &utils.Cpu{}
		cpu_entity.Time = cputime

		cpuinfos, err := cpu.Info()
		if err != nil {
			// err = errors.Errorf("failed to collect cpu info: %s", err.Error())
			logger.Error("failed to collect cpu info: %s", err.Error())
			return err
		}
		cpu_entity.Info = cpuinfos[i]

		pc.Cpus_1 = append(pc.Cpus_1, cpu_entity)
	}

	return nil
}
