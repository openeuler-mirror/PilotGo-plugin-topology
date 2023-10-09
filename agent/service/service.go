package service

import (
	"gitee.com/openeuler/PilotGo-plugin-topology-agent/collector"
	"gitee.com/openeuler/PilotGo-plugin-topology-agent/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-agent/utils"
	"github.com/pkg/errors"
)

func DataCollectorService() (utils.Data_collector, error) {
	datasource := conf.Config().Topo.Datasource
	switch datasource {
	case "gopsutil":
		gops := collector.CreatePsutilCollector()
		err := gops.Collect_host_data()
		if err != nil {
			err = errors.Wrap(err, "**2")
			return nil, err
		}

		err = gops.Collect_netconnection_all_data()
		if err != nil {
			err = errors.Wrap(err, "**2")
			return nil, err
		}

		err = gops.Collect_interfaces_io_data()
		if err != nil {
			err = errors.Wrap(err, "**2")
			return nil, err
		}

		err = gops.Collect_process_instant_data()
		if err != nil {
			err = errors.Wrap(err, "**2")
			return nil, err
		}

		err = gops.Collect_addrInterfaceMap_data()
		if err != nil {
			err = errors.Wrap(err, "**2")
			return nil, err
		}

		err = gops.Collect_disk_data()
		if err != nil {
			err = errors.Wrap(err, "**2")
			return nil, err
		}

		err = gops.Collect_cpu_data()
		if err != nil {
			err = errors.Wrap(err, "**2")
			return nil, err
		}

		return gops, nil
	case "ebpf":

	}

	return nil, errors.New("wrong data source")
}
