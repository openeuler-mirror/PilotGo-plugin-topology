package client

import (
	"github.com/skydive-project/skydive/config"
	shttp "github.com/skydive-project/skydive/graffiti/http"
	"github.com/skydive-project/skydive/graffiti/logging"
)

// NewCrudClientFromConfig creates a new REST client on /api
func NewCrudClientFromConfig(authOptions *shttp.AuthenticationOpts) (*shttp.CrudClient, error) {
	restClient, err := NewRestClientFromConfig(authOptions)
	if err != nil {
		return nil, err
	}

	return shttp.NewCrudClient(restClient), nil
}

// NewRestClientFromConfig creates a new REST client
func NewRestClientFromConfig(authOptions *shttp.AuthenticationOpts) (*shttp.RestClient, error) {
	tlsConfig, err := config.GetTLSClientConfig(true)
	if err != nil {
		return nil, err
	}

	sa, err := config.GetOneAnalyzerServiceAddress()
	if err != nil && err != config.ErrNoAnalyzerSpecified {
		logging.GetLogger().Errorf("Unable to parse analyzer client %s", err.Error())
		return nil, err
	}

	url, err := config.GetURL("http", sa.Addr, sa.Port, "/api/")
	if err != nil {
		return nil, err
	}
	return shttp.NewRestClient(url, authOptions, tlsConfig), nil
}
