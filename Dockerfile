###################################################################################
# Manjaro Linux runtime environment
FROM manjarolinux/base AS runtime-manjaro
RUN pacman -Syu --noconfirm && \
    pacman -S --noconfirm \
    boost \
    boost-libs \
    libsodium \
    postgresql-libs \
    postgresql \
    gcc \
    sqlite && \
    # Debug library locations
    echo "Library locations:" && \
    ls -la /usr/lib/libboost_* && \
    ls -la /usr/lib/libsodium* && \
    echo "Library dependencies:" && \
    ldd /vtcpd/vtcpd || true

# Create a non-root user for running the daemon
RUN useradd -r -s /usr/bin/nologin vtcpd && \
    # Create postgres user for database operations
    useradd -r -s /bin/bash postgres || true


###################################################################################
# Ubuntu runtime environment
FROM ubuntu:24.04 AS runtime-ubuntu
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    libboost-system1.83.0 \
    libboost-filesystem1.83.0 \
    libboost-program-options1.83.0 \
    libsodium23 \
    libpq5 \
    postgresql-16 \
    postgresql-client-16 \
    libasan8 \
    vim \
    sqlite3 && \
    ln -s /usr/lib/postgresql/*/bin/pg_ctl /usr/local/bin/pg_ctl && \
    ln -s /usr/lib/postgresql/*/bin/initdb /usr/local/bin/initdb && \
    rm -rf /var/lib/apt/lists/* && \
    useradd -r -s /usr/sbin/nologin vtcpd && \
    # Create postgres user for database operations
    useradd -r -s /bin/bash postgres || true


###################################################################################
# PostgreSQL pre-initialization stage for Ubuntu
FROM runtime-ubuntu AS postgres-init-ubuntu

# Create PostgreSQL data directory
RUN mkdir -p /var/lib/postgresql/data && \
    chown -R postgres:postgres /var/lib/postgresql && \
    chmod 700 /var/lib/postgresql/data

# Create run directory for PostgreSQL socket
RUN mkdir -p /var/run/postgresql && \
    chown -R postgres:postgres /var/run/postgresql && \
    chmod 775 /var/run/postgresql

# Initialize PostgreSQL database cluster as postgres user
RUN su - postgres -c "initdb -D /var/lib/postgresql/data -U postgres"

# Start PostgreSQL temporarily, create user and database, then stop
RUN su - postgres -c "pg_ctl -D /var/lib/postgresql/data -o '-c listen_addresses=127.0.0.1' -w start" && \
    su - postgres -c "psql -v ON_ERROR_STOP=1 -c \"DO \\\$\\\$ BEGIN IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'vtcpd_user') THEN CREATE ROLE vtcpd_user LOGIN PASSWORD 'vtcpd_pass'; END IF; END \\\$\\\$;\"" && \
    su - postgres -c "psql -v ON_ERROR_STOP=1 -c 'CREATE DATABASE storagedb OWNER vtcpd_user;'" && \
    su - postgres -c "pg_ctl -D /var/lib/postgresql/data -m fast stop"

# Create flag file to indicate database is initialized
RUN touch /var/lib/postgresql/data/.initialized && \
    chown postgres:postgres /var/lib/postgresql/data/.initialized


###################################################################################
# PostgreSQL pre-initialization stage for Manjaro
FROM runtime-manjaro AS postgres-init-manjaro

# Create PostgreSQL data directory
RUN mkdir -p /var/lib/postgresql/data && \
    chown -R postgres:postgres /var/lib/postgresql && \
    chmod 700 /var/lib/postgresql/data

# Create run directory for PostgreSQL socket
RUN mkdir -p /var/run/postgresql && \
    chown -R postgres:postgres /var/run/postgresql && \
    chmod 775 /var/run/postgresql

# Initialize PostgreSQL database cluster as postgres user
RUN su - postgres -c "initdb -D /var/lib/postgresql/data -U postgres"

# Start PostgreSQL temporarily, create user and database, then stop
RUN su - postgres -c "pg_ctl -D /var/lib/postgresql/data -o '-c listen_addresses=127.0.0.1' -w start" && \
    su - postgres -c "psql -v ON_ERROR_STOP=1 -c \"DO \\\$\\\$ BEGIN IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'vtcpd_user') THEN CREATE ROLE vtcpd_user LOGIN PASSWORD 'vtcpd_pass'; END IF; END \\\$\\\$;\"" && \
    su - postgres -c "psql -v ON_ERROR_STOP=1 -c 'CREATE DATABASE storagedb OWNER vtcpd_user;'" && \
    su - postgres -c "pg_ctl -D /var/lib/postgresql/data -m fast stop"

# Create flag file to indicate database is initialized
RUN touch /var/lib/postgresql/data/.initialized && \
    chown postgres:postgres /var/lib/postgresql/data/.initialized


###################################################################################
# Final stage for Manjaro Linux
FROM runtime-manjaro AS final-manjaro

# Copy pre-initialized PostgreSQL data from init stage
COPY --from=postgres-init-manjaro --chown=postgres:postgres /var/lib/postgresql/data /var/lib/postgresql/data
# Fix permissions for PostgreSQL data directory
RUN chmod 700 /var/lib/postgresql/data

# Create vtcpd directory and set permissions
RUN mkdir -p /vtcpd
WORKDIR /vtcpd
RUN mkdir -p vtcpd

# Copy pre-built binaries from host
COPY ./deps/vtcpd /vtcpd/
COPY ./deps/cli/cli /vtcpd/cli
COPY ./start-postgres.sh /usr/local/bin/start-postgres.sh
RUN chmod +x /usr/local/bin/start-postgres.sh

# Create run directory for PostgreSQL socket
RUN mkdir -p /var/run/postgresql && \
    chown -R postgres:postgres /var/run/postgresql && \
    chmod 775 /var/run/postgresql

# Define build arguments and convert them to runtime environment variables
ARG VTCPD_LISTEN_ADDRESS=127.0.0.1
ARG VTCPD_LISTEN_PORT=2000
ARG VTCPD_EQUIVALENTS_REGISTRY=eth
ARG VTCPD_MAX_HOPS=5
ARG CLI_LISTEN_ADDRESS=127.0.0.1
ARG CLI_LISTEN_PORT=3000
ARG CLI_LISTEN_PORT_TESTING=3001
ARG VTCPD_DATABASE_CONFIG=sqlite3:///io

ENV VTCPD_LISTEN_ADDRESS=${VTCPD_LISTEN_ADDRESS}
ENV VTCPD_LISTEN_PORT=${VTCPD_LISTEN_PORT}
ENV VTCPD_EQUIVALENTS_REGISTRY=${VTCPD_EQUIVALENTS_REGISTRY}
ENV VTCPD_MAX_HOPS=${VTCPD_MAX_HOPS}
ENV CLI_LISTEN_ADDRESS=${CLI_LISTEN_ADDRESS}
ENV CLI_LISTEN_PORT=${CLI_LISTEN_PORT}
ENV CLI_LISTEN_PORT_TESTING=${CLI_LISTEN_PORT_TESTING}
ENV VTCPD_DATABASE_CONFIG=${VTCPD_DATABASE_CONFIG}

# Create startup script that uses runtime environment variables
RUN echo '#!/bin/bash\n\
cd /vtcp\n\
# Create vtcpd config file
cat <<EOF > /vtcp/vtcpd/conf.json\n\
{\n\
  "addresses": [\n\
    {\n\
      "type": "ipv4",\n\
      "address": "${VTCPD_LISTEN_ADDRESS}:${VTCPD_LISTEN_PORT}"\n\
    }\n\
  ],\n\
  "database_config": "${VTCPD_DATABASE_CONFIG}",\n\
  "equivalents_registry_address": "${VTCPD_EQUIVALENTS_REGISTRY}",\n\
  "max_hops_count": ${VTCPD_MAX_HOPS}\n\
}\n\
EOF\n\
# Create cli config file
    cat <<EOF > /vtcp/conf.yaml\n\
workdir: "/vtcp/vtcpd/"\n\
vtcpd_path: "/vtcp/vtcpd/vtcpd"\n\
http:\n\
  host: "${CLI_LISTEN_ADDRESS}"\n\
  port: ${CLI_LISTEN_PORT}\n\
http_testing:\n\
  host: "${CLI_LISTEN_ADDRESS}"\n\
  port: ${CLI_LISTEN_PORT_TESTING}\n\
EOF\n\
# Start PostgreSQL if configured (database already initialized)\n\
if [[ "${VTCPD_DATABASE_CONFIG}" == *"postgresql"* ]]; then\n\
  echo "[startup] Starting pre-initialized PostgreSQL..."\n\
  # Start PostgreSQL server using the dedicated script\n\
  if command -v runuser &>/dev/null; then\n\
      runuser -u postgres -- /usr/local/bin/start-postgres.sh\n\
  else\n\
      su - postgres -c /usr/local/bin/start-postgres.sh\n\
  fi\n\
fi\n\
exec "$@"' > /docker-entrypoint.sh && \
chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/vtcpd/vtcpd"]


###################################################################################
# Final stage for Ubuntu
FROM runtime-ubuntu AS final-ubuntu

# Copy pre-initialized PostgreSQL data from init stage
COPY --from=postgres-init-ubuntu --chown=postgres:postgres /var/lib/postgresql/data /var/lib/postgresql/data
# Fix permissions for PostgreSQL data directory
RUN chmod 700 /var/lib/postgresql/data

# Create vtcpd directory and set permissions
RUN mkdir -p /vtcp
WORKDIR /vtcp
RUN mkdir -p vtcpd

# Copy pre-built binaries from host
COPY ./deps/vtcpd/vtcpd /vtcp/vtcpd/
COPY ./deps/cli/cli /vtcp/cli
COPY ./start-postgres.sh /usr/local/bin/start-postgres.sh
RUN chmod +x /usr/local/bin/start-postgres.sh

# Create run directory for PostgreSQL socket
RUN mkdir -p /var/run/postgresql && \
    chown -R postgres:postgres /var/run/postgresql && \
    chmod 775 /var/run/postgresql

# Define build arguments and convert them to runtime environment variables
ARG VTCPD_LISTEN_ADDRESS=127.0.0.1
ARG VTCPD_LISTEN_PORT=2000
ARG VTCPD_EQUIVALENTS_REGISTRY=eth
ARG VTCPD_MAX_HOPS=5
ARG CLI_LISTEN_ADDRESS=127.0.0.1
ARG CLI_LISTEN_PORT=3000
ARG CLI_LISTEN_PORT_TESTING=3001
ARG VTCPD_DATABASE_CONFIG=sqlite3:///io

ENV VTCPD_LISTEN_ADDRESS=${VTCPD_LISTEN_ADDRESS}
ENV VTCPD_LISTEN_PORT=${VTCPD_LISTEN_PORT}
ENV VTCPD_EQUIVALENTS_REGISTRY=${VTCPD_EQUIVALENTS_REGISTRY}
ENV VTCPD_MAX_HOPS=${VTCPD_MAX_HOPS}
ENV CLI_LISTEN_ADDRESS=${CLI_LISTEN_ADDRESS}
ENV CLI_LISTEN_PORT=${CLI_LISTEN_PORT}
ENV CLI_LISTEN_PORT_TESTING=${CLI_LISTEN_PORT_TESTING}
ENV VTCPD_DATABASE_CONFIG=${VTCPD_DATABASE_CONFIG}

# Create startup script that uses runtime environment variables
RUN echo '#!/bin/bash\n\
cd /vtcp\n\
# Create vtcpd config file
cat <<EOF > /vtcp/vtcpd/conf.json\n\
{\n\
  "addresses": [\n\
    {\n\
      "type": "ipv4",\n\
      "address": "${VTCPD_LISTEN_ADDRESS}:${VTCPD_LISTEN_PORT}"\n\
    }\n\
  ],\n\
  "database_config": "${VTCPD_DATABASE_CONFIG}",\n\
  "equivalents_registry_address": "${VTCPD_EQUIVALENTS_REGISTRY}",\n\
  "max_hops_count": ${VTCPD_MAX_HOPS}\n\
}\n\
EOF\n\
# Create cli config file
    cat <<EOF > /vtcp/conf.yaml\n\
workdir: "/vtcp/vtcpd/"\n\
vtcpd_path: "/vtcp/vtcpd/vtcpd"\n\
http:\n\
  host: "${CLI_LISTEN_ADDRESS}"\n\
  port: ${CLI_LISTEN_PORT}\n\
http_testing:\n\
  host: "${CLI_LISTEN_ADDRESS}"\n\
  port: ${CLI_LISTEN_PORT_TESTING}\n\
EOF\n\
# Start PostgreSQL if configured (database already initialized)\n\
if [[ "${VTCPD_DATABASE_CONFIG}" == *"postgresql"* ]]; then\n\
  echo "[startup] Starting pre-initialized PostgreSQL..."\n\
  # Start PostgreSQL server using the dedicated script\n\
  if command -v runuser &>/dev/null; then\n\
      runuser -u postgres -- /usr/local/bin/start-postgres.sh\n\
  else\n\
      su - postgres -c /usr/local/bin/start-postgres.sh\n\
  fi\n\
fi\n\
exec "$@"' > /docker-entrypoint.sh && \
chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["./cli", "start-http"]
