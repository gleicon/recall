.PHONY: all build test clean

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X github.com/gleicon/recall/cmd.Version=$(VERSION)

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o recall .

test:
	go test -v ./...

clean:
	rm -f recall
