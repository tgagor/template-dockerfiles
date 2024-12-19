Ideas and wishlist
==================

2. I might release a dedicated CLI tools that just squash docker images! Which I might use as a library in the templater!

2.  Allow templating not only Dockerfiles, but any TPL files

    Not that hard to do, but I will need to "clone" workdir per build process to ensure that files are not leaking between parallel builds.

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
9. Add `--workdir` flag, to allow defining where temporary templates or tar packages should be placed as current dir or Dockerfiles dir might not be preffered in some situations
10. Add `--progress` flag, that would enable progressbar, maybe: https://github.com/schollz/progressbar
