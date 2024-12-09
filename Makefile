VERSION=$(git describe --tags --always)


run:
	go run \
		-ldflags="-X main.version=$(VERSION)" \
		./cmd/template-dockerfiles \
		--config example/build.yaml \
		--tag v1.2.3

bin/template-dockerfiles: build

build:
	go build \
		-ldflags="-X main.version=$(VERSION)" \
		-o bin/template-dockerfiles ./cmd/template-dockerfiles

clean:
	@rm -rfv bin
