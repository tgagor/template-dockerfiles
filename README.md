Template-Dockerfiles
====================

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/tgagor/template-dockerfiles/build-go.yml)
![GitHub](https://img.shields.io/github/license/tgagor/template-dockerfiles)
![GitHub Release Date](https://img.shields.io/github/release-date/tgagor/template-dockerfiles)

A versatile Docker image builder that uses [Go Templates](https://pkg.go.dev/text/template) extended with [Sprig functions](http://masterminds.github.io/sprig/lists.html) to dynamically generate Dockerfiles, validate configurations, and build container images efficiently. The app supports parameterized builds, parallel execution, and customization for streamlined container development.


Parameters
----------

```bash
A CLI tool for building Docker images with configurable Dockerfile templates and multi-threaded execution.

When 'docker build' is just not enough. :-)

Usage:
  td [flags]

Flags:
  -b, --build           Build Docker images after templating
  -c, --config string   Path to the configuration file (required)
  -d, --delete          Delete templated Dockerfiles after successful building
  -h, --help            help for td
      --parallel int    Specify the number of threads to use, defaults to number of CPUs (default 20)
  -p, --push            Push Docker images after building
  -t, --tag string      Tag to use as the image version
  -v, --verbose         Increase verbosity of output
  -V, --version         Display the application version and exit

```

Installation
------------

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


## **Validation and Recommendations**

### Required Fields
1. `images`: At least one image must be defined with a valid `dockerfile`.

### Optional Enhancements
1. Use the `prefix` field for consistent image organization.
2. Add meaningful labels to enhance discoverability and traceability.
3. Keep in mind that order of variables, determine order of labeling and some labels might overwrite previously created. Use `--dry-run` mode to determine the result.

### Parallelism
1. Tool detects number of available CPU and run as many jobs as possible.
2. For debugging, it might be easier to use `--parallel 1 --verbose` to limit amount of messages produced.

### Debugging
1. Use `--verbose` flag. It will produce a lot of debug information.
2. Use `--dry-run` flag just to see what would be produced without actually building anything.

## **Advanced Tips**

1. **Dynamic Tags**: Use [Go Templates](https://pkg.go.dev/text/template) support by [Sprig functions](http://masterminds.github.io/sprig/lists.html) to create dynamic expressions like `{{ .alpine | splitList "." | first }}` to generate tags or labels dynamically.
