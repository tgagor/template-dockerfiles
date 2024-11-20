import argparse
import logging as log
import os
import sys

import templater

parser = None
args = None


def parse(argv=None):
    global parser
    global args

    parser = argparse.ArgumentParser(
        prog="template-dockerfiles",
        description="A Docker image builder that uses Jinja2 templates to dynamically generate Dockerfiles.",
        epilog="When 'docker build' is just not enough :-)",
    )
    parser.add_argument(
        "-c",
        "--config",
        action="store",
        dest="CONFIG_FILE",
        help="path to the configuration file",
        type=is_valid_file,
        required=True,
    )

    parser.add_argument(
        "--dry-run",
        help="print what would be done, but don't do anything",
        action="store_const",
        dest="DRY_RUN",
        const=True,
        default=False,
    )

    parser.add_argument(
        "--push",
        help="push Docker images when successfully build",
        action="store_const",
        dest="PUSH",
        const=True,
        default=False,
    )

    parser.add_argument(
        "--parallel",
        help="specify the number of threads to use (default: number of CPUs).",
        type=validate_threads,
        dest="THREADS",
        default=1,
    )

    parser.add_argument(
        "-v",
        "--verbose",
        help="be verbose",
        action="store_const",
        dest="LOG_LEVEL",
        const=log.DEBUG,
        default=log.INFO,
    )

    parser.add_argument(
        "--version",
        help="show the version of the application and exit",
        action="version",
        version=f"%(prog)s {templater.__version__}",
    )

    # parser.add_argument('--debug',
    #                     help="Print lots of debugging statements",
    #                     action="store_const",
    #                     dest="LOG_LEVEL",
    #                     const=log.DEBUG)

    parser.add_argument(
        "-t",
        "--tag",
        action="store",
        dest="TAG",
        help="tag that could be used as an image version",
        required=True,
        default=None,
    )

    # parse only once
    if args is None:
        args = vars(parser.parse_args(argv or sys.argv[1:]))

    return args


def validate_threads(value):
    """Validate that the provided thread count is a positive integer."""
    if value.lower() == "max":
        return "max"
    try:
        threads = int(value)
        if threads < 1:
            raise ValueError
        return threads
    except ValueError:
        raise argparse.ArgumentTypeError(
            f"Invalid value for threads: {value}. Must be 'max' or a positive integer."
        )


def is_valid_file(file_path):
    if not os.path.isfile(file_path):
        raise argparse.ArgumentTypeError(
            f"The file '{file_path}' does not exist or is not a valid file."
        )
    return file_path


def __getattr__(name):
    global args
    if args is None:
        raise RuntimeError("Arguments have not been parsed yet.")
    if name in args.keys():
        return args[name]
    raise AttributeError(f"module '{__name__}' has no attribute '{name}'")


def all():
    global args
    if args is None:
        raise RuntimeError("Arguments have not been parsed yet.")
    return args


def add(name, value):
    global args
    args.update({name: value})
