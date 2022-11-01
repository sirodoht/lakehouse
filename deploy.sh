#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
if [[ "${TRACE-0}" == "1" ]]; then
    set -o xtrace
fi

if [[ "${1-}" =~ ^-*h(elp)?$ ]]; then
    echo 'Usage: ./deploy.sh

This script deploys the service in the production server.'
    exit
fi

cd "$(dirname "$0")"

main() {
    # make sure linting checks pass
    make lint

    # make sure tests pass
    go test -v ./...

    # push origin srht
    git push -v origin master

    # push on github
    git push -v github master
}

main "$@"
