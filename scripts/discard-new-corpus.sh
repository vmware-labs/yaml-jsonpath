#!/usr/bin/env bash
set -euo pipefail

readonly script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

pushd "$script_dir/../pkg/yamlpath/fuzz/corpus"

git clean -f .

popd