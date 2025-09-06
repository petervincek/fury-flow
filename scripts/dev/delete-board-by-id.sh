#!/bin/sh

if [ -z "$1" ]; then
    echo "Usage: $0 <id>"
    exit 1
fi

curl -v -X DELETE "http://localhost:7070/boards/$1"