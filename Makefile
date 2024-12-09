VERSION=$(git describe --tags --always)


run:
	go run \
		-ldflags="-X main.Version=$(VERSION)" \
		./cmd/template-dockerfiles \
		--config example/build.yaml \
		--tag v1.2.3

build: bin/template-dockerfiles
	go build \
		-ldflags="-X main.Version=$(VERSION)" \
		-o bin/template-dockerfiles ./cmd/template-dockerfiles

clean:
	@rm -rfv bin
