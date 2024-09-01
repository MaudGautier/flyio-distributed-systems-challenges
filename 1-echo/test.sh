#!/bin/bash

set -e
go build -o bin
$MAELSTROM test -w echo --bin bin --node-count 1 --time-limit 10 > /dev/null 2>&1

echo "-- VALIDITY RESULTS:"
tail -n1 ./store/latest/jepsen.log
