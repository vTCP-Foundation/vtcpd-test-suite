#!/usr/bin/env bash
# docker-entrypoint-initdb.sh
# Idempotent initialization of PostgreSQL for vtcpd test containers.
set -euo pipefail

# Predefine variables to avoid unbound errors under 'set -u'
INITDB_BIN=""
PG_CTL_BIN=""

FLAG_FILE="/var/lib/postgresql/data/.initialized"

# Exit early if database was already initialized on a previous container start
if [ -f "$FLAG_FILE" ]; then
  echo "[initdb] PostgreSQL already initialized, skipping setup."
  exit 0
fi

# Ensure postgres user exists
if ! id "postgres" &>/dev/null; then
    # In some minimal images, useradd is in /usr/sbin
    PATH=$PATH:/usr/sbin useradd -r -s /bin/bash postgres
fi

# Create directory for socket
mkdir -p /var/run/postgresql
chown -R postgres:postgres /var/run/postgresql
chmod 775 /var/run/postgresql

# Ensure PostgreSQL data directory exists and has correct ownership
PGDATA="/var/lib/postgresql/data"
mkdir -p "$PGDATA"
chown -R postgres:postgres "$PGDATA"
chmod 700 "$PGDATA"

# Run initdb as postgres user
if [ -z "$(ls -A "$PGDATA")" ]; then
  echo "[initdb] Running initdb to create cluster..."
  if command -v runuser &>/dev/null; then
    runuser -u postgres -- initdb -D "$PGDATA" -U postgres
  else
    su - postgres -c "initdb -D \"$PGDATA\" -U postgres"
  fi
fi

# Start PostgreSQL temporarily (detached) as postgres user
if command -v runuser &>/dev/null; then
  runuser -u postgres -- pg_ctl -D "$PGDATA" -o "-c listen_addresses='127.0.0.1'" -w start
else
  su - postgres -c "pg_ctl -D \"$PGDATA\" -o \"-c listen_addresses='127.0.0.1'\" -w start"
fi

# Create user and database
if command -v runuser &>/dev/null; then
    runuser -u postgres -- psql -v ON_ERROR_STOP=1 <<'EOSQL'
DO
$$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'vtcpd_user') THEN
      CREATE ROLE vtcpd_user LOGIN PASSWORD 'vtcpd_pass';
   END IF;
END
$$;
EOSQL
    runuser -u postgres -- psql -v ON_ERROR_STOP=1 -c "CREATE DATABASE storagedb OWNER vtcpd_user;"
else
    su - postgres -c "psql -v ON_ERROR_STOP=1" <<'EOSQL'
DO
$$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'vtcpd_user') THEN
      CREATE ROLE vtcpd_user LOGIN PASSWORD 'vtcpd_pass';
   END IF;
END
$$;
EOSQL
    su - postgres -c "psql -v ON_ERROR_STOP=1 -c 'CREATE DATABASE storagedb OWNER vtcpd_user;'"
fi

# Stop PostgreSQL instance that we started
if command -v runuser &>/dev/null; then
  runuser -u postgres -- pg_ctl -D "$PGDATA" -m fast stop
else
  su - postgres -c "pg_ctl -D \"$PGDATA\" -m fast stop"
fi

echo "[initdb] PostgreSQL initialisation complete."

touch "$FLAG_FILE"
chown postgres:postgres "$FLAG_FILE" 