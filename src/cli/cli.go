package cli

import (
	"flag"
	"fmt"

	"github.com/SkycoinPro/cxo-2-node/src/cli/client"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const version = "0.1.1"

// NewCLI creates a cli instance
func NewCLI(client *client.TrackerClient) (*cobra.Command, error) {

	cxoNodeCLI := &cobra.Command{
		Short: fmt.Sprintf("The cxo-node command line interface"),
		Use:   fmt.Sprintf("cxo-node-cli"),
	}

	commands := []*cobra.Command{
		subscribeCmd(client),
		saveDataCmd(client),
	}

	cxoNodeCLI.Version = version
	cxoNodeCLI.AddCommand(commands...)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	return cxoNodeCLI, nil
}
