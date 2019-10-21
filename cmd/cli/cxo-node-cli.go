package main

import (
	"fmt"
	"os"

	"github.com/SkycoinPro/cxo-2-node/src/cli/client"

	"github.com/SkycoinPro/cxo-2-node/src/cli"
	"github.com/SkycoinPro/cxo-2-node/src/config"
)

func main() {
	cfg := config.LoadConfig()
	c := client.NewTrackerClient(cfg)

	cxoNodeCLI, err := cli.NewCLI(c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := cxoNodeCLI.Execute(); err != nil {
		os.Exit(1)
	}
}
