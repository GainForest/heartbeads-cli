.PHONY: build test lint fmt clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags="-s -w -X main.Version=$(VERSION)" -o hb .

test:
	go test ./...

lint:
	golangci-lint run

fmt:
	go fmt ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -f hb hb.exe coverage.out coverage.html
