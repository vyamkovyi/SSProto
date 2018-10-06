#!/bin/sh

if [ $# -ne 4 ]; then
    echo "Usage: ./build.sh KEY-FILE CERTIFICATE SERVER-ADDRESS FILENAME"
    echo "E.g. ./build.sh ss.key cert.pem doggoat.de Updater"
    echo "Also you can use EXTRABUILDFLAGS envvar to specify additional"
    echo "arguments to pass to go build."
    exit 1
fi

key=$(tail -n 1 "$1")
cert=$(head -n -1 "$2" | tail -n +2 | paste -s -d "")

go build -o $4 --ldflags="-s -w -X main.certEnc=$cert -X main.keyEnc=$key -X main.targetHost=$3" $EXTRABUILDFLAGS
