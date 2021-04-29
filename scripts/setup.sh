#!/usr/bin/env bash
DKCM="docker-compose -p lmstfy"

cd scripts/docker && $DKCM up -d --remove-orphans && cd ../..

curl -q -s -XPOST -d "description=test&token=01F4CKPFY8XYVH6WVEQB3747CW" "http://127.0.0.1:7778/token/test-ns" >> /dev/null