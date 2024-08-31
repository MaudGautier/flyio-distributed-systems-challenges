#!/bin/bash

# Build
go build -o bin

# Run
$MAELSTROM test -w g-counter --bin bin --node-count 3 --rate 100 --time-limit 20 --nemesis partition > /dev/null 2>&1

echo -e "\n-- VALIDITY RESULTS:"
tail -n1 ./store/latest/jepsen.log
