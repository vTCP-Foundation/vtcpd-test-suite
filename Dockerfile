###################################################################################
# Builder for Ubuntu: OpenSSL (>=3.5 dev), liboqs, oqs-provider
FROM ubuntu:24.04 AS builder-ubuntu-crypto
ARG DEBIAN_FRONTEND=noninteractive
ARG OPENSSL_REF=openssl-3.5.0
ARG LIBOQS_REF=main
ARG OQS_PROVIDER_REF=main
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      build-essential ca-certificates git curl wget perl pkg-config \
      cmake ninja-build python3 zlib1g-dev && \
    rm -rf /var/lib/apt/lists/*
# Build OpenSSL from source (master → 3.5+/3.6.0-dev)
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
# Build oqs-provider (installs into /usr/local/lib{,64}/ossl-modules)
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
ARG OPENSSL_REF=openssl-3.5.0
ARG LIBOQS_REF=main
ARG OQS_PROVIDER_REF=main
RUN pacman -Syu --noconfirm && \
    pacman -S --noconfirm \
      base-devel git cmake ninja perl python \
      zlib && \
    true
# Build OpenSSL from source (master → 3.5+/3.6.0-dev)
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
    ca-certificates \
    curl \
    git \
    cmake \
    make \
    rsync \
    libsodium \
    postgresql-libs \
    postgresql \
    gcc \
    sqlite && \
    # Debug library locations
    echo "Library locations:" && \
    ls -la /usr/lib/libboost_* || true && \
    ls -la /usr/lib/libsodium* && \
    echo "Library dependencies:" && \
    ldd /vtcpd/vtcpd || true

# Install custom-built OpenSSL + oqs-provider from builder and auto-load providers
COPY --from=builder-manjaro-crypto /usr/local /usr/local
ENV OPENSSL_MODULES=/usr/local/lib64/ossl-modules
ENV OPENSSL_CONF=/usr/local/ssl/openssl.cnf
ENV LD_LIBRARY_PATH=/usr/local/lib64:/usr/local/lib
RUN set -eux; \
    echo '/usr/local/lib' > /etc/ld.so.conf.d/00-openssl_local.conf; \
    echo '/usr/local/lib64' >> /etc/ld.so.conf.d/00-openssl_local.conf; \
    ldconfig; \
    if [ ! -d /usr/local/lib64/ossl-modules ] && [ -d /usr/local/lib/ossl-modules ]; then \
      mkdir -p /usr/local/lib64 && ln -s /usr/local/lib/ossl-modules /usr/local/lib64/ossl-modules; \
    fi; \
    mkdir -p /usr/local/ssl && cat > /usr/local/ssl/openssl.cnf <<'EOF' 
openssl_conf = openssl_init

[openssl_init]
providers = provider_sect

[provider_sect]
default = default_sect
oqsprovider = oqs_sect

[default_sect]
activate = 1

[oqs_sect]
module = oqsprovider
activate = 1
EOF
RUN /usr/local/bin/openssl version && \
    /usr/local/bin/openssl list -signature-algorithms -provider oqsprovider -provider-path "$OPENSSL_MODULES" -provider default | grep -Ei 'slh-.*dsa|sphincs'

# Create a non-root user for running the daemon
RUN useradd -r -s /usr/bin/nologin vtcpd && \
    # Create postgres user for database operations
    useradd -r -s /bin/bash postgres || true

# --- OR-Tools (source build for Manjaro) ---
ARG ORTOOLS_VER=v9.14
ARG ORTOOLS_ROOT=/opt/or-tools-${ORTOOLS_VER#v}
RUN set -eux; \
    cd /tmp && git clone https://github.com/google/or-tools.git && cd or-tools; \
    git checkout "${ORTOOLS_VER}"; \
    cmake -S . -B build -DCMAKE_BUILD_TYPE=Release -DBUILD_DEPS=ON -DBUILD_TESTING=OFF -DBUILD_EXAMPLES=OFF -DBUILD_PYTHON=OFF -DBUILD_JAVA=OFF -DBUILD_DOTNET=OFF; \
    cmake --build build -j; \
    cmake --install build --prefix "${ORTOOLS_ROOT}"; \
    ln -sfn "${ORTOOLS_ROOT}" /opt/or-tools; \
    if [ -d "${ORTOOLS_ROOT}/lib" ]; then echo "${ORTOOLS_ROOT}/lib" > /etc/ld.so.conf.d/ortools.conf; ldconfig; fi; \
    rm -rf /tmp/or-tools
ENV CMAKE_PREFIX_PATH=/opt/or-tools

# --- Boost (source build 1.87.0 for Manjaro) ---
ARG BOOST_VER=1.87.0
ARG BOOST_DIR=boost_1_87_0
ARG BOOST_ROOT=/opt/boost-${BOOST_VER}
RUN set -eux; \
    cd /tmp && curl -fL -o boost.tar.gz https://archives.boost.io/release/${BOOST_VER}/source/${BOOST_DIR}.tar.gz; \
    tar -xzf boost.tar.gz && cd ${BOOST_DIR}; \
    ./bootstrap.sh --prefix="${BOOST_ROOT}"; \
    ./b2 -j$(nproc) --with-system --with-filesystem --with-program_options --with-thread --with-date_time --with-atomic install; \
    ln -sfn "${BOOST_ROOT}" /opt/boost; \
    if [ -d "${BOOST_ROOT}/lib" ]; then echo "${BOOST_ROOT}/lib" > /etc/ld.so.conf.d/boost.conf; ldconfig; fi; \
    rm -rf /tmp/boost.tar.gz "/tmp/${BOOST_DIR}"


###################################################################################
# Ubuntu runtime environment
FROM ubuntu:24.04 AS runtime-ubuntu
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    libsodium23 \
    libpq5 \
    postgresql-16 \
    postgresql-client-16 \
    libasan8 \
    libubsan1 \
    ca-certificates \
    curl \
    xz-utils \
    unzip \
    rsync \
    vim \
    sqlite3 \
    valgrind && \
    ln -s /usr/lib/postgresql/*/bin/pg_ctl /usr/local/bin/pg_ctl && \
    ln -s /usr/lib/postgresql/*/bin/initdb /usr/local/bin/initdb && \
    rm -rf /var/lib/apt/lists/* && \
    useradd -r -s /usr/sbin/nologin vtcpd && \
    # Create postgres user for database operations
    useradd -r -s /bin/bash postgres || true

# Install custom-built OpenSSL + oqs-provider from builder and auto-load providers
COPY --from=builder-ubuntu-crypto /usr/local /usr/local
ENV OPENSSL_MODULES=/usr/local/lib64/ossl-modules
ENV OPENSSL_CONF=/usr/local/ssl/openssl.cnf
ENV LD_LIBRARY_PATH=/usr/local/lib64:/usr/local/lib
RUN set -eux; \
    echo '/usr/local/lib' > /etc/ld.so.conf.d/00-openssl_local.conf; \
    echo '/usr/local/lib64' >> /etc/ld.so.conf.d/00-openssl_local.conf; \
    ldconfig; \
    if [ ! -d /usr/local/lib64/ossl-modules ] && [ -d /usr/local/lib/ossl-modules ]; then \
      mkdir -p /usr/local/lib64 && ln -s /usr/local/lib/ossl-modules /usr/local/lib64/ossl-modules; \
    fi; \
    mkdir -p /usr/local/ssl && cat > /usr/local/ssl/openssl.cnf <<'EOF' 
openssl_conf = openssl_init

[openssl_init]
providers = provider_sect

[provider_sect]
default = default_sect
oqsprovider = oqs_sect

[default_sect]
activate = 1

[oqs_sect]
module = oqsprovider
activate = 1
EOF
RUN /usr/local/bin/openssl version && \
    /usr/local/bin/openssl list -signature-algorithms -provider oqsprovider -provider-path "$OPENSSL_MODULES" -provider default | grep -Ei 'slh-.*dsa|sphincs'

# --- OR-Tools (prebuilt C++ archive for Ubuntu) ---
ARG ORTOOLS_VER=v9.14
ARG ORTOOLS_ROOT=/opt/or-tools-${ORTOOLS_VER#v}
# Provide a fixed URL via build-arg to avoid API rate limits or variability.
# Defaults to the Ubuntu 24.04 C++ prebuilt archive for v9.14.
# Override at build time if needed:
#   --build-arg ORTOOLS_PREBUILT_URL=https://github.com/google/or-tools/releases/download/v9.14/or-tools_amd64_ubuntu-24.04_cpp_v9.14.6206.tar.gz
ARG ORTOOLS_PREBUILT_URL="https://github.com/google/or-tools/releases/download/v9.14/or-tools_amd64_ubuntu-24.04_cpp_v9.14.6206.tar.gz"
RUN set -eux; \
    mkdir -p "${ORTOOLS_ROOT}" /tmp/ortools && cd /tmp/ortools; \
    if [ -n "${ORTOOLS_PREBUILT_URL}" ]; then \
      URL="${ORTOOLS_PREBUILT_URL}"; \
    else \
      URLS=$(curl -fsSL "https://api.github.com/repos/google/or-tools/releases/tags/${ORTOOLS_VER}" | awk -F\" '/browser_download_url/ {print $4}'); \
      echo "${URLS}" > urls.txt; \
      URL=$(echo "${URLS}" | grep -Ei 'linux.*(cpp|c\+\+)' | grep -Ei '(x86_64|amd64|linux64|ubuntu)' | head -n1 || true); \
      if [ -z "${URL}" ]; then echo "No matching OR-Tools asset found. Inspect /tmp/ortools/urls.txt"; exit 1; fi; \
      echo "Using: ${URL}"; \
    fi; \
    curl -fL -o or-tools.pkg "${URL}"; \
    case "${URL}" in \
      *.zip) unzip -q or-tools.pkg ;; \
      *.tar.gz) tar -xzf or-tools.pkg ;; \
      *.tar.xz) tar -xJf or-tools.pkg ;; \
      *) echo "Unknown archive format: ${URL}"; exit 1 ;; \
    esac; \
    ROOT_DIR=$(find . -maxdepth 1 -type d -name 'or-tools*' -print -quit); \
    rsync -a "${ROOT_DIR}"/ "${ORTOOLS_ROOT}"/; \
    ln -sfn "${ORTOOLS_ROOT}" /opt/or-tools; \
    if [ -d "${ORTOOLS_ROOT}/lib" ]; then echo "${ORTOOLS_ROOT}/lib" > /etc/ld.so.conf.d/ortools.conf; ldconfig; fi; \
    rm -rf /tmp/ortools
ENV CMAKE_PREFIX_PATH=/opt/or-tools

# --- Boost (source build 1.87.0 for Ubuntu) ---
ARG BOOST_VER=1.87.0
ARG BOOST_DIR=boost_1_87_0
ARG BOOST_ROOT=/opt/boost-${BOOST_VER}
RUN set -eux; \
    mkdir -p /tmp/boost && cd /tmp/boost; \
    for URL in \
      "https://archives.boost.io/release/${BOOST_VER}/source/${BOOST_DIR}.tar.gz" \
      "https://boostorg.jfrog.io/artifactory/main/release/${BOOST_VER}/source/${BOOST_DIR}.tar.gz"; do \
      if curl -fL -o boost.tar.gz "$URL"; then break; fi; \
    done; \
    tar -xzf boost.tar.gz && cd ${BOOST_DIR}; \
    ./bootstrap.sh --prefix="${BOOST_ROOT}"; \
    ./b2 -j$(nproc) --with-system --with-filesystem --with-program_options --with-thread --with-date_time --with-atomic install; \
    ln -sfn "${BOOST_ROOT}" /opt/boost; \
    if [ -d "${BOOST_ROOT}/lib" ]; then echo "${BOOST_ROOT}/lib" > /etc/ld.so.conf.d/boost.conf; ldconfig; fi; \
    rm -rf /tmp/boost


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
ARG VTCPD_MAX_HOPS=6
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
  "max_hops_count": ${VTCPD_MAX_HOPS},\n\
  "observers": [\n\
    {\n\
      "type": "ipv4",\n\
      "address": "172.17.0.1:8085"\n\
    }\n\
  ]\n\
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
COPY ./vtcpd-valgrind-wrapper.sh /usr/local/bin/vtcpd-valgrind-wrapper.sh
RUN chmod +x /usr/local/bin/start-postgres.sh && \
    chmod +x /usr/local/bin/vtcpd-valgrind-wrapper.sh

# Create run directory for PostgreSQL socket
RUN mkdir -p /var/run/postgresql && \
    chown -R postgres:postgres /var/run/postgresql && \
    chmod 775 /var/run/postgresql

# Define build arguments and convert them to runtime environment variables
ARG VTCPD_LISTEN_ADDRESS=127.0.0.1
ARG VTCPD_LISTEN_PORT=2000
ARG VTCPD_EQUIVALENTS_REGISTRY=eth
ARG VTCPD_MAX_HOPS=6
ARG CLI_LISTEN_ADDRESS=127.0.0.1
ARG CLI_LISTEN_PORT=3000
ARG CLI_LISTEN_PORT_TESTING=3001
ARG VTCPD_DATABASE_CONFIG=sqlite3:///io
ARG VALGRIND_ENABLED=false
ARG VALGRIND_OPTS=--leak-check=full --track-origins=yes --log-file=/vtcp/valgrind.log

ENV VTCPD_LISTEN_ADDRESS=${VTCPD_LISTEN_ADDRESS}
ENV VTCPD_LISTEN_PORT=${VTCPD_LISTEN_PORT}
ENV VTCPD_EQUIVALENTS_REGISTRY=${VTCPD_EQUIVALENTS_REGISTRY}
ENV VTCPD_MAX_HOPS=${VTCPD_MAX_HOPS}
ENV CLI_LISTEN_ADDRESS=${CLI_LISTEN_ADDRESS}
ENV CLI_LISTEN_PORT=${CLI_LISTEN_PORT}
ENV CLI_LISTEN_PORT_TESTING=${CLI_LISTEN_PORT_TESTING}
ENV VTCPD_DATABASE_CONFIG=${VTCPD_DATABASE_CONFIG}
ENV VALGRIND_ENABLED=${VALGRIND_ENABLED}
ENV VALGRIND_OPTS=${VALGRIND_OPTS}

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
  "max_hops_count": ${VTCPD_MAX_HOPS},\n\
  "observers": [\n\
    {\n\
      "type": "ipv4",\n\
      "address": "172.17.0.1:8085"\n\
    }\n\
  ]\n\
}\n\
EOF\n\
# Create cli config file
    cat <<EOF > /vtcp/conf.yaml\n\
workdir: "/vtcp/vtcpd/"\n\
vtcpd_path: "/usr/local/bin/vtcpd-valgrind-wrapper.sh"\n\
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
