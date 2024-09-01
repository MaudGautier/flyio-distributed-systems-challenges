#!/bin/bash

set -e
go build -o bin
$MAELSTROM test -w unique-ids --bin bin --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition > /dev/null 2>&1

echo "-- VALIDITY RESULTS:"
tail -n1 ./store/latest/jepsen.log
