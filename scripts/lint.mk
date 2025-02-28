.PHONY: fmt swag golines lint

ROOT_DIR := $(realpath $(dir $(lastword $(MAKEFILE_LIST)))/..)

fmt:
	cd $(ROOT_DIR) && go fmt ./...

swag:
	cd $(ROOT_DIR) && swag fmt

golines:
	cd $(ROOT_DIR) && golines -w -m 120 --no-reformat-tags .

lint:
	cd $(ROOT_DIR) && golangci-lint run --timeout 10m

all: fmt swag golines lint