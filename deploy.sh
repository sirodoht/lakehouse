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
    make test

    # push origin srht
    git push -v origin master

    # push on github
    git push -v github master

    # pull on server and reload
    #ssh deploy@5.75.194.9 'cd /var/www/lakehouse ' \
    #    '&& git pull ' \
    #    '&& source ~/.profile && make build ' \
    #    '&& sudo systemctl restart lakehouse-web'
}

main "$@"
