.PHONY: build test vet fmt compose-build compose-up compose-down compose-logs run-api run-ocr

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

compose-build:
	podman compose build

compose-up:
	podman compose up -d

compose-down:
	podman compose down

compose-logs:
	podman compose logs -f

run-api:
	go run ./cmd/api

run-ocr:
	go run ./cmd/ocr
