import logging as log
import os

import yaml
from templater import flags


def read(config_file):
    log.debug(f"Opening config file: {config_file}")
    with open(config_file, "r") as file:
        log.info(f"Loading config file: {config_file}")
        playbook = yaml.safe_load(file)
        log.debug(f"Config loaded: \n{yaml.dump(playbook)}")
        return playbook


def validate():
    """Validate the configuration file and provide suggestions for missing or incorrect arguments."""
    registry = None
    prefix = None
    images = []

    config_file = flags.CONFIG_FILE
    playbook = read(config_file)

    # Validate 'registry'
    try:
        registry = playbook["registry"]
        if not isinstance(registry, str) or not registry.strip():
            raise ValueError(
                f"'registry' must be a non-empty string. Found: {registry}"
            )
        log.info(f"Setting registry to: {registry}")
    except KeyError:
        log.warning(
            "'registry' is not set. Consider adding it for correct configuration."
        )
    except ValueError as e:
        log.error(str(e))
        raise

    # Validate 'prefix'
    try:
        prefix = playbook["prefix"]
        if not isinstance(prefix, str) or not prefix.strip():
            raise ValueError(f"'prefix' must be a non-empty string. Found: {prefix}")
        log.info(f"Setting prefix to: {prefix}")
    except KeyError:
        log.warning("'prefix' is not set. Consider adding it for better organization.")
    except ValueError as e:
        log.error(str(e))
        raise

    # Validate 'images'
    try:
        images = playbook["images"]
        if not isinstance(images, dict):
            raise ValueError(
                f"'images' must be a dictionary. Found: {type(images).__name__}"
            )
        log.debug(f"Reading images configuration: {images}")
    except KeyError:
        log.error(
            "'images' is not set. You need to define at least one image configuration."
        )
        raise
    except ValueError as e:
        log.error(str(e))
        raise

    # Validate individual images
    for image_name, image_config in images.items():
        if not isinstance(image_config, dict):
            log.error(
                f"Image '{image_name}' must have a dictionary configuration. Found: {type(image_config).__name__}"
            )
            continue

        # Check for 'dockerfile'
        if "dockerfile" not in image_config:
            log.error(
                f"Image '{image_name}' is missing the required 'dockerfile' field."
            )
        elif (
            not isinstance(image_config["dockerfile"], str)
            or not image_config["dockerfile"].strip()
        ):
            log.error(f"Image '{image_name}': 'dockerfile' must be a non-empty string.")

        # Check for 'variables'
        if "variables" in image_config:
            variables = image_config["variables"]
            if not isinstance(variables, dict):
                log.error(
                    f"Image '{image_name}': 'variables' must be a dictionary. Found: {type(variables).__name__}"
                )
            else:
                # Ensure all variables have valid values
                for var_name, var_values in variables.items():
                    if not isinstance(var_values, list) or not all(
                        isinstance(v, (int, str)) for v in var_values
                    ):
                        log.error(
                            f"Image '{image_name}': Variable '{var_name}' must have a list of integers or strings."
                        )
        else:
            log.warning(
                f"Image '{image_name}' does not define 'variables'. Consider adding it for customization."
            )

        # Check for 'labels'
        if "labels" in image_config:
            labels = image_config["labels"]
            if not isinstance(labels, list) or not all(
                isinstance(label, str) for label in labels
            ):
                log.error(f"Image '{image_name}': 'labels' must be a list of strings.")
        else:
            log.warning(
                f"Image '{image_name}' does not define 'labels'. Consider adding it for better tagging."
            )

    # Set build context directory
    build_context_dir = os.path.dirname(config_file)
    flags.add("build_context", build_context_dir)
    log.info(f"Setting build context dir: {build_context_dir}")

    # Add build context to playbook
    playbook.update({"build_context": build_context_dir})

    return playbook
