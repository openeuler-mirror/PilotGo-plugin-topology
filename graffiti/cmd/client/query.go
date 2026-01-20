package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	api "atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/api/client"
	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/logging"
)

var (
	delete       bool
	outputFormat string
)

// QueryCmd skydive topology query command
var QueryCmd = &cobra.Command{
	Use:   "query [gremlin]",
	Short: "Issue Gremlin queries",
	Long:  "Issue Gremlin queries",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || args[0] == "" {
			cmd.Usage()
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		gremlinQuery := args[0]
		queryHelper := api.NewGremlinQueryHelper(CrudClient.RestClient)

		switch outputFormat {
		case "json":
			data, err := queryHelper.Query(gremlinQuery)
			if err != nil {
				exitOnError(err)
			}

			var out bytes.Buffer
			json.Indent(&out, data, "", "\t")
			out.WriteTo(os.Stdout)
		case "dot":
			header := make(http.Header)
			header.Set("Accept", "text/vnd.graphviz")
			resp, err := queryHelper.Request(gremlinQuery, header)
			if err != nil {
				exitOnError(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				data, _ := io.ReadAll(resp.Body)
				exitOnError(fmt.Errorf("%s: %s", resp.Status, string(data)))
			}
			bufio.NewReader(resp.Body).WriteTo(os.Stdout)
		case "pcap":
			header := make(http.Header)
			header.Set("Accept", "application/vnd.tcpdump.pcap")
			resp, err := queryHelper.Request(gremlinQuery, header)
			if err != nil {
				exitOnError(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				data, _ := io.ReadAll(resp.Body)
				exitOnError(fmt.Errorf("%s: %s", resp.Status, string(data)))
			}

			bufio.NewReader(resp.Body).WriteTo(os.Stdout)
		default:
			logging.GetLogger().Errorf("Invalid output format %s", outputFormat)
			os.Exit(1)
		}
	},
}

func init() {
	QueryCmd.Flags().StringVarP(&outputFormat, "format", "", "json", "Output format (json, dot or pcap)")
	QueryCmd.Flags().BoolVarP(&delete, "delete", "", false, "Deletes the query output nodes and edges")
}
