# CLI Vendor Directory

This directory contains symlinks to the CLI binary and its configuration file.

## Expected Structure
```
.
├── cli             # Symlink to CLI binary
└── cli-conf.json   # Symlink to CLI configuration file
```

## Requirements

### CLI Binary
- The CLI binary should be executable
- Recommended location: outside of the test suite directory
- Symlinked via `make init-deps-symlinks` using the `CLI_BIN` path from Makefile

### Configuration File
- Format: YAML or JSON (based on your CLI requirements)
- Recommended location: outside of the test suite directory
- Symlinked via `make init-deps-symlinks` using the `CLI_CONF` path from Makefile