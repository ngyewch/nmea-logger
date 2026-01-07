#!/usr/bin/env bash

INPUT_FILE=$1

if [ -z "${INPUT_FILE}" ]; then
  echo "Input file not specified."
  exit 1
fi

docker run --rm -it \
    -v $(realpath ${INPUT_FILE}):/data/$(basename ${INPUT_FILE}) \
    -p 8080:8080 \
    nmea-logger:latest \
    -- \
    ais view /data/$(basename ${INPUT_FILE})
