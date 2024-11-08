package service

import (
	"fmt"
	"sync"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/collector"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/agent/global"
	"github.com/pkg/errors"
)

func DataCollectorService() (global.Data_collector, error) {
	datasource := conf.Config().Topo.Datasource
	cost_time := []string{}
	var wg sync.WaitGroup
	type err_and_time struct {
		Err  error
		Time string
	}
	ch := make(chan *err_and_time, 6)

	switch datasource {
	case "gopsutil":
		psutildata := collector.CreatePsutilCollector()
		err := psutildata.Collect_host_data()
		if err != nil {
			err = errors.Wrap(err, " ")
			return nil, err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			time_start := time.Now()
			err := psutildata.Collect_process_instant_data()
			time_elapse := time.Since(time_start)
			ch <- &err_and_time{
				Err:  err,
				Time: fmt.Sprintf("process 耗时：%v", time_elapse),
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			time_start := time.Now()
			err := psutildata.Collect_netconnection_all_data()
			time_elapse := time.Since(time_start)
			ch <- &err_and_time{
				Err:  err,
				Time: fmt.Sprintf("netconnection 耗时：%v", time_elapse),
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			time_start := time.Now()
			err = psutildata.Collect_interfaces_io_data()
			time_elapse := time.Since(time_start)
			ch <- &err_and_time{
				Err:  err,
				Time: fmt.Sprintf("interfaces io 耗时：%v", time_elapse),
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			time_start := time.Now()
			err = psutildata.Collect_addrInterfaceMap_data()
			time_elapse := time.Since(time_start)
			ch <- &err_and_time{
				Err:  err,
				Time: fmt.Sprintf("addrinterfacemap 耗时：%v", time_elapse),
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			time_start := time.Now()
			err = psutildata.Collect_disk_data()
			time_elapse := time.Since(time_start)
			ch <- &err_and_time{
				Err:  err,
				Time: fmt.Sprintf("disk 耗时：%v", time_elapse),
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			time_start := time.Now()
			err = psutildata.Collect_cpu_data()
			time_elapse := time.Since(time_start)
			ch <- &err_and_time{
				Err:  err,
				Time: fmt.Sprintf("cpu 耗时：%v", time_elapse),
			}
		}()

		for data := range ch {
			if data.Err != nil {
				err = errors.Wrap(err, " ")
				return nil, err
			}
			cost_time = append(cost_time, data.Time)
			if len(cost_time) == 6 {
				close(ch)
				break
			}
		}

		wg.Wait()

		global.ERManager.ErrorTransmit("service", "debug", errors.New("==========collect=========="), false, false)
		for _, t := range cost_time {
			global.ERManager.ErrorTransmit("service", "debug", errors.New(t), false, false)
		}

		return psutildata, nil
	default:
		return nil, errors.New("wrong data source")
	}
}
