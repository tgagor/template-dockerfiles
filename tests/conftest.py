import pytest
import testinfra

APP = "template-dockerfiles"
CLI = f"poetry run {APP}"


@pytest.fixture
def host():
    return testinfra.host.get_host("local://")


@pytest.fixture
def cli():
    return CLI


@pytest.fixture
def app_name():
    return APP
