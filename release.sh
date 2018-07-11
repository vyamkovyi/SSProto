#!/bin/bash

if [ $# -ne 1 ]; then
    echo "Usage: ./release.sh PORT"
    echo "E.g. ./release.sh 48879"
    exit 1
fi

go build --ldflags="-s -w -X main.address=\":$1\""
scp -P 42 ./ssserver hexawolf@doggy:~/server/ssserver
