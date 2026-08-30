.PHONY: build test test-go regression vet smoke release-check install

build:
	go build ./cmd/ddocs ./cmd/demon

test: test-go regression

test-go:
	go test ./... -count=1

regression:
	go test ./tests -run 'TestGoCLIRegressionMatrix' -count=1 -v

vet:
	go vet ./...

smoke:
	go run ./tools/smoke

release-check: test vet build smoke

install:
	go install ./cmd/ddocs ./cmd/demon
