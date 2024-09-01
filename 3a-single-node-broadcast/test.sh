#!/bin/bash

set -e
go build -o bin
$MAELSTROM test -w broadcast --bin bin --node-count 1 --time-limit 20 --rate 10 > /dev/null 2>&1

echo "-- VALIDITY RESULTS:"
tail -n1 ./store/latest/jepsen.log
