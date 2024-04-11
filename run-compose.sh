#!/usr/bin/env bash
mkdir -p volume

if [ ! -d "$PWD/volume/postgresql" ]; then
   mkdir -p volume/postgresql/data
   chmod 755 -R $PWD/volume/postgresql/
fi

if [ -z $1 ]; then
  chmod 755 -R $PWD/volume/
  docker compose -f docker-compose.yaml up -d
elif [[ $1 = "up" ]]; then
  chmod 755 -R $PWD/volume/
elif [[ $1 = "down" ]]; then
  docker compose -f docker-compose.yaml down
fi