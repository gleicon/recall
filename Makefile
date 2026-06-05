.PHONY: all build test clean

all: build

build:
	go build -o recall .

test:
	go test -v ./...

clean:
	rm -f recall
