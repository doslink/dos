.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/doslink/dos/cmd/swarm
	go build -o bin/gdos     github.com/doslink/dos/cmd/gdos
	go build -o bin/bootnode github.com/doslink/dos/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/doslink/dos/swarm/...
