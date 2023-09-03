#!/bin/sh

NC_SERVER_ADDRESS="${NC_SERVER_ADDRESS:-server:12345}"
MESSAGE='Testing EchoServer...'

server="${NC_SERVER_ADDRESS%:*}"
port=${NC_SERVER_ADDRESS#*:}

docker run --rm --network tp0_testing_net alpine \
    sh -c "apk add --update --no-cache netcat-openbsd ;
           echo '$MESSAGE' | nc '$server' '$port'"
