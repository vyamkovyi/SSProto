#!/bin/bash

if [ $# -ne 3 ]; then
    echo "Usage: ./release.sh KEY-FILE CERTIFICATE SERVER-ADDRESS"
    echo "E.g. ./release.sh ../server/ss.key ../server/cert.pem doggoat.de:48879"
    exit 1
fi

keyPath="$1"
certPath="$2"
serverName="$3"

EXTRABUILDFLAGS=--buildmode=pie ./build.sh $keyPath $certPath $serverName
GOOS=windows ./build.sh $keyPath $certPath $serverName
mv ssclient Updater
mv ssclient.exe Updater.exe
