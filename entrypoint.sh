#!/bin/bash

set -e

umask 000
if [ "$1" = 'server' ]; then
	/data-provider-server
else
    exec "$@"
fi

