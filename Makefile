################################################################################
# 
# deps binaries
#

VTCPD_BIN := /home/mc/Personal/vtcp/vtcpd/build-debug/bin/vtcpd
CLI_BIN := /home/mc/Personal/vtcp/vtcpd-cli/build/vtcpd-cli

init-deps-symlinks:
	rm -f deps/vtcpd/vtcpd deps/cli/cli
	ln -f $(VTCPD_BIN) deps/vtcpd/vtcpd
	ln -f $(CLI_BIN) deps/cli/cli


################################################################################
#
# Testing image build
#
# To keep binary compatibility between provided binaries and the testing environment, 
# the testing image build script is provided for multiple target platforms.
#

# Builds testing image for Manjaro.
# WARN: Ensure deps binaries are properly set up before running this command!
docker-build-test-manjaro:
	docker build --target final-manjaro \
		--build-arg VTCPD_LISTEN_ADDRESS=0.0.0.0 \
		--build-arg VTCPD_LISTEN_PORT=2000 \
		--build-arg VTCPD_EQUIVALENTS_REGISTRY=eth \
		--build-arg VTCPD_MAX_HOPS=5 \
		--build-arg CLI_LISTEN_ADDRESS=0.0.0.0 \
		--build-arg CLI_LISTEN_PORT=3000 \
		--build-arg CLI_LISTEN_PORT_TESTING=3001 \
		-t vtcpd-test:manjaro .


# Builds testing image for Ubuntu.
# WARN: Ensure deps binaries are properly set up before running this command!
docker-build-test-ubuntu:
	docker build --target final-ubuntu \
		--build-arg VTCPD_LISTEN_ADDRESS=0.0.0.0 \
		--build-arg VTCPD_LISTEN_PORT=2000 \
		--build-arg VTCPD_EQUIVALENTS_REGISTRY=eth \
		--build-arg VTCPD_MAX_HOPS=5 \
		--build-arg CLI_LISTEN_ADDRESS=0.0.0.0 \
		--build-arg CLI_LISTEN_PORT=3000 \
		--build-arg CLI_LISTEN_PORT_TESTING=3001 \
		-t vtcpd-test:ubuntu .


################################################################################
#
# Testing
#

test:
	go test ./tests/... 

