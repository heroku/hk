.PHONY: default build

default: build
	./hk ${COMMAND}

build:
	go build -o hk
