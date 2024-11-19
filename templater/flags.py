import argparse
import logging as log
import sys

import templater

parser = None
args = None


def parse(argv=None):
    global parser
    global args

    parser = argparse.ArgumentParser(
        prog="template-dockerfiles",
        description="Dockerfile templater and image builder",
        epilog="When 'docker build' is just not enough :-)",
    )
    parser.add_argument(
        "-c",
        "--config",
        action="store",
        dest="CONFIG_FILE",
        help="configuration file",
        required=True,
        default="build.yaml",
    )

    parser.add_argument(
        "--dry-run",
        help="Print what would be done, but don't do anything",
        action="store_const",
        dest="DRY_RUN",
        const=True,
        default=False,
    )

    parser.add_argument(
        "--push",
        help="Push Docker images when successfully build",
        action="store_const",
        dest="PUSH",
        const=True,
        default=False,
    )

    parser.add_argument(
        "--parallel",
        help="Specify the number of threads to use (default: number of CPUs).",
        type=validate_threads,
        dest="THREADS",
        default=1,
    )

    parser.add_argument(
        "-v",
        "--verbose",
        help="Be verbose",
        action="store_const",
        dest="LOG_LEVEL",
        const=log.DEBUG,
        default=log.INFO,
    )

    parser.add_argument(
        "--version",
        help="Show the version of the application and exit",
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
        help="Tag that could be used as an image version",
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
