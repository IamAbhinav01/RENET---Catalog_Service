.PHONY: dev run test fmt vet build

dev:
    air

run:
    go run .

test:
    go test ./...

fmt:
    gofmt -w .

vet:
    go vet ./...

build:
    go build ./...