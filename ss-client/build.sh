#!/bin/sh

if [ $# -ne 3 ]; then
    echo "Usage: ./build.sh CERTIFICATE SERVER-ADDRESS FILENAME"
    echo "E.g. ./build.sh cert.pem doggoat.de Updater"
    echo "Also you can use EXTRABUILDFLAGS envvar to specify additional"
    echo "arguments to pass to go build."
    exit 1
fi

cert=$(head -n -1 "$1" | tail -n +2 | paste -s -d "")

go build -o $3 --ldflags="-s -w -X main.certEnc=$cert -X main.targetHost=$2" $EXTRABUILDFLAGS
