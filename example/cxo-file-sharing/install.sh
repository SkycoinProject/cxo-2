#!/bin/bash

#TODO check this, failing on Ubuntu currently
#-set -e -o pipefail

env go build -o "$GOPATH/bin/cxo-file-sharing.exe" "$GOPATH/src/github.com/SkycoinProject/cxo-2/example/cxo-file-sharing/cxo-file-sharing.go"
env go build -o "$GOPATH/bin/cxo-file-sharing-cli.exe" "$GOPATH/src/github.com/SkycoinProject/cxo-2/example/cxo-file-sharing/cli/cxo-file-sharing-cli.go"