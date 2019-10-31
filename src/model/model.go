package model

import "time"

// RootHash model
type RootHash struct {
	Publisher      string `json:"publisher"`
	Signature      string `json:"signature"`
	Sequence       uint64 `json:"sequence"`
	DataObjectHash string `json:"dataObjectHash"`
}

// DataObject model
type DataObject struct {
	Header   Header   `json:"header"`
	Manifest Manifest `json:"manifest"`
	Object   Object   `json:"object"`
}

// Header model
type Header struct {
	Timestamp    time.Time `json:"timestamp"`
	ManifestHash string    `json:"manifestHash"`
	ManifestSize uint64    `json:"manifestSize"`
	DataHash     string    `json:"dataHash"`
	DataSize     uint64    `json:"dataSize"`
}

// Manifest model
type Manifest struct {
	Length uint64            `json:"length"`
	Hashes []ObjectStructure `json:"objects"`
	Meta   []string          `json:"meta"`
}

// ObjectStructure model
type ObjectStructure struct {
	Index                   uint64 `json:"index"`
	Hash                    string `json:"hash"`
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
	RootHash   RootHash   `json:"rootHash"`
	DataObject DataObject `json:"dataObject"`
}
