package model

import (
	"fmt"
	"time"
)

// RootHash model
type RootHash struct {
	Publisher        string    `json:"publisher"`
	Signature        string    `json:"signature"`
	Sequence         uint64    `json:"sequence"`
	Timestamp        time.Time `json:"timestamp"`
	ObjectHeaderHash string    `json:"objectHeaderHash"`
}

// Key returns parcel key constructed in "sequence_publisher" format
func (r *RootHash) Key() string {
	return fmt.Sprintf("%v_%s", r.Sequence, r.Publisher)
}

// Parcel model
type Parcel struct {
	ObjectHeaders []ObjectHeader `json:"objectHeaders"`
	Objects       []Object       `json:"objects"`
}

// ObjectHeader model
type ObjectHeader struct {
	ObjectHash             string               `json:"objectHash"`
	ObjectSize             uint64               `json:"objectSize"`
	ExternalReferences     []ExternalReferences `json:"externalReferences"`
	ExternalReferencesSize uint64               `json:"externalReferencesSize"`
	Meta                   []Meta               `json:"meta"`
}

// Meta model
type Meta struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ExternalReferences model
type ExternalReferences struct {
	//Index                   uint64 `json:"index"` // FIXME - probably not needed
	ObjectHeaderHash        string `json:"objectHeaderHash"`
	Size                    uint64 `json:"size"`
	RecursiveSizeFirstLevel uint64 `json:"recursiveSizeFirstLevel"`
	RecursiveSizeFirstTotal uint64 `json:"recursiveSizeTotal"`
}

// Object model
type Object struct {
	Length uint64 `json:"length"`
	Data   []byte `json:"data"`
}

// PublishDataRequest model
type PublishDataRequest struct {
	RootHash RootHash `json:"rootHash"`
	Parcel   Parcel   `json:"parcel"`
}

// GetObjectHeadersResponse model
type GetObjectHeadersResponse struct {
	ObjectHeaders []ObjectHeader `json:"objectHeaders"`
}
