#!/usr/bin/env bash

curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ./bin

./bin/just b

./bin/qosproxy -h

