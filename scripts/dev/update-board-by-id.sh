#!/bin/sh

if [ -z "$1" ]; then
    echo "Usage: $0 <id>"
    exit 1
fi

curl -v -X PUT "http://localhost:7070/boards/$1" \
    -H "Content-Type: application/json" \
    -d '{"boardName": "Updated Board", "description": "Board description (updated)", "updatedBy": "John (curl)"}'