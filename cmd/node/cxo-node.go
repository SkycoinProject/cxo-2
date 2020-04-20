package main

import (
	"flag"

	"github.com/SkycoinProject/cxo-2/pkg/config"
	"github.com/SkycoinProject/cxo-2/pkg/node"
	"github.com/SkycoinProject/cxo-2/pkg/node/data"
)

func main() {
	local := flag.Bool("local", false, "enables node to run from the path it's been started from")
	flag.Parse()
	cfg := config.LoadConfig(local)
	tearDown := data.Init()
	defer tearDown()
	node.NewService(cfg).Run()
}
