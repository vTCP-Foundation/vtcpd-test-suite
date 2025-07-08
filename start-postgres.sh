#!/bin/bash
# start-postgres.sh
# Starts the PostgreSQL server as the postgres user.
set -eu

PGDATA="/var/lib/postgresql/data"

# The entrypoint script already runs initdb which creates the cluster.
# This script's only job is to start the server.

# Log file is now specified here, ensuring correct permissions.
LOGFILE="$PGDATA/pg_server.log"
touch "$LOGFILE"
chown postgres:postgres "$LOGFILE"

echo "[start-postgres] Starting PostgreSQL server..."
pg_ctl -D "$PGDATA" -l "$LOGFILE" -o "-c listen_addresses='*'" start

# Give it a moment to start up and log potential issues.
sleep 2
echo "[start-postgres] Server startup command issued. Check logs at '$LOGFILE' inside the container."
cat "$LOGFILE" 