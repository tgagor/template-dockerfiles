Ideas and wishlist
==================

1.  Image squashing

    Use flow as below:

    ```bash
    # build
    docker build \
    --build-arg TAG=${{ matrix.tag }} \
    --tag tgagor/centos:${{ matrix.tag }} ${{ matrix.tag }}/

    # squash
    docker run --name tgagor-${{ matrix.tag }} tgagor/centos:${{ matrix.tag }} true
    docker export tgagor-${{ matrix.tag }} | docker import \
    --change 'CMD ["/bin/bash"]' \
    --change 'LABEL maintainer="Tomasz Gągor <https://timor.site>"' \
    --change 'LABEL org.opencontainers.image.authors="Tomasz Gągor"' \
    --change 'LABEL org.opencontainers.image.licenses=GPL-2.0' \
    --change "LABEL org.opencontainers.image.version=$DOCKER_TAG" \
    --change "LABEL org.opencontainers.image.source=$GITHUB_SERVER_URL/$GITHUB_REPOSITORY" \
    --change "LABEL org.opencontainers.image.url=$GITHUB_SERVER_URL/$GITHUB_REPOSITORY" \
    --change "LABEL org.opencontainers.image.revision=$GITHUB_SHA" \
    --change "LABEL org.opencontainers.image.branch=${GITHUB_REF#refs/*/}" \
    --change "LABEL org.opencontainers.image.created=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    - tgagor/centos:${{ matrix.tag }}

    # tag
    docker tag tgagor/centos:${{ matrix.tag }} ghcr.io/tgagor/centos:${{ matrix.tag }}-${{ github.sha }}

    # push
    docker push ghcr.io/tgagor/centos:${{ matrix.tag }}-${{ github.sha }}
    ```

    With `docker inspect` or some go library, get `CMD`, `ENTRYPOINT`, `VOLUMES` and `LABELS` from image and squash it.

2. I might release a dedicated CLI tools that just squash docker images! Which I might use as a library in the templater!

2.  Allow templating not only Dockerfiles, but any TPL files

    Not that hard to do, but I will need to "clone" workdir per build process to ensure that files are not leaking between parallel builds.

3.  Allow to just generate template files, instead of building.

    This could even became a main app behavior, when building and pushing could be done with additional flags.

    - Build by default
        - `--template-only` flag introduced, which would set `build` and `push` to false

    - Template by default
        - `--build` flag introduced
        - `--del/--delete` flag introduced, to delete files after successful build
            - leaving files after unsuccessful build could even ease debugging
        - `--push` only valid when `--build` set
        - looks like more work, but feels more natural

4.  Add support for alternative builders
    - `--builder` flag introduced with params
        - podman
        - https://github.com/GoogleContainerTools/kaniko
        - buildx
    - depending on builder, some feature might not be available (like squashing)

5.  Add support for multi-arch image builds with buildx
    - https://www.docker.com/blog/multi-arch-build-and-images-the-simple-way/
    - https://docs.docker.com/build/building/multi-platform/
    - with a proper builder (`buildx` probably), allow to configure `platform` in config, eg:
      ```yaml
        images:
          demo:
            dockerfile: Dockerfile.tpl
            platforms:
              - linux/amd64
              - linux/arm64
              - linux/arm/v7
            variables:
              alpine:
                - "3.18"
                - "3.19"
                - "3.20"
      ```
    - then platforms would be used to add to `buildx` command flags and injected into `variables` if anyone need it for templating
    - might make sense to expose other variables, that could be easily used to match directives:
      ```bash
      BUILDPLATFORM — matches the current machine. (e.g. linux/amd64)
      BUILDOS — os component of BUILDPLATFORM, e.g. linux
      BUILDARCH — e.g. amd64, arm64, riscv64
      BUILDVARIANT — used to set ARM variant, e.g. v7
      TARGETPLATFORM — The value set with --platform flag on build
      TARGETOS - OS component from --platform, e.g. linux
      TARGETARCH - Architecture from --platform, e.g. arm64
      TARGETVARIANT
      ```
    - proper verification for allowed values have to be added
6. Add "name" of image to the configSet so it would be available as variable when templating
7. Use ordered YAML read as it happen now randomly that elements are executed in wrong order
   https://blog.labix.org/2014/09/22/announcing-yaml-v2-for-go
8. Add `--image` flag that would allow to build only one image, by it's name, which would support debugging of issues, without the need to rebuild all the images again and again
9. Add `--workdir` flag, to allow defining where temporary templates or tar packages should be placed as current dir or Dockerfiles dir might not be preffered in some situations
