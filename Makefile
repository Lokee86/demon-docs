.PHONY: build test test-go test-python parity install clean

build:
	go build -o bin/doc-ledger ./cmd/doc-ledger

test: test-go

test-go:
	go test ./...

test-python:
	python -m pytest tests -q

parity:
	go test ./tests -run TestPythonGoParity -count=1

install:
	go install ./cmd/doc-ledger

clean:
	rm -rf bin
