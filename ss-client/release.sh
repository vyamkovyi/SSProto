#!/bin/bash

if [ $# -ne 2 ]; then
    echo "Usage: ./release.sh CERTIFICATE SERVER-ADDRESS"
    echo "E.g. ./release.sh ../server/cert.pem doggoat.de:48879"
    exit 1
fi

certPath="$1"
serverName="$2"

GOOS=linux EXTRABUILDFLAGS=--buildmode=pie ./build.sh $certPath $serverName "Updater"
GOOS=windows ./build.sh $certPath $serverName "Updater.exe"
GOOS=darwin ./build.sh $certPath $serverName "Updater-mac"
