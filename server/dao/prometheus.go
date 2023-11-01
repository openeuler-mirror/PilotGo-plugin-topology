package dao

import (
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/pkg/errors"
)

var Prome *Prometheus

type Prometheus struct {
	addr string
	Api  v1.API
}

func CreatePrometheus(url string) *Prometheus {
	return &Prometheus{
		addr: url,
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
