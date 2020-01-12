package errors

import "errors"

var (
	ErrCannotFindObjectHeader = errors.New("cannot find object header by hash")
	ErrCannotFindRootHash     = errors.New("cannot find root hash by key")
	ErrCannotFindObjectPath   = errors.New("cannot find object path by hash")
	ErrUnableToProcessRequest = errors.New("unable to process fields from the request")
)
