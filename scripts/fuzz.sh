#!/usr/bin/env bash
set -euo pipefail

readonly script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

pushd "$script_dir/../pkg/yamlpath/fuzz"

go-fuzz-build

go-fuzz -procs 20

popd

git checkout go.mod go.sum