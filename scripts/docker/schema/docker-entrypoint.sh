#!/bin/sh

migrate_command="$1"

if [ "$migrate_command" != "up" ] && [ "$migrate_command" != "down" ]; then
  echo "wrong command: expect up or down"
  exit 1
fi

DSN="$PERCONA_DSN"

HOSTPORT=$DSN
HOSTPORT=${HOSTPORT#*(}
HOSTPORT=${HOSTPORT%)*}

HOST=${HOSTPORT%:*}
PORT=${HOSTPORT#*:}

until nc -z -v -w30 "$HOST" "$PORT"
do
  echo "Waiting for database connection..."
  # wait for 5 seconds before check again
  sleep 5
done

migrate \
  -path /migrations \
  -database "$DSN" \
  "$migrate_command"
