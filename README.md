Template-Dockerfiles
====================

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/tgagor/template-dockerfiles/build-go.yml)
![GitHub](https://img.shields.io/github/license/tgagor/template-dockerfiles)
![GitHub Release Date](https://img.shields.io/github/release-date/tgagor/template-dockerfiles)

A versatile Docker image builder that uses [Go Templates](https://pkg.go.dev/text/template) extended with [Sprig functions](http://masterminds.github.io/sprig/lists.html) to dynamically generate Dockerfiles, validate configurations, and build container images efficiently. The app supports parameterized builds, parallel execution, and customization for streamlined container development.


## **Parameters**

```bash
A CLI tool for building Docker images with configurable Dockerfile templates and multi-threaded execution.

When 'docker build' is just not enough. :-)

Usage:
  td [flags]

Flags:
  -b, --build           Build Docker images after templating
  -c, --config string   Path to the configuration file (required)
  -d, --delete          Delete templated Dockerfiles after successful building
  -e, --engine string   Select the container engine to use: docker or buildx (default "docker")
  -h, --help            help for td
  -i, --image string    Limit the build to a single image
      --parallel int    Specify the number of threads to use, defaults to number of CPUs (default 20)
  -p, --push            Push Docker images after building
  -s, --squash          Squash images to reduce size (experimental)
  -t, --tag string      Tag to use as the image version
  -v, --verbose         Increase verbosity of output
  -V, --version         Display the application version and exit

```

## **Installation**

Download latest version of file:

```bash
sudo curl -sLfo /usr/local/bin/td https://github.com/tgagor/template-dockerfiles/releases/latest/download/td-linux-amd64
sudo chmod +x /usr/local/bin/td
```
Alternatively, extract it to any location under `PATH`. Or download with Go Lang:

```bash
go install github.com/tgagor/template-dockerfiles/cmd/td@latest
```
Ensure you have `GOPATH`, in your `PATH`.


## **Example Configuration**

### Complete Examples

For complete examples of configuration check [example](./example/) directory.

### Configuration File

Start defining your build configuration file. It's a playbook by which your Docker image templates will be generated and build:

```yaml
registry: repo.local
prefix: my-base
maintainer: Awesome Developer <awesome@mail>

images:
  jdk:
    dockerfile: jdk/Dockerfile.tpl
    variables:
      alpine:
        - "3.19"
        - "3.20"
      java:
        - 11
        - 17
        - 21
    tags:
      - jdk:{{ .tag }}-{{ .java }}-alpine{{ .alpine }}
      - jdk:{{ .java }}-alpine{{ .alpine | splitList "." | first }}
```

Call build like:

```bash
td --config build.yaml --tag v1.2.3 --build --delete
```

Which will produce 2x3 -> 6 images, with 12 labels:

```bash
repo.local/my-base/jdk:v1.2.3-11-alpine3.19
repo.local/my-base/jdk:11-alpine3
repo.local/my-base/jdk:v1.2.3-17-alpine3.19
repo.local/my-base/jdk:17-alpine3
repo.local/my-base/jdk:v1.2.3-21-alpine3.19
repo.local/my-base/jdk:21-alpine3
repo.local/my-base/jdk:v1.2.3-11-alpine3.20
repo.local/my-base/jdk:11-alpine3
repo.local/my-base/jdk:v1.2.3-17-alpine3.20
repo.local/my-base/jdk:17-alpine3
repo.local/my-base/jdk:v1.2.3-21-alpine3.20
repo.local/my-base/jdk:21-alpine3
```

**Order of values under `variables` block is used to determine the order of labels creation.**

## **Configuration Format**

This file format defines the configuration for dynamically generating Docker images using Jinja2 templates. It specifies global settings, image definitions, and build parameters.


### **`registry`** (Optional)
- **Description**: The Docker registry to which images will be pushed. Skip to use Docker Hub.
- **Type**: String
- **Example**:
  ```yaml
  registry: repo.local
  ```

### **`prefix`** (Optional)
- **Description**: A prefix applied to all image names for organizational purposes. Might be a Docker Hub user name.
- **Type**: String
- **Example**:
  ```yaml
  prefix: my-base
  ```

### **`maintainer`** (Optional)
- **Description**: The maintainer's name and contact information.
- **Type**: String
- **Example**:
  ```yaml
  maintainer: Name <email@domain>
  ```

### **`labels`** (Optional)
- **Description**: Global labels that would be added to each image automatically.
- **Type**: Dictionary of strings
- **Example**:
  ```yaml
  labels:
    - org.opencontainers.image.licenses: License(s) under which contained software is distributed as an [SPDX License Expression](https://spdx.github.io/spdx-spec/v2.3/SPDX-license-expressions/).
    - org.opencontainers.image.title: Human-readable title of the image (string).
    - org.opencontainers.image.description: |
     Human-readable description of the software packaged in the image.
     (multiline string).
  ```
- **Notes**:
  - I recommend to follow [OCI Label Schema](https://github.com/opencontainers/image-spec/blob/main/annotations.md), app will add some of them automatically.
  - Even those labels can be templated, but as they're global, you should only use variables available in all images. Otherwise they might be evaluated to: `<no value>`, unless you filter those out with additional conditions.

## **Images Section**

### **`images`** (Required)
Defines the Docker images to build. Each image has specific attributes such as its Dockerfile, variables, and labels. Images are build in order, top-down, which allows to construct dependencies between images.

#### **Image Definition**
Each image is identified by a key (e.g., `base`, `jdk`, `jre`) and contains the following attributes:

### **`dockerfile`** (Required)
- **Description**: The path to the Dockerfile template used to build the image. [Go Templates](https://pkg.go.dev/text/template) extended with [Sprig functions](http://masterminds.github.io/sprig/lists.html) are supported.
- **Type**: String
- **Example**:
  ```yaml
  dockerfile: base/Dockerfile.tpl
  ```

### **`variables`** (Optional)
- **Description**: A dictionary of variables used to parameterize the Dockerfile template.
- **Type**: Dictionary of lists
- **Example**:
  ```yaml
  variables:
    alpine:
      - "3.20"
      - "3.19"
  ```

- **Notes**:
  - Builder generates a Cartesian product of all variables (all combinations).
  - The variables can have multiple values, allowing builds for different configuration sets.
  - Variables are substituted into the template during build.

### **`tags`** (Required)
- **Description**: A list of names and tags to tag the generated Docker images.
- **Type**: List of strings
- **Example**:
  ```yaml
  labels:
    - base:{{ .tag }}-alpine{{ .alpine }}
    - base:alpine{{ .alpine }}
  ```
- **Notes**:
  - Labels support [Go Templates](https://pkg.go.dev/text/template) extended with [Sprig functions](http://masterminds.github.io/sprig/lists.html). For example, `{{ .alpine | splitList "." | first }}` extracts the major version from `alpine`.
  - `tag` argument is provided by `--tag`/`-t` parameter, which reflects the image version.

### **`labels`** (Optional)
- **Description**: Per image labels, that would be added to each image.
- **Type**: Dictionary of strings
- **Example**:
  ```yaml
  labels:
    - org.opencontainers.image.base.name: alpine:{{ .alpine }}
    - org.opencontainers.image.description: |
     Human-readable description of the software packaged in the image.
     (multiline string).
  ```
- **Notes**:
  - I recommend to follow [OCI Label Schema](https://github.com/opencontainers/image-spec/blob/main/annotations.md), app will add some of them automatically.
  - Labels can be templated and they will override global labels of same name.

## **Multi-Platform Builds**

For multi-platform builds, you need to prepare your build environment. This guide uses QEMU emulation, which provides a broad list of platforms available out of the box.

Building multi-platform images requires support from base images and tools that work on specific platforms. Verify compatibility before you start.

### Install QEMU

This example uses Ubuntu, but the steps should be similar on other platforms.

1. Update your package list and install QEMU:

    ```bash
    sudo apt update
    sudo apt-get install -y qemu-system
    ```

2. Use the [tonistiigi/binfmt](https://github.com/tonistiigi/binfmt) image to install QEMU and register the executable types on the host. This allows the `buildx` builder to recognize available target platforms:

    ```bash
    docker run --privileged --rm tonistiigi/binfmt --install all
    ```

3. Verify the installation with:

    ```bash
    docker buildx ls
    ```

    You should see output similar to:

    ```bash
    NAME/NODE     DRIVER/ENDPOINT   STATUS    BUILDKIT   PLATFORMS
    default*      docker
     \_ default    \_ default       running   v0.17.3    linux/amd64 (+3), linux/arm64, linux/arm (+2), linux/ppc64le, (3 more)
    ```

### Enable Containerd Image Store

To enable the containerd snapshotters feature, follow these steps:

1. Add the following configuration to your `/etc/docker/daemon.json` file:

    ```json
    {
      "features": {
        "containerd-snapshotter": true
      }
    }
    ```

2. Save the file.

3. Restart the Docker daemon for the changes to take effect:

    ```bash
    sudo systemctl restart docker
    ```

4. After restarting the daemon, verify that you're using containerd snapshotter storage drivers:

    ```bash
    docker info -f '{{ .DriverStatus }}'
    ```

    You should see output similar to:

    ```bash
    [[driver-type io.containerd.snapshotter.v1]]
    ```

### Config file

Now it's time to add required platforms to your configuration, you can put them in the global scope or per image, for example:

```yaml
platforms:
  - linux/amd64
  - linux/arm64

images:
  base:
    dockerfile: base/Dockerfile.tpl
    ...
  jre:
    dockerfile: jre/Dockerfile.tpl
    # overwrite platforms for this image only
    platforms:
      - linux/amd64
```

You can also use few variables in templates that would refer to your current platform, like:
  - `BUILDPLATFORM` — matches the current machine. (e.g. linux/amd64)
  - `BUILDOS` — os component of BUILDPLATFORM, e.g. linux
  - `BUILDARCH` — e.g. amd64, arm64, riscv64
  - `BUILDVARIANT` — used to set ARM variant, e.g. v7
  - `TARGETPLATFORM` — The value set with --platform flag on build
  - `TARGETOS` - OS component from --platform, e.g. linux
  - `TARGETARCH` - Architecture from --platform, e.g. arm64
  - `TARGETVARIANT` -

### Now build something

Check your images, for multiple platforms, with:

```bash
docker image inspect \
  --format "{{.ID}} {{.RepoTags}} {{.Architecture}}" \
  $(docker image ls -q)
```

### For more information, check the official documentation:
- [Building Multi-Platform Images](https://docs.docker.com/build/building/multi-platform/)
- [Containerd Storage](https://docs.docker.com/engine/storage/containerd/)
- [Managing Builders](https://docs.docker.com/build/builders/manage/)

## **Validation and Recommendations**

### Required Fields
1. `images`: At least one image must be defined with a valid `dockerfile`.

### Optional Enhancements
1. Use the `prefix` field for consistent image organization.
2. Add meaningful labels to enhance discoverability and traceability.
3. Keep in mind that order of variables, determine order of labeling and some labels might overwrite previously created.

### Parallelism
1. Tool detects number of available CPU and run as many jobs as possible.
2. For debugging, it might be easier to use `--parallel 1 --verbose` to limit amount of messages produced.

### Debugging
1. Use `--verbose` flag. It will produce a lot of debug information.
2. Without `--build` flag, script will just template Dockerfiles, so you can check them for correctness.
3. Use `--image` to build just single set of images, instead of building them all.

## **Advanced Tips**

### **Dynamic Tags**
Use [Go Templates](https://pkg.go.dev/text/template) support by [Sprig functions](http://masterminds.github.io/sprig/lists.html) to create dynamic expressions like `{{ .alpine | splitList "." | first }}` to generate tags or labels dynamically.

### **Dynamic Labels**
Same approach as for tags applies to labels. You can template them as you whish.

### **Additional Variables**

All the variables available for templating:
  - `registry` - If provided.
  - `prefix` - If provided.
  - `maintainer` - If provided.
  - `tag` - If `--tag` flag used.
  - `image` - A key from `images` in the config file, useful for conditions.
  -	`labels` - Some generated automatically, then from "global scope" (at the top of config file), merged with "per image" labels,
  - And finally, whatever you define in `variables` blocks.
