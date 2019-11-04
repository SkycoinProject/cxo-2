package main

import (
	"fmt"
	"os"

	"github.com/SkycoinPro/cxo-2-node/src/cli"
	"github.com/SkycoinPro/cxo-2-node/src/config"
)

func main() {
	cfg := config.LoadConfig()

	cxoNodeCLI, err := cli.NewCLI(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := cxoNodeCLI.Execute(); err != nil {
		os.Exit(1)
	}
}
