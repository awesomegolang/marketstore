################################################################################
# Commands to automate integration tests
################################################################################

LATEST_TAG ?= $(shell bin/dockertags alpacamarkets/marketstore | tail -1)
IMAGE_NAME ?= alpacamarkets/marketstore:${LATEST_TAG}
CONTAINER_NAME = integration_tests_mstore
MARKETSTORE_DEV_DB_URL = https://s3.amazonaws.com/dev.alpaca.markets/gobroker/mktsdb.tar.gz
MARKETSTORE_DEV_DB_TMPFILE = /tmp/mktsdb.tgz

# User targets
################################################################################

# start a marketstore docker container and check if ListSymbols API can be consumed
.PHONY: connect
connect: run
	@curl -s \
	--data-binary '{"jsonrpc":"2.0", "method":"DataService.ListSymbols", "id":1, "params": {"parameters": {}}}' \
	-H 'Content-Type: application/json' http://localhost:5993/rpc ; \
	\
	if [ $$? -ne 0 ]; then \
		echo "Failed"; \
	else \
		echo "Passed"; \
	fi

.PHONY: run
run: _init
	@if [ `bin/check_running ${CONTAINER_NAME}` -eq 0 ]; then \
		$(MAKE) _startup; \
	fi

.PHONY: stop
stop:
	-if [ `bin/check_running ${CONTAINER_NAME}` -eq 1 ]; then \
		docker stop ${CONTAINER_NAME}; fi
	-docker rm -f ${CONTAINER_NAME}

.PHONY: test
test:
	@bin/runtests.sh

.PHONY: clean
clean: stop
	rm -rf data tests/__pycache__


# Utils
################################################################################
.PHONY: _init
_init:
	@if [ ! -d data/mktsdb ]; then \
		rm -rf data; \
		$(MAKE) _get_data; \
	fi

.PHONY: _get_data
_get_data:
	@rm -rf data && mkdir data
	@if [ ! -f ${MARKETSTORE_DEV_DB_TMPFILE} ]; then \
		wget ${MARKETSTORE_DEV_DB_URL} -O ${MARKETSTORE_DEV_DB_TMPFILE}; \
	fi
	@tar -C data -xzf ${MARKETSTORE_DEV_DB_TMPFILE}

.PHONY: _startup
_startup: stop
	@echo "Starting a marketstore instance..."
	docker run --name ${CONTAINER_NAME} -d -p 5993:5993 -v $(CURDIR):/project -w /project $(IMAGE_NAME) \
		start --config /project/bin/mkts.yml
	@sleep 2
	@if [ `bin/check_running ${CONTAINER_NAME}` -eq 0 ]; then \
		echo "Failed to start a marketstore instance"; \
		false; \
	fi


# Tests
################################################################################
.PHONY: _start_pyclient_container
_start_pyclient_container:
	make -C dockerfiles/pyclient rm build run

.PHONY: test_driver
test_driver: clean run _start_pyclient_container
	TEST_FILENAME='/project/tests/$@.py'; \
	make -C dockerfiles/pyclient test

.PHONY: test_not_empty
test_not_empty: clean run _start_pyclient_container
	TEST_FILENAME='/project/tests/$@.py'; \
	make -C dockerfiles/pyclient test
