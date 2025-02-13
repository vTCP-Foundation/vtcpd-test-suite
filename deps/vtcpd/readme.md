# vTCP Daemon Vendor Directory

This directory contains symlinks to the vTCP daemon binary and its configuration file.

## Expected Structure
```
.
├── vtcpd               # Symlink to vTCP daemon binary
└── vtcpd-conf.json     # Symlink to vTCP daemon configuration file
```

## Requirements

### vTCP Daemon Binary
- The daemon binary should be executable
- Recommended location: outside of the test suite directory
- Symlinked via `make init-deps-symlinks` using the `VTCPD_BIN` path from Makefile

### Configuration File
- Format: YAML or JSON (based on your daemon requirements)
- Recommended location: outside of the test suite directory
- Symlinked via `make init-deps-symlinks` using the `VTCPD_CONF` path from Makefile
