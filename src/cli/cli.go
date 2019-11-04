package cli

import (
	"flag"
	"fmt"

	"github.com/SkycoinPro/cxo-2-node/src/cli/client"
	"github.com/SkycoinPro/cxo-2-node/src/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const version = "0.1.1"

// NewCLI creates a cli instance
func NewCLI(cfg config.Config) (*cobra.Command, error) {
	c := client.NewTrackerClient(cfg)

	cxoNodeCLI := &cobra.Command{
		Short: fmt.Sprintf("The cxo-node command line interface"),
		Use:   fmt.Sprintf("cxo-node-cli"),
	}

	commands := []*cobra.Command{
		subscribeCmd(c),
		announceDataCmd(c, cfg),
	}

	cxoNodeCLI.Version = version
	cxoNodeCLI.AddCommand(commands...)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	return cxoNodeCLI, nil
}
