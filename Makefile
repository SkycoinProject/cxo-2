
build : ## Build test
	rm -f ./integration/executables/*
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./integration/executables/node ./cmd/node
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./integration/executables/node-cli ./cmd/cli
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./integration/executables/test-runner ./integration/test-runner.go
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./integration/executables/cxo-file-sharing ./example/cxo-file-sharing/cxo-file-sharing.go
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./integration/executables/cxo-file-sharing-cli ./example/cxo-file-sharing/cli/cxo-file-sharing-cli.go


run : ## Run test
	rm -rf ./files/node-*/*
	docker rm node-1 node-2 node-3 node-4 node-5
	cd integration && \
	docker-compose up