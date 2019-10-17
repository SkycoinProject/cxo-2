#!/usr/bin/env bash

set -e -o pipefail

go build -o "$GOPATH/bin/cxo-node ." "$GOPATH/src/github.com/SkycoinPro/cxo-2-node/cmd/node"
go build -o "$GOPATH/bin/cxo-node-cli ." "$GOPATH/src/github.com/SkycoinPro/cxo-2-node/cmd/cli"