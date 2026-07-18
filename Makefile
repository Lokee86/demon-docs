.PHONY: build test test-go regression vet smoke release-check install clean

build:
	go build -o bin/doc-ledger ./cmd/doc-ledger

test: test-go regression

test-go:
	go test ./internal/... ./cmd/... -count=1

regression:
	go test ./tests -run 'TestGoCLIRegressionMatrix' -count=1 -v

vet:
	go vet ./...

smoke: build
	./bin/doc-ledger --help >/dev/null
	./bin/doc-ledger --version
	./bin/doc-ledger config paths >/dev/null

release-check: test vet build smoke

install:
	go install ./cmd/doc-ledger

clean:
	rm -rf bin
