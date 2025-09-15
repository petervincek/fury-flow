#!/bin/sh

curl -v -X POST -H "Content-Type: application/json" -d '{"boardName": "New Board", "description": "Board description", "createdBy": "Peter (curl)"}' http://localhost:7070/boards 