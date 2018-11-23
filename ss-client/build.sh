#!/bin/sh

if [ $# -ne 3 ]; then
    echo "Usage: ./build.sh CERTIFICATE SERVER-ADDRESS FILENAME"
    echo "E.g. ./build.sh cert.pem doggoat.de:48879 Updater"
    echo "Also you can use EXTRABUILDFLAGS envvar to specify additional"
    echo "arguments to pass to go build."
    exit 1
fi

cert=$(printf "%s" "$(< $1)" | head -n -1 | tail -n +2 | paste -s -d "")
timestamp=$(LC_ALL=C date -u +"%a, %d %b %Y %H:%M:%S GMT")

echo "$3: Build timestamp: $timestamp"

go build -o $3 --ldflags="-s -w -X \"main.buildStamp=$timestamp\" -X main.certEnc=$cert -X main.targetHost=$2" $EXTRABUILDFLAGS
if [ $? -eq 0 ]; then
    touch -c -d "$timestamp" $3
    echo "$3: Make sure file modification time on server matches the value above!"
fi
