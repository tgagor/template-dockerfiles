#!/usr/bin/env python3

import logging as log
import pprint

from templater import config, flags, logger, parser


def main():
    flags.parse()
    logger.init(flags.LOG_LEVEL)
    log.debug(f"Parsed arguments: {pprint.saferepr(flags.all())}")

    playbook = config.validate()
    parser.execute(playbook)


if __name__ == "__main__":
    main()
