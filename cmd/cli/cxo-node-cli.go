package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/SkycoinProject/cxo-2/pkg/cli"
	"github.com/SkycoinProject/cxo-2/pkg/config"
)

func main() {
	local := flag.Bool("local", false, "enables node to run from the path it's been started from")
	flag.Parse()
	cfg := config.LoadConfig(local)

	cxoNodeCLI, err := cli.NewCLI(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := cxoNodeCLI.Execute(); err != nil {
		os.Exit(1)
	}
}
