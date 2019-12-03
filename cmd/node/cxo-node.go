package main

import (
	"github.com/SkycoinPro/cxo-2-node/pkg/config"
	"github.com/SkycoinPro/cxo-2-node/pkg/node"
	"github.com/SkycoinPro/cxo-2-node/pkg/node/database/boltdb"
)

func main() {
	cfg := config.LoadConfig()
	tearDown := boltdb.Init()
	defer tearDown()
	node.NewService(cfg).Run()
}
