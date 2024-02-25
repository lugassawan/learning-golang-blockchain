#!/bin/sh

id=${1:-1}
compile=${2:-false}
default_id=0

if [ $id -lt $default_id ]; then
  id=1
fi

export NODE_ID=$id

read -p "Command : " args

if [ "$compile" = true ]; then
  echo "Compiling...\n"
  go build && ./learning-golang-blockchain $args
else
  go run . $args
fi