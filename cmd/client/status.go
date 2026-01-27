package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
)

// StatusCmd implents the skydive 'status' command that
// return the status of an analyzer by quering its API
var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show analyzer status",
	Long:  "Show analyzer status",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := CrudClient.Request("GET", "status", nil, nil)
		if err != nil {
			exitOnError(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			data, _ := ioutil.ReadAll(resp.Body)
			exitOnError(fmt.Errorf("Failed to get status, %s: %s", resp.Status, data))
		}

		var status interface{}
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&status); err != nil {
			exitOnError(err)
		}

		printJSON(&status)
	},
}
