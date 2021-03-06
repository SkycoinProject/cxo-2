package data

import (
	"time"

	"github.com/SkycoinProject/cxo-2/pkg/model"
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

type objectDAO struct {
	ID               string
	ObjectHeaderHash string `storm:"index"`
	Object           model.Object
}

type objectInfo struct {
	ID   string
	Path string `storm:"index"`
}

type app struct {
	Pk      int `storm:"id,increment"`
	Address string
	Name    string
}
