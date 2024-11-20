import datetime
import itertools
import logging as log
import os
from collections.abc import Iterable

import git
from jinja2 import Template
from templater import flags
from templater.runner import Runner
from termcolor import colored


def template_file(template_path, params):
    log.debug(f"Templating {template_path} with {params}")
    with open(template_path) as file:
        template = Template(file.read())
        output = template.render(**params)
        log.debug(f"Templated output: \n{output}")
        return output


def flatten(nested_list):
    result = []
    for i in nested_list:
        if isinstance(i, Iterable) and not isinstance(i, str):
            result.extend(flatten(i))  # Recursively flatten and extend the result
        else:
            result.append(i)  # Add non-iterable or string elements directly
    return result


def get_per_image_config_sets(params):
    keys = params["variables"].keys()
    values = params["variables"].values()

    # Generate all combinations and convert them to dictionaries
    return [dict(zip(keys, combo)) for combo in itertools.product(*values)]


def collect_params(config_set, playbook):
    params = config_set.copy()

    registry = playbook.get("registry")
    if registry and registry.strip():
        params.update({"registry": registry})

    prefix = playbook.get("prefix")
    if prefix and prefix.strip():
        params.update({"prefix": prefix})

    return params


def collect_labels(config_set, label_templates):
    labels = []
    args = config_set.copy()

    if flags.TAG:
        args.update({"tag": flags.TAG})
        log.debug(f"Preparing {args}")

    for template in label_templates:
        labels.append(Template(template).render(**args))

    return labels


def get_dockerfile_path(docker_file_template, config_set):
    dirname = os.path.dirname(docker_file_template)
    log.debug(dirname)
    filename = "".join(
        ["-".join([f"{k}-{v}" for k, v in config_set.items()]), ".Dockerfile"]
    )
    log.debug(filename)
    return os.path.join(dirname, filename)


def image_name(*image_parts):
    # ignore registry or prefix if not set
    return "/".join([a for a in image_parts if a is not None])


def execute(playbook):

    images = playbook["images"]
    build_context_dir = playbook["build_context"]

    images_to_push = []

    for image, params in images.items():
        log.debug(f"\n\nProcessing image: {image} {params}")

        builder = Runner()
        temp_files = []

        try:
            template_path = os.path.join(build_context_dir, params["dockerfile"])
            with open(template_path, "r") as file:
                log.debug(f"Reading {template_path}")

        except FileNotFoundError:
            log.error(f"Can't read Dockerfile template for {image}: {template_path}!")

        for config_set in get_per_image_config_sets(params):
            log.debug(f"Current config set: {config_set}")

            image_params = collect_params(config_set, playbook)
            registry = image_params["registry"] if "registry" in image_params else None
            prefix = image_params["prefix"] if "prefix" in image_params else None
            templated_dockerfile = template_file(template_path, image_params)
            labels = collect_labels(config_set, params["labels"])
            dockerfile = get_dockerfile_path(template_path, config_set)
            temp_files.append(dockerfile)  # for later cleanup

            log.debug(f"Creating temporary Dockerfile: {dockerfile}")
            with open(dockerfile, "w") as file:
                file.write(templated_dockerfile)

            build_cmd = flatten(
                [
                    # "echo",
                    "docker",
                    "build",
                    "-f",
                    dockerfile,
                    *[("-t", image_name(registry, prefix, label)) for label in labels],
                    *get_opencontainer_labels(playbook),
                    os.path.dirname(template_path),
                ]
            )
            builder.add(build_cmd)
            log.debug(f"Collecting build command: {build_cmd}")

            collected_images = [image_name(registry, prefix, label) for label in labels]
            images_to_push.extend(collected_images)
            collected_images = "\n".join([f" - {img}" for img in collected_images])
            log.debug(f"Collecting images to push:\n{collected_images}")

        log.info(
            f"{colored('Starting build of image set', 'white')}: {colored(image, 'blue')}"
        )
        builder.run()

        log.debug(f"Removing temporary Dockerfiles: {temp_files}")
        for file in temp_files:
            os.remove(file)

    if flags.DRY_RUN:
        imgs = "\n ".join(images_to_push)
        log.warning(f"DRY-RUN mode, would create:\n {imgs}")

    if flags.PUSH:
        log.info(colored("Pushing images", "white"))
        pusher = Runner()
        for img in images_to_push:
            if flags.LOG_LEVEL < log.INFO:
                pusher.add(["docker", "push", img])
            else:
                pusher.add(["docker", "push", "--quiet", img])

        pusher.run()
        if flags.DRY_RUN:
            imgs = "\n ".join(images_to_push)
            log.warning(f"DRY-RUN mode, would push:\n {imgs}")


def get_opencontainer_labels(playbook):
    labels = []

    maintainer = playbook.get("maintainer")
    if maintainer and maintainer.strip():
        labels.extend(["--label", f"maintainer={maintainer}"])

    if flags.TAG:
        labels.extend(["--label", f"org.opencontainers.image.version={flags.TAG}"])

    repo = git.Repo(search_parent_directories=True)
    try:
        labels.extend(
            ["--label", f"org.opencontainers.image.source={repo.remotes.origin.url}"]
        )
    except AttributeError:
        pass

    try:
        labels.extend(
            ["--label", f"org.opencontainers.image.revision={repo.head.object.hexsha}"]
        )
    except AttributeError:
        pass

    try:
        labels.extend(
            ["--label", f"org.opencontainers.image.branch={repo.active_branch}"]
        )
    except AttributeError:
        pass

    n = datetime.datetime.now(datetime.timezone.utc)
    labels.extend(["--label", f"org.opencontainers.image.created={n.isoformat()}"])

    return labels
