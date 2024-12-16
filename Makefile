VERSION ?= $(shell git describe --tags --always)


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

clean:
	@rm -rfv bin
