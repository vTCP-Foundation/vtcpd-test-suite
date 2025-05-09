###################################################################################
# Manjaro Linux runtime environment
FROM manjarolinux/base AS runtime-manjaro
RUN pacman -Syu --noconfirm && \
    pacman -S --noconfirm \
    boost \
    boost-libs \
    libsodium \
    sqlite && \
    # Debug library locations
    echo "Library locations:" && \
    ls -la /usr/lib/libboost_* && \
    ls -la /usr/lib/libsodium* && \
    echo "Library dependencies:" && \
    ldd /vtcpd/vtcpd || true

# Create a non-root user for running the daemon
RUN useradd -r -s /usr/bin/nologin vtcpd


###################################################################################
# Ubuntu runtime environment
FROM ubuntu:24.04 AS runtime-ubuntu
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    libboost-system1.83.0 \
    libboost-filesystem1.83.0 \
    libboost-program-options1.83.0 \
    libsodium23 \
    vim \
    sqlite3 && \
    rm -rf /var/lib/apt/lists/* && \
    useradd -r -s /usr/sbin/nologin vtcpd


###################################################################################
# Final stage for Manjaro Linux
FROM runtime-manjaro AS final-manjaro

# Create vtcpd directory and set permissions
RUN mkdir -p /vtcpd
WORKDIR /vtcpd
RUN mkdir -p vtcpd

# Copy pre-built binaries from host
COPY ./deps/vtcpd /vtcpd/
COPY ./deps/cli/cli /vtcpd/cli

# Define build arguments and convert them to runtime environment variables
ARG VTCPD_LISTEN_ADDRESS=127.0.0.1
ARG VTCPD_LISTEN_PORT=2000
ARG VTCPD_EQUIVALENTS_REGISTRY=eth
ARG VTCPD_MAX_HOPS=5
ARG CLI_LISTEN_ADDRESS=127.0.0.1
ARG CLI_LISTEN_PORT=3000
ARG CLI_LISTEN_PORT_TESTING=3001

ENV VTCPD_LISTEN_ADDRESS=${VTCPD_LISTEN_ADDRESS}
ENV VTCPD_LISTEN_PORT=${VTCPD_LISTEN_PORT}
ENV VTCPD_EQUIVALENTS_REGISTRY=${VTCPD_EQUIVALENTS_REGISTRY}
ENV VTCPD_MAX_HOPS=${VTCPD_MAX_HOPS}
ENV CLI_LISTEN_ADDRESS=${CLI_LISTEN_ADDRESS}
ENV CLI_LISTEN_PORT=${CLI_LISTEN_PORT}
ENV CLI_LISTEN_PORT_TESTING=${CLI_LISTEN_PORT_TESTING}

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
exec "$@"' > /docker-entrypoint.sh && \
chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/vtcpd/vtcpd"]


###################################################################################
# Final stage for Ubuntu
FROM runtime-ubuntu AS final-ubuntu

# Create vtcpd directory and set permissions
RUN mkdir -p /vtcp
WORKDIR /vtcp
RUN mkdir -p vtcpd

# Copy pre-built binaries from host
COPY ./deps/vtcpd/vtcpd /vtcp/vtcpd/
COPY ./deps/cli/cli /vtcp/cli

# Define build arguments and convert them to runtime environment variables
ARG VTCPD_LISTEN_ADDRESS=127.0.0.1
ARG VTCPD_LISTEN_PORT=2000
ARG VTCPD_EQUIVALENTS_REGISTRY=eth
ARG VTCPD_MAX_HOPS=5
ARG CLI_LISTEN_ADDRESS=127.0.0.1
ARG CLI_LISTEN_PORT=3000
ARG CLI_LISTEN_PORT_TESTING=3001

ENV VTCPD_LISTEN_ADDRESS=${VTCPD_LISTEN_ADDRESS}
ENV VTCPD_LISTEN_PORT=${VTCPD_LISTEN_PORT}
ENV VTCPD_EQUIVALENTS_REGISTRY=${VTCPD_EQUIVALENTS_REGISTRY}
ENV VTCPD_MAX_HOPS=${VTCPD_MAX_HOPS}
ENV CLI_LISTEN_ADDRESS=${CLI_LISTEN_ADDRESS}
ENV CLI_LISTEN_PORT=${CLI_LISTEN_PORT}
ENV CLI_LISTEN_PORT_TESTING=${CLI_LISTEN_PORT_TESTING}

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
exec "$@"' > /docker-entrypoint.sh && \
chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["./cli", "start-http"]
