#!/bin/sh

n=1
nclients=${1:-1}
name=$0

is_nonnegative_int() {
    printf %d -"$1" >/dev/null 2>&1
}

error() {
    >&2 printf 'usage: %s [nclients]\n  %s\n' "$name" "$1"
    exit
}

if ! is_nonnegative_int $nclients
then
    error "nclients must be a non-negative integer"
fi

echo "version: '3.9'
name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net"

while [ $n -le $nclients ]
do
    printf '
  client%d:
    container_name: client%d
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=%d
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server\n' $n $n $n
    n=$((n+1))
done

echo '
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24'
