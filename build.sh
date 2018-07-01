#!/bin/sh

if [ $# -ne 3 ]; then
    echo "Usage: ./build.sh KEY-FILE CERTIFICATE SERVER-ADDRESS"
    echo "E.g. ./build.sh ss.key cert.pem doggoat.de"
    echo "Also you can use EXTRABUILDFLAGS envvar to specify additional"
    echo "arguments to pass to go build."
    exit 1
fi

key=$(tail -n 1 "$1")
cert=$(head -n -1 "$2" | tail -n +2 | paste -s -d "")

if command -v openssl &>/dev/null; then
    cn=$(openssl x509 -in "$2" -text -noout | grep -Po "Subject: CN = [^ ]+" | cut -f 2 -d '=' | sed -E 's/^[ ]+([^ ]+)/\1/')
    if [ "$cn" != "$3" ]; then
        echo "CN in certificate ($2) DOES NOT match specified server address ($3). This will not work!"
        exit 2
    fi
fi

go build --ldflags="-s -w -X main.certEnc=$cert -X main.keyEnc=$key -X main.targetHost=$3" $EXTRABUILDFLAGS
