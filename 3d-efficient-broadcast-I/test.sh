#!/bin/bash

# Build
go build -o bin

# With nemesis
$MAELSTROM test -w broadcast --bin bin --node-count 25 --time-limit 20 --rate 100 --latency 100 --nemesis partition > /dev/null 2>&1

echo -e "\n-- VALIDITY RESULTS WITH NEMESIS:"
tail -n1 ./store/latest/jepsen.log


# Without nemesis - checking performance
$MAELSTROM test -w broadcast --bin bin --node-count 25 --time-limit 20 --rate 100 --latency 100 > /dev/null 2>&1

echo -e "\n-- VALIDITY RESULTS:"
tail -n1 ./store/latest/jepsen.log
echo -e "\n-- MESSAGE PER OPS RESULTS:"
grep -A 9 ".net" ./store/latest/results.edn
echo -e "\n-- LATENCY RESULTS:"
grep -A 4 ":stable-latencies" ./store/latest/results.edn

echo -e "\n-- EXPECTATIONS:"
echo -e "   - msgs-per-ops   < 30"
echo -e "   - median latency < 400ms"
echo -e "   - max latency    < 600ms"


