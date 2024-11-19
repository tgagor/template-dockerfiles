import logging
import sys

from termcolor import colored


def init(log_level):
    # logging.basicConfig(
    #     level=args['LOG_LEVEL'],
    #     format="%(levelname)s: %(message)s"
    #     # format="%(levelname)s:%(funcName)s: %(message)s"
    # )

    fmt = MyFormatter()
    handler = logging.StreamHandler(sys.stdout)

    handler.setFormatter(fmt)
    logging.root.addHandler(handler)
    logging.root.setLevel(log_level)


class MyFormatter(logging.Formatter):

    error_fmt = f"{colored('%(levelname)s', 'red')}: %(message)s"
    warn_fmt = f"{colored('%(levelname)s', 'yellow')}: %(message)s"
    debug_fmt = "".join(
        [
            colored("%(levelname)s", "blue"),
            ":",
            colored("%(funcName)s", "cyan"),
            ":line",
            colored("%(lineno)d", "yellow"),
            ": %(message)s",
        ]
    )
    info_fmt = "%(message)s"

    def __init__(self):
        super().__init__(fmt="%(levelno)d: %(msg)s", datefmt=None, style="%")

    def format(self, record):

        # Save the original format configured by the user
        # when the logger formatter was instantiated
        format_orig = self._style._fmt

        # Replace the original format with one customized by logging level
        if record.levelno == logging.DEBUG:
            self._style._fmt = MyFormatter.debug_fmt

        elif record.levelno == logging.INFO:
            self._style._fmt = MyFormatter.info_fmt

        elif record.levelno == logging.WARNING:
            self._style._fmt = MyFormatter.warn_fmt

        elif record.levelno == logging.ERROR:
            self._style._fmt = MyFormatter.error_fmt

        # Call the original formatter class to do the grunt work
        result = logging.Formatter.format(self, record)

        # Restore the original format configured by the user
        self._style._fmt = format_orig

        return result
