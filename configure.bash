#!/usr/bin/env bash

first_names=( 'Lucas' 'Nahuel' 'Sofia' 'María Victoria' 'Jorge' )
last_names=( 'Villanueva' 'Pérez' 'Ramos' 'Sanchez' 'Santos' )
documents=( 40382057 33956394 38291054 25183928 30392849 )
birthdates=( '1997-01-30' '1990-12-10' '1995-10-08' '1977-04-28' '1986-08-14' )
numbers=( 9713 5784 1230 2744 7574 )
nclients=${1:-5}
name=$0

is_nonnegative_int() {
    printf %d -"$1" >/dev/null 2>&1
}

error() {
    >&2 printf 'usage: %s [nclients]\n  %s\n' "$name" "$1"
    exit
}

if ! is_nonnegative_int $nclients || [ $nclients -gt 5 ]
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
      - testing_net
    volumes:
      - ./server/config.ini:/config.ini"

for ((i = 0, n = 1; i < nclients; ++i, ++n))
do
    printf "
  client%d:
    container_name: client%d
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_AGENCY=%d
      - CLI_LOG_LEVEL=DEBUG
      - NOMBRE=%s
      - APELLIDO=%s
      - DOCUMENTO=%s
      - NACIMIENTO=%s
      - NUMERO=%d
    networks:
      - testing_net
    depends_on:
      - server
    volumes:
      - ./client/config.yaml:/config.yaml\n" $n $n $n \
          "${first_names[$i]}" \
          "${last_names[$i]}" \
          "${documents[$i]}" \
          "${birthdates[$i]}" \
          "${numbers[$i]}"
done

echo '
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24'
