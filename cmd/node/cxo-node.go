package main

import (
	"flag"

	"github.com/SkycoinPro/cxo-2-node/pkg/config"
	"github.com/SkycoinPro/cxo-2-node/pkg/node"
	"github.com/SkycoinPro/cxo-2-node/pkg/node/data"
)

func main() {
	local := flag.Bool("local", false, "enables node to run from the path it's been started from")
	flag.Parse()
	cfg := config.LoadConfig(local)
	tearDown := data.Init()
	defer tearDown()
	node.NewService(cfg).Run()
}
