import pytest

# import testinfra


class TestSmokes:
    def test_without_arguments(self, host, cli):
        cmd = host.run(f"{cli}")
        assert cmd.failed
        assert "error: the following arguments are required:" in cmd.stderr
        # check if required parameters are reported
        assert "-c/--config" in cmd.stderr
        assert "-t/--tag" in cmd.stderr

    @pytest.mark.parametrize(
        "argument",
        [
            "-h",
            "--help",
        ],
    )
    def test_help(self, host, cli, app_name, argument):
        cmd = host.run(f"{cli} {argument}")
        assert cmd.succeeded
        assert f"usage: {app_name}" in cmd.stdout
        assert "options:" in cmd.stdout
        assert "When 'docker build' is just not enough :-)" in cmd.stdout

    def test_version(self, host, cli, app_name):
        cmd = host.run(f"{cli} --version")
        assert cmd.succeeded
        assert f"{app_name} unknown" in cmd.stdout

    def test_dry_run_normal(self, host, cli):
        cmd = host.run(f"{cli} -c example/build.yaml -t 1.2.3 --dry-run")
        assert cmd.succeeded
        assert "Loading config file: example/build.yaml" in cmd.stdout
        assert "Setting registry to: repo.local" in cmd.stdout
        assert "Setting prefix to: alpine" in cmd.stdout
        assert "Setting build context dir: example" in cmd.stdout
        assert "DRY-RUN mode" in cmd.stdout

    @pytest.mark.parametrize(
        "parallelism",
        [
            "2",
            "max",
        ],
    )
    def test_dry_run_parallel(self, host, cli, parallelism):
        cmd = host.run(
            " ".join(
                [
                    f"{cli}",
                    "-c example/build.yaml",
                    "-t 1.2.3",
                    f"--parallel {parallelism}",
                    "--dry-run",
                ]
            )
        )
        assert cmd.succeeded
        assert "DRY-RUN mode" in cmd.stdout
        assert "Setting parallelism to:" in cmd.stdout

    def test_dry_run_push(self, host, cli):
        cmd = host.run(
            " ".join(
                [f"{cli}", "-c example/build.yaml", "-t 1.2.3", "--push", "--dry-run"]
            )
        )
        assert cmd.succeeded
        assert "DRY-RUN mode" in cmd.stdout
        assert "Pushing images" in cmd.stdout
        assert " - docker push --quiet" in cmd.stdout
