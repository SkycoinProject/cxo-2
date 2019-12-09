package data

import (
	"time"

	"github.com/SkycoinPro/cxo-2-node/pkg/model"
)

type rootHashDAO struct {
	ID       string
	RootHash model.RootHash
}

type objectHeaderDAO struct {
	ID           string
	RootHashKey  string    `storm:"index"`
	Timestamp    time.Time `storm:"index"`
	ObjectHeader model.ObjectHeader
}

type objectInfo struct {
	ID   string
	Path string
}
