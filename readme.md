# vTCP Testsuite

## Overview
This repository contains a test suite for vTCP, including automated testing processes using Docker containers.

## Prerequisites
- Docker installed and running
- Make utility
- Access to vTCP daemon and CLI binaries

## Building
### 1. Repository Setup

1. Clone this repository
2. Copy `Makefile.example` to `Makefile`:
   ```bash
   cp Makefile.example Makefile
   ```

### 2. Configure Deps Binaries

The test suite requires vTCP daemon and CLI binaries to run. Follow these steps:

1. Update the `Makefile` with the correct paths to your existing binaries and config files:
   ```bash
   VTCPD_BIN := /path/to/vtcpd/binary
   VTCPD_CONF := /path/to/vtcpd/config
   CLI_BIN := /path/to/cli/binary
   CLI_CONF := /path/to/cli/config
   ```

2. Initialize the deps directory structure by creating necessary symlinks:
   ```bash
   make init-deps-symlinks
   ```
   This will automatically set up the required binaries and config files in the deps directory.

3. Verify your setup by checking the deps directory structure:
   ```bash
   ls -la deps/cli/
   ls -la deps/vtcpd/
   ```

   You should see something like this:
   ```
   deps/cli/
   ├── cli -> /path/to/cli/binary
   └── config -> /path/to/cli/config

   deps/vtcpd/
   ├── vtcpd -> /path/to/vtcpd/binary
   └── config -> /path/to/vtcpd/config
   ```
   
   If the symlinks are missing or pointing to incorrect locations, double-check your `Makefile` paths and run `make init-deps-symlinks` again.

### 3. Build Test Environment

1. Build the Docker test image:
   ```bash
   make docker-build-test-<distro>
   ```
   - Replace `<distro>` with your target distribution (e.g., ubuntu, manjaro) <br>
   Example: `make docker-build-test-ubuntu` <br>
   - Update `Makefile` to support your distro if needed.

## Running Tests

Execute the test suite:
```bash
make test
```

## Directory Structure
```
.
├── Makefile.example    # Template for Makefile configuration
├── deps/
│   ├── cli/            # CLI binary and configuration
│   └── vtcpd/          # vTCP daemon binary and configuration
└── tests/              # Test suite files
```

## Troubleshooting

If you encounter any issues:
1. Verify the paths in your `Makefile` point to valid binaries and configs
2. Check that Docker is running and has proper permissions
3. Ensure the symlinks were created correctly by running `make init-deps-symlinks` again

If the steps above don't resolve your issue, please [open a new issue](https://github.com/vTCP-Foundation/vtcp-test/issues/new) with:
- Description of the problem
- Steps to reproduce
- Output of `ls -la deps/*/`
- Your OS and environment details

For more detailed information about the binaries and configurations, see:
- [CLI README](deps/cli/readme.md)
- [vTCP Daemon README](deps/vtcpd/readme.md)


