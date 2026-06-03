.PHONY: all build test clean

all: build

build:
	go build -o technocore .

test:
	go test -v ./...

clean:
	rm -f technocore
