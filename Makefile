.PHONY: default build

default: build
	./gonpm ${COMMAND}

build:
	go build
