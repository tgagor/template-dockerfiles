## Known bugs

- None, for now.
- **not a bug** Parallelism on per image level.

  Because of how dependencies between images work, tool builds images starting from top to down. Parallel build happen only in the scope of specific image. This means, that if you can run 10 parallel processes, but there will be only 2 images - then only 2 parallel processes would be started.

  There might be images, that will be built one-by-one, where they could be built in parallel, because they do not depend on each other.

  There are 2 ways to achieve that:
  1. Analyze final/templated Dockerfile for image references and if they're not referring to other images, they can be started faster -> it would be hard!
  2. Allow to order images with additional field/flag in configuration, for ex.: `priority` or `order`. All images with same number would be build at the same time. This could be error prone (human mistakes) and would result in more complex configuration files.

  For now, I don't see it worth to improve.


Ideas and wishlist
==================

1.  **Feature**: Allow templating any `*.tpl` files, instead of just Dockerfiles.

    Not that hard to do, but I will need to "clone" workdir per build process to ensure that files are not leaking between parallel builds.

    I have to also consider if it would be a good practice, as templating "hidden" in files other than Dockerfiles, might be harder to track and error prone. For now you can just just few different files and `COPY` in templated condition.

2.  **Feature**: It might make sense to release a dedicated CLI tool just for Docker image squashing. This tool I can use as a library in my tool.

3.  **Feature**: Add support for alternative builders.

    Probably by:  flag, like:
    - adding dedicated `--builder` flag, with params like:
        - `podman`
        - `kaniko` -> https://github.com/GoogleContainerTools/kaniko
        - `buildx`
    - depending on builder, some feature might not be available or not (for ex. squashing)

5.  **Feature**: Add support for multi-arch image builds

    Easiest to achieve with `buildx` builder. Could be a "killer" feature as with multiple platforms you have those many corner cases, small package names differences, different URLs for dependency downloads, so exactly what I try to improve with my tool.

    Ofc it would add a lot of complexity, so I have to stabilize basic functionality and refactor code to be ready for that.

    It might work like:

    - allow to configure `platform` in config, for ex:
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
    - proper verification for allowed `platform` values have to be added, plus a check if we will be able to build for it
    - then platforms would be used to add to `buildx` command flags and injected into `variables` if anyone need it for templating
    - might make sense to expose other variables, that could be used to simplify templating:
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

    More hints:
    - https://www.docker.com/blog/multi-arch-build-and-images-the-simple-way/
    - https://docs.docker.com/build/building/multi-platform/

9.  **Feature**: Add `--workdir` flag

    To allow defining where temporary template files or tar packages should be placed, as current dir or Dockerfile's dir might not be preferred in some situations (for ex. RO).

10. **Feature**: Add `--progress` flag

    To improve visibility of tasks in the background.
    Hint: https://github.com/schollz/progressbar

11. **Refactor**: The chaining in Runner and Cmd modules returns full objects, when it probably should return references. Maybe fixing it, would allow me to simplify calls like:
```
b.tagTasks = b.tagTasks.AddTask(tagger)
```
to just
```
b.tagTasks.AddTask(tagger)
```
