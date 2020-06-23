#!/usr/bin/env bash
set -euo pipefail

readonly script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

pushd "$script_dir/../web"

gcloud app deploy --version=$(git rev-parse --short HEAD) --quiet

popd