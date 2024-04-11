package promemanager

import (
	"context"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

var Global_Prome *Prometheus

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
		err = errors.Errorf("failed to create prometheus client: %s **errstack**2", err.Error())
		return err
	}

	promapi := v1.NewAPI(client)
	p.Api = promapi

	return nil
}

func (p *Prometheus) GetTargets() ([]map[string]string, error) {
	targets := make([]map[string]string, 0)

	result, err := p.Api.Targets(p.Ctx)
	if err != nil {
		err = errors.Errorf("failed to get prometheus targets: %s **errstack**2", err.Error())
		return nil, err
	}

	for _, t := range result.Active {
		targets = append(targets, map[string]string{
			"instance": string(t.Labels["instance"]),
			"job":      string(t.Labels["job"]),
		})
	}

	return targets, nil
}

func (p *Prometheus) GetMetrics() ([]string, error) {
	metrics := make([]string, 0)

	result, err := p.Api.Metadata(p.Ctx, "", "")
	if err != nil {
		err = errors.Errorf("failed to get prometheus metric list: %s **errstack**2", err.Error())
		return nil, err
	}

	for m := range result {
		metrics = append(metrics, m)
	}

	return metrics, nil
}
