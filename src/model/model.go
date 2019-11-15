package model

import (
	"fmt"
	"time"
)

// RootHash model
type RootHash struct {
	Publisher string    `json:"publisher"`
	Signature string    `json:"signature"`
	Sequence  uint64    `json:"sequence"`
	Timestamp time.Time `json:"timestamp"`
}

// Key returns parcel key constructed in "sequence_publisher" format
func (r *RootHash) Key() string {
	return fmt.Sprintf("%v_%s", r.Sequence, r.Publisher)
}

// Parcel model
type Parcel struct {
	Headers []Header     `json:"headers"`
	Objects []DataObject `json:"objects"`
}

// Header model
type Header struct {
	ManifestHash       string               `json:"manifestHash"`
	ManifestSize       uint64               `json:"manifestSize"`
	Length             uint64               `json:"manifestLength"` // FIXME - probably not needed?
	DataHash           string               `json:"dataHash"`
	DataSize           uint64               `json:"dataSize"`
	Meta               []string             `json:"meta"`
	ExternalReferences []ExternalReferences `json:"externalReferences"`
}

// ExternalReferences model
type ExternalReferences struct {
	Index                   uint64 `json:"index"` // FIXME - probably not needed
	Hash                    string `json:"hash"`
	Size                    uint64 `json:"size"`
	RecursiveSizeFirstLevel uint64 `json:"recursiveSizeFirstLevel"`
	RecursiveSizeFirstTotal uint64 `json:"recursiveSizeTotal"`
}

// DataObject model
type DataObject struct {
	Length uint64 `json:"length"`
	Data   []byte `json:"data"`
}

// PublishDataRequest model
type PublishDataRequest struct {
	RootHash RootHash `json:"rootHash"`
	Parcel   Parcel   `json:"parcel"`
}
