package main

import (
	"github.com/SkycoinPro/cxo-2-node/pkg/config"
	"github.com/SkycoinPro/cxo-2-node/pkg/node"
	"github.com/SkycoinPro/cxo-2-node/pkg/node/data"
)

func main() {
	cfg := config.LoadConfig()
	tearDown := data.Init()
	defer tearDown()
	node.NewService(cfg).Run()
}
