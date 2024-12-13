VERSION ?= $(shell git describe --tags --always)


run:
	go run \
		-ldflags="-X main.Version=$(VERSION)" \
		./cmd/td \
		--config example/build.yaml \
		--tag v1.2.3

bin/td: build

build:
	go build \
		-ldflags="-X main.Version=$(VERSION)" \
		-o bin/td ./cmd/td

clean:
	@rm -rfv bin
