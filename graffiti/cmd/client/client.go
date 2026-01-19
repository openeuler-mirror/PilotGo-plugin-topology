package client

import (
	"crypto/tls"
	"os"

	"github.com/spf13/cobra"

	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/http"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/logging"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/service"
	gtls "atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/tls"
)

const (
	CLIService service.Type = "cli"
)

var (
	// AuthenticationOpts Authentication options
	AuthenticationOpts http.AuthenticationOpts
	// CrudClient holds an API CRUD client
	CrudClient *http.CrudClient
	// Host holds the hostname
	Host string
	// HubAddress holds the hub address
	HubAddress string

	hubService service.Address
	tlsConfig  *tls.Config
	cafile     string
	certfile   string
	keyfile    string
	skipVerify bool
)

// ClientCmd describes the skydive client root command
var ClientCmd = &cobra.Command{
	Use:          "client",
	Short:        "Skydive client",
	Long:         "Skydive client",
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		hubService, err = service.AddressFromString(HubAddress)
		if err != nil {
			exitOnError(err)
		}

		if certfile != "" && keyfile != "" {
			tlsConfig, err = gtls.SetupTLSClientConfig(certfile, keyfile)
			if err != nil {
				exitOnError(err)
			}

			tlsConfig.InsecureSkipVerify = skipVerify

			if cafile != "" {
				tlsConfig.RootCAs, err = gtls.SetupTLSLoadCA(cafile)
				if err != nil {
					exitOnError(err)
				}
			}
		}

		url, _ := http.MakeURL("http", hubService.Addr, hubService.Port, "/api/", tlsConfig != nil)
		restClient := http.NewRestClient(url, &AuthenticationOpts, tlsConfig)
		CrudClient = http.NewCrudClient(restClient)
	},
}

func exitOnError(err error) {
	logging.GetLogger().Error(err)
	os.Exit(1)
}

func init() {
	ClientCmd.PersistentFlags().StringVarP(&AuthenticationOpts.Username, "username", "", AuthenticationOpts.Username, "username auth parameter")
	ClientCmd.PersistentFlags().StringVarP(&AuthenticationOpts.Password, "password", "", AuthenticationOpts.Password, "password auth parameter")

	ClientCmd.PersistentFlags().StringVarP(&cafile, "cacert", "", "", "certificate file to verify the peer")
	ClientCmd.PersistentFlags().StringVarP(&certfile, "cert", "", "", "client certificate file")
	ClientCmd.PersistentFlags().StringVarP(&keyfile, "key", "", "", "private key file name")
	ClientCmd.PersistentFlags().BoolVarP(&skipVerify, "insecure", "", false, "do not check server's certificate")

	ClientCmd.AddCommand(AlertCmd)
	ClientCmd.AddCommand(QueryCmd)
	ClientCmd.AddCommand(EdgeCmd)
	ClientCmd.AddCommand(NodeCmd)
	ClientCmd.AddCommand(ShellCmd)
	ClientCmd.AddCommand(TopologyCmd)
	ClientCmd.AddCommand(WorkflowCmd)
}
