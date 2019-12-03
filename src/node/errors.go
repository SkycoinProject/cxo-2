package node

import "errors"

var (
	errCannotFindObjectHeader = errors.New("cannot find object header by hash")
)
