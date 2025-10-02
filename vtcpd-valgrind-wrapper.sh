#!/bin/bash

# VTCPD Valgrind Wrapper Script
# This script wraps vtcpd execution with valgrind for memory debugging

# Default valgrind options
VALGRIND_ENABLED="${VALGRIND_ENABLED:-false}"
VALGRIND_OPTS="${VALGRIND_OPTS:---leak-check=full --track-origins=yes --log-file=/vtcp/valgrind.log}"
VTCPD_BINARY="${VTCPD_BINARY:-/vtcp/vtcpd/vtcpd}"

# Check if valgrind should be enabled
if [[ "${VALGRIND_ENABLED,,}" == "true" ]]; then
    echo "[valgrind-wrapper] Starting vtcpd under valgrind..."
    echo "[valgrind-wrapper] Valgrind options: ${VALGRIND_OPTS}"
    echo "[valgrind-wrapper] VTCPD binary: ${VTCPD_BINARY}"
    echo "[valgrind-wrapper] Log file: /vtcp/valgrind.log"
    echo "[valgrind-wrapper] Working directory: $(pwd)"
    echo "[valgrind-wrapper] Current user: $(whoami)"
    
    # Create log file beforehand to ensure it exists
    touch /vtcp/valgrind.log
    chmod 666 /vtcp/valgrind.log
    echo "[valgrind-wrapper] Created log file: $(ls -la /vtcp/valgrind.log)"
    
    # Execute vtcpd under valgrind
    exec valgrind ${VALGRIND_OPTS} "${VTCPD_BINARY}" "$@"
else
    echo "[valgrind-wrapper] Starting vtcpd directly (valgrind disabled)..."
    echo "[valgrind-wrapper] VTCPD binary: ${VTCPD_BINARY}"
    
    # Execute vtcpd directly
    exec "${VTCPD_BINARY}" "$@"
fi