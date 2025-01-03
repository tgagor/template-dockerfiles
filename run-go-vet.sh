#!/usr/bin/env bash
set -e
pkg=github.com/tgagor/template-dockerfiles
for dir in ; do
  go vet /
done
