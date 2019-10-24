package main

import (
	"github.com/SkycoinPro/cxo-2-node/src/config"
	"github.com/SkycoinPro/cxo-2-node/src/node"
	"github.com/SkycoinPro/cxo-2-node/src/node/database/boltdb"
)

func main() {
	cfg := config.LoadConfig()
	tearDown := boltdb.Init()
	defer tearDown()
	node.NewService(cfg).Run()
}
