package main

import (
	"github.com/SkycoinPro/cxo-2-node/src/config"
	"github.com/SkycoinPro/cxo-2-node/src/node"
)

func main() {
	cfg := config.LoadConfig()
	node.NewService(cfg).Run()
}
