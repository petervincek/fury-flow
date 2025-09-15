#!/bin/sh
set -e

# Run migrations
./goose up

# Start the API server
./fury-flow