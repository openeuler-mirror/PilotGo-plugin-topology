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
		collector.Psutildata = collector.CreatePsutilCollector()
		err := collector.Psutildata.Collect_host_data()
		if err != nil {
			err = errors.Wrap(err, "**2")
			return nil, err
		}

		err = collector.Psutildata.Collect_netconnection_all_data()
		if err != nil {
			err = errors.Wrap(err, "**2")
			return nil, err
		}

		err = collector.Psutildata.Collect_interfaces_io_data()
		if err != nil {
			err = errors.Wrap(err, "**2")
			return nil, err
		}

		err = collector.Psutildata.Collect_process_instant_data()
		if err != nil {
			err = errors.Wrap(err, "**2")
			return nil, err
		}

		err = collector.Psutildata.Collect_addrInterfaceMap_data()
		if err != nil {
			err = errors.Wrap(err, "**2")
			return nil, err
		}

		err = collector.Psutildata.Collect_disk_data()
		if err != nil {
			err = errors.Wrap(err, "**2")
			return nil, err
		}

		err = collector.Psutildata.Collect_cpu_data()
		if err != nil {
			err = errors.Wrap(err, "**2")
			return nil, err
		}

		return collector.Psutildata, nil
	case "ebpf":

	}

	return nil, errors.New("wrong data source")
}
