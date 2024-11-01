package influxmanager

import (
	"context"
	"fmt"
	"time"

	"gitee.com/openeuler/PilotGo-plugin-topology/server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology/server/global"
	influx "github.com/influxdata/influxdb-client-go/v2"
	"github.com/pkg/errors"
)

var Global_Influx *InfluxClient

type InfluxClient struct {
	ServerURL    string
	Token        string
	Organization string
	Bucket       string

	Client influx.Client
}

func InfluxdbInit(conf *conf.InfluxConf) *InfluxClient {
	i := &InfluxClient{
		ServerURL:    conf.Addr,
		Token:        conf.Token,
		Organization: conf.Org,
		Bucket:       conf.Bucket,
	}

	global.Global_influx_client = influx.NewClient(i.ServerURL, i.Token)
	i.Client = global.Global_influx_client
	return i
}

func (i *InfluxClient) WriteWithLineProtocol(measurement string, tags map[string]string, fields map[string]interface{}) error {
	if measurement == "" || len(fields) == 0 {
		err := errors.Errorf("write to influxdb failed: measurement(%s), tags(%+v), fields(%+v)", measurement, tags, fields)
		return err
	}

	writeAPI := i.Client.WriteAPI(i.Organization, i.Bucket)
	p := influx.NewPoint(measurement, tags, fields, time.Now())
	writeAPI.WritePoint(p)
	writeAPI.Flush()

	return nil
}

func (i *InfluxClient) Query(measurement, start, end string) error {

	query := fmt.Sprintf("from(bucket:%s)|> range(start: -1h) |> filter(fn: (r) => r._measurement == %s)", i.Bucket, measurement)

	queryAPI := i.Client.QueryAPI(i.Organization)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		err = errors.Errorf(err.Error())
		return err
	}

	for result.Next() {
		if result.TableChanged() {
			fmt.Printf("table: %s\n", result.TableMetadata().String())
		}

		fmt.Printf("time: %v, field: %v, value: %v\n", result.Record().Time().Format("2006-01-02 15:04:05"), result.Record().Field(), result.Record().Value())

	}

	if result.Err() != nil {
		fmt.Printf("query parsing error: %s\n", result.Err().Error())
	}

	return nil
}
