#!/bin/bash

go build -o bin
$MAELSTROM test -w echo --bin bin --node-count 1 --time-limit 10
