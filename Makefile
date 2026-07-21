.PHONY: build test test-go regression vet smoke release-check install clean

build:
	go build -o bin/ddocs ./cmd/ddocs
	go build -o bin/demon ./cmd/demon

test: test-go regression

test-go:
	go test ./internal/... ./cmd/... -count=1

regression:
	go test ./tests -run 'TestGoCLIRegressionMatrix' -count=1 -v

vet:
	go vet ./...

smoke:
	go run ./tools/smoke

release-check: test vet build smoke

install:
	go install ./cmd/ddocs
	go install ./cmd/demon

clean:
	rm -rf bin
