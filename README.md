Template-Dockerfiles
====================

A versatile Docker image builder that uses Jinja2 templates to dynamically generate Dockerfiles, validate configurations, and build container images efficiently. The app supports parameterized builds, parallel execution, and customization for streamlined container development.


## *Configuration Format**

This file format defines the configuration for dynamically generating Docker images using Jinja2 templates. It specifies global settings, image definitions, and build parameters.

---

## **Global Configuration**

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

---

## **Images Section**

### **`images`** (Required)
Defines the Docker images to build. Each image has specific attributes such as its Dockerfile, variables, and labels. Images are build in order, top-down, which allows to construct dependencies between images.

#### **Image Definition**
Each image is identified by a key (e.g., `base`, `jdk`, `jre`) and contains the following attributes:

### **`dockerfile`** (Required)
- **Description**: The path to the Dockerfile template used to build the image. [Jinja2 format](https://jinja.palletsprojects.com/en/stable/templates/) templates are supported.
- **Type**: String
- **Example**:
  ```yaml
  dockerfile: base/Dockerfile.jinja2
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
  - Builder generates a Cartesian product of all vaiables (all combinations).
  - The variables can have multiple values, allowing builds for different configurations.
  - Variables are substituted into the template during build.

### **`labels`** (Required)
- **Description**: A list of labels to tag the generated Docker images.
- **Type**: List of strings
- **Example**:
  ```yaml
  labels:
    - base:{{ tag }}-alpine{{ alpine }}
    - base:alpine{{ alpine }}
  ```
- **Notes**:
  - Labels support Jinja2 expressions. For example, `{{ alpine.split('.')[0] }}` extracts the major version from `alpine`.
  - `tag` argument is provided by `--tag`/`-t` parameter, which reflects the image version.

---

## **Example Configuration**

### Global Settings
```yaml
registry: repo.local
prefix: my-base
maintainer: Awesome Developer <awesome@mail>

images:
  jdk:
    dockerfile: jdk/Dockerfile.jinja2
    variables:
      alpine:
        - "3.19"
        - "3.20"
      java:
        - 11
        - 17
        - 21
    labels:
      - jdk:{{ tag }}-{{ java }}-alpine{{ alpine }}
      - jdk:{{ java }}-alpine{{ alpine.split('.')[0] }}
```

Call it like:

```bash
template-dockerfiles --config build.yaml --tag 1.2.3
```

Which will produce 2x4 -> 8 images, with 16 labels:

```bash
repo.local/my-base/jdk:1.2.3-11-alpine3.19
repo.local/my-base/jdk:11-alpine3
repo.local/my-base/jdk:1.2.3-17-alpine3.19
repo.local/my-base/jdk:17-alpine3
repo.local/my-base/jdk:1.2.3-21-alpine3.19
repo.local/my-base/jdk:21-alpine3
repo.local/my-base/jdk:1.2.3-11-alpine3.20
repo.local/my-base/jdk:11-alpine3
repo.local/my-base/jdk:1.2.3-17-alpine3.20
repo.local/my-base/jdk:17-alpine3
repo.local/my-base/jdk:1.2.3-21-alpine3.20
repo.local/my-base/jdk:21-alpine3
```

Order of values under `variables` block is used to determine the order of labels creation.

---

## **Validation and Recommendations**

### Required Fields
1. `images`: At least one image must be defined with a valid `dockerfile`.

### Optional Enhancements
1. Use the `prefix` field for consistent image organization.
2. Add meaningful labels to enhance discoverability and traceability.
3. Keep in mind that order of variables, determine order of labeling and some labels might overwrite previously created. Use `--dry-run` mode to determine the result.

## **Advanced Tips**

1. **Dynamic Tags**: Use Jinja2 expressions like `{{ os.split('.')[0] }}` to generate tags dynamically.

Parameters
----------

```bash
usage: template-dockerfiles [-h] -c CONFIG_FILE [--dry-run] [--push] [--parallel THREADS] [-v] [--version] -t TAG

A Docker image builder that uses Jinja2 templates to dynamically generate Dockerfiles.

options:
  -h, --help            show this help message and exit
  -c, --config CONFIG_FILE
                        configuration file
  --dry-run             print what would be done, but don't do anything
  --push                push Docker images when successfully build
  --parallel THREADS    specify the number of threads to use (default: number of CPUs).
  -v, --verbose         be verbose
  --version             show the version of the application and exit
  -t, --tag TAG         tag that could be used as an image version

When 'docker build' is just not enough :-)
```
