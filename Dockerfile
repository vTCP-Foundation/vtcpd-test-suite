###################################################################################
# Builder for Ubuntu: OpenSSL (>=3.5 dev), liboqs, oqs-provider
FROM ubuntu:24.04 AS builder-ubuntu-crypto
ARG DEBIAN_FRONTEND=noninteractive
ARG OPENSSL_REF=master
ARG LIBOQS_REF=main
ARG OQS_PROVIDER_REF=main
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      build-essential ca-certificates git curl wget perl pkg-config \
      cmake ninja-build python3 zlib1g-dev && \
    rm -rf /var/lib/apt/lists/*
# Build OpenSSL from source (master → 3.5.0-dev at time of build)
RUN git clone --depth 1 --branch "$OPENSSL_REF" https://github.com/openssl/openssl.git /tmp/openssl && \
    cd /tmp/openssl && \
    env LDFLAGS='-Wl,-rpath,/usr/local/lib:/usr/local/lib64' ./Configure --prefix=/usr/local --openssldir=/usr/local/ssl shared && \
    make -j"$(nproc)" && make install_sw && \
    rm -rf /tmp/openssl
# Build liboqs
RUN git clone --depth 1 --branch "$LIBOQS_REF" https://github.com/open-quantum-safe/liboqs.git /tmp/liboqs && \
    cmake -S /tmp/liboqs -B /tmp/liboqs/build -GNinja \
      -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_INSTALL_PREFIX=/usr/local \
      -DOQS_BUILD_ONLY_LIB=ON \
      -DBUILD_SHARED_LIBS=ON \
      -DOQS_DIST_BUILD=ON && \
    ninja -C /tmp/liboqs/build && ninja -C /tmp/liboqs/build install && \
    rm -rf /tmp/liboqs
# Build oqs-provider (installs into /usr/local/lib/ossl-modules)
RUN git clone --depth 1 --branch "$OQS_PROVIDER_REF" https://github.com/open-quantum-safe/oqs-provider.git /tmp/oqs-provider && \
    cmake -S /tmp/oqs-provider -B /tmp/oqs-provider/build -GNinja \
      -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_INSTALL_PREFIX=/usr/local \
      -DOPENSSL_ROOT_DIR=/usr/local \
      -DLIBOQS_INSTALL_DIR=/usr/local && \
    ninja -C /tmp/oqs-provider/build && ninja -C /tmp/oqs-provider/build install && \
    rm -rf /tmp/oqs-provider

###################################################################################
# Builder for Manjaro: OpenSSL (>=3.5 dev), liboqs, oqs-provider
FROM manjarolinux/base AS builder-manjaro-crypto
ARG OPENSSL_REF=master
ARG LIBOQS_REF=main
ARG OQS_PROVIDER_REF=main
RUN pacman -Syu --noconfirm && \
    pacman -S --noconfirm \
      base-devel git cmake ninja perl python \
      zlib && \
    true
# Build OpenSSL from source (master → 3.5.0-dev at time of build)
RUN git clone --depth 1 --branch "$OPENSSL_REF" https://github.com/openssl/openssl.git /tmp/openssl && \
    cd /tmp/openssl && \
    env LDFLAGS='-Wl,-rpath,/usr/local/lib:/usr/local/lib64' ./Configure --prefix=/usr/local --openssldir=/usr/local/ssl shared && \
    make -j"$(nproc)" && make install_sw && \
    rm -rf /tmp/openssl
# Build liboqs
RUN git clone --depth 1 --branch "$LIBOQS_REF" https://github.com/open-quantum-safe/liboqs.git /tmp/liboqs && \
    cmake -S /tmp/liboqs -B /tmp/liboqs/build -GNinja \
      -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_INSTALL_PREFIX=/usr/local \
      -DOQS_BUILD_ONLY_LIB=ON \
      -DBUILD_SHARED_LIBS=ON \
      -DOQS_DIST_BUILD=ON && \
    ninja -C /tmp/liboqs/build && ninja -C /tmp/liboqs/build install && \
    rm -rf /tmp/liboqs
# Build oqs-provider
RUN git clone --depth 1 --branch "$OQS_PROVIDER_REF" https://github.com/open-quantum-safe/oqs-provider.git /tmp/oqs-provider && \
    cmake -S /tmp/oqs-provider -B /tmp/oqs-provider/build -GNinja \
      -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_INSTALL_PREFIX=/usr/local \
      -DOPENSSL_ROOT_DIR=/usr/local \
      -DLIBOQS_INSTALL_DIR=/usr/local && \
    ninja -C /tmp/oqs-provider/build && ninja -C /tmp/oqs-provider/build install && \
    rm -rf /tmp/oqs-provider

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
# Install custom-built OpenSSL + oqs-provider from builder
COPY --from=builder-manjaro-crypto /usr/local /usr/local
ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/local/lib64
RUN echo '/usr/local/lib' > /etc/ld.so.conf.d/00-openssl_local.conf && echo '/usr/local/lib64' >> /etc/ld.so.conf.d/00-openssl_local.conf && ldconfig && \
    /usr/local/bin/openssl version && \
    bash -lc '/usr/local/bin/openssl version | grep -Eq "OpenSSL 3\.(5|[6-9])|OpenSSL [4-9]\\."' && \
    MODULES_DIR=$([ -f /usr/local/lib64/ossl-modules/oqsprovider.so ] && echo /usr/local/lib64/ossl-modules || echo /usr/local/lib/ossl-modules) && \
    /usr/local/bin/openssl list -signature-algorithms -provider oqsprovider -provider-path "$MODULES_DIR" -provider default | grep -Ei 'slh-.*dsa|sphincs'

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
    libboost-atomic1.83.0 \
    libboost-thread1.83.0 \
    libboost-date-time1.83.0 \
    libsodium23 \
    libpq5 \
    postgresql-16 \
    postgresql-client-16 \
    libasan8 \
    libubsan1 \
    vim \
    sqlite3 && \
    ln -s /usr/lib/postgresql/*/bin/pg_ctl /usr/local/bin/pg_ctl && \
    ln -s /usr/lib/postgresql/*/bin/initdb /usr/local/bin/initdb && \
    rm -rf /var/lib/apt/lists/* && \
    useradd -r -s /usr/sbin/nologin vtcpd && \
    # Create postgres user for database operations
    useradd -r -s /bin/bash postgres || true
# Install custom-built OpenSSL + oqs-provider from builder
COPY --from=builder-ubuntu-crypto /usr/local /usr/local
ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/local/lib64
RUN bash -lc 'echo /usr/local/lib > /etc/ld.so.conf.d/00-openssl_local.conf' && echo /usr/local/lib64 >> /etc/ld.so.conf.d/00-openssl_local.conf && ldconfig && \
    /usr/local/bin/openssl version && \
    bash -lc '/usr/local/bin/openssl version | grep -Eq "OpenSSL 3\.(5|[6-9])|OpenSSL [4-9]\\."' && \
    MODULES_DIR=$([ -f /usr/local/lib64/ossl-modules/oqsprovider.so ] && echo /usr/local/lib64/ossl-modules || echo /usr/local/lib/ossl-modules) && \
    /usr/local/bin/openssl list -signature-algorithms -provider oqsprovider -provider-path "$MODULES_DIR" -provider default | grep -Ei 'slh-.*dsa|sphincs'


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
