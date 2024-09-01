#!/bin/bash

set -e
go build -o bin
$MAELSTROM test -w broadcast --bin bin --node-count 5 --time-limit 20 --rate 10 --nemesis partition > /dev/null 2>&1

echo -e "\n-- VALIDITY RESULTS:"
tail -n1 ./store/latest/jepsen.log
