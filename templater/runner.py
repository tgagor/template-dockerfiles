import concurrent.futures
import logging as log
import os
import subprocess
import sys

from templater import flags


class Runner:
    def __init__(self):
        # Initialize an empty list to store commands
        self.commands = []

    def add(self, cmd):
        # Add a command to the list (each command is a list of arguments for subprocess)
        self.commands.append(cmd)

    def run(self):
        # Get the number of available CPU cores
        max_workers = os.cpu_count() if flags.THREADS == "max" else flags.THREADS
        if max_workers > 1:
            log.info(f"Setting parallelism to: {max_workers}")

        if flags.DRY_RUN:
            cmds = [" ".join(cmd) for cmd in self.commands]
            cmds = "\n".join([f" - {cmd}" for cmd in cmds])
            log.debug(f"DRY-RUN mode, normally would execute:\n{cmds}")
            return

        # Run all commands in parallel
        with concurrent.futures.ProcessPoolExecutor(
            max_workers=max_workers
        ) as executor:
            # Submit each command as a separate job
            futures = [
                executor.submit(self._execute_command, cmd) for cmd in self.commands
            ]

            # Collect results (this will also raise exceptions if any command fails)
            for future in concurrent.futures.as_completed(futures):
                try:
                    future.result()
                except Exception as e:
                    log.error(f"Command failed with error: {e}")
                    sys.exit(2)

    # def _execute_command(self, cmd):
    #     # Run the command using subprocess and capture the output
    #     result = subprocess.run(cmd, capture_output=True, text=True)

    #     # Raise an exception if the command fails
    #     result.check_returncode()

    #     # Return the command output (stdout)
    #     return result.stdout
    def _execute_command(self, cmd):
        # Run the command using subprocess.Popen and stream the output
        process = subprocess.Popen(
            cmd,
            stdout=sys.stdout,  # Direct stdout to the console
            stderr=sys.stderr,  # Direct stderr to the console
            text=True,  # Enable text mode for streaming output
        )

        # Wait for the process to complete and check the return code
        process.communicate()  # This will block until the process finishes
        if process.returncode != 0:
            raise subprocess.CalledProcessError(returncode=process.returncode, cmd=cmd)
