package dao

import (
	"context"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

var Prome *Prometheus

type Prometheus struct {
	addr string
	Api  v1.API
	Ctx  context.Context
}

func CreatePrometheus(url string) *Prometheus {
	return &Prometheus{
		addr: url,
		Ctx:  context.Background(),
	}
}

func (p *Prometheus) CreateAPI() error {
	client, err := api.NewClient(api.Config{Address: p.addr})
	if err != nil {
		err = errors.Errorf("failed to create prometheus client: %s **2", err.Error())
		return err
	}

	promapi := v1.NewAPI(client)
	p.Api = promapi

	return nil
}

func (p *Prometheus) GetMetricList() ([]string, error) {
	metrics := make([]string, 0)

	result, err := p.Api.Metadata(p.Ctx, "", "")
	if err != nil {
		err = errors.Errorf("failed to get prometheus metric list: %s **2", err.Error())
		return nil, err
	}

	for m := range result {
		metrics = append(metrics, m)
	}

	return metrics, nil
}
