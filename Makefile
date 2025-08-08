VERSION ?= $(shell git describe --tags --always)
GOBIN ?= $(shell go env GOPATH)/bin

run:
	go run \
		-ldflags="-X main.BuildVersion=$(VERSION)" \
		./cmd/td \
		--config example/complex.yaml \
		--tag v1.2.3

bin/td: build

build:
	go build \
		-ldflags="-X main.BuildVersion=$(VERSION)" \
		-o bin/td ./cmd/td

test: bin/td
	go test -v ./...

install: bin/td
	@mkdir -p $(GOBIN)
	@cp bin/td $(GOBIN)/td
	@echo "Installed td to $(GOBIN)/td"

$(GOBIN)/goimports:
	@go install golang.org/x/tools/cmd/goimports@v0.32.0

$(GOBIN)/gocyclo:
	@go install github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0

$(GOBIN)/golangci-lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

$(GOBIN)/gocritic:
	@go install github.com/go-critic/go-critic/cmd/gocritic@v0.13.0

install-linters: $(GOBIN)/goimports $(GOBIN)/gocyclo $(GOBIN)/golangci-lint $(GOBIN)/gocritic
	@echo "Linters installed successfully."

lint: install-linters
	@pre-commit run -a

clean:
	@rm -rfv bin
	@find example -name '*.Dockerfile' -delete
	@find tests -name '*.Dockerfile' -delete
