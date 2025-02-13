################################################################################
# 
# deps binaries
#

VTCPD_BIN := /home/hsc/impact/vtcp/vtcpd/build-debug/bin/vtcpd
VTCPD_CONF := /home/hsc/impact/vtcp/vtcpd/build-debug/bin/conf.json

CLI_BIN := /home/hsc/impact/vtcp/cli/bin/cli
CLI_CONF := /home/hsc/impact/vtcp/cli/bin/vtcpd-cli-conf.json

init-deps-symlinks:
	rm -f deps/vtcpd/vtcpd deps/vtcpd/vtcpd-conf.json deps/cli/cli deps/cli/cli-conf.json
	ln -f $(VTCPD_BIN) deps/vtcpd/vtcpd
	ln -f $(VTCPD_CONF) deps/vtcpd/vtcpd-conf.json
	ln -f $(CLI_BIN) deps/cli/cli
	ln -f $(CLI_CONF) deps/cli/cli-conf.json


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
		--build-arg LISTEN_ADDRESS=0.0.0.0 \
		--build-arg LISTEN_PORT=2000 \
		--build-arg EQUIVALENTS_REGISTRY=eth \
		--build-arg MAX_HOPS=5 \
		-t vtcpd-test:manjaro .


# Builds testing image for Ubuntu.
# WARN: Ensure deps binaries are properly set up before running this command!
docker-build-test-ubuntu:
	docker build --target final-ubuntu \
		--build-arg LISTEN_ADDRESS=0.0.0.0 \
		--build-arg LISTEN_PORT=2000 \
		--build-arg EQUIVALENTS_REGISTRY=eth \
		--build-arg MAX_HOPS=5 \
		-t vtcpd-test:ubuntu .


################################################################################
#
# Testing
#

test:
	go test ./tests/... 

