package errors

import "errors"

var (
	ErrCannotFindObjectHeader = errors.New("cannot find object header by hash")
	ErrCannotFindObject       = errors.New("cannot find object by hash")
	ErrCannotFindRootHash     = errors.New("cannot find root hash by key")
	ErrUnableToProcessRequest = errors.New("unable to process fields from the request")
)
