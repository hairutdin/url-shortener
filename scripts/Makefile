SHELL=/bin/bash

UTILS = shortenerstress

all: prep build-utils

prep:
	go mod tidy

build-utils:
	$(foreach TARGET,$(UTILS), \
		GOOS=linux GOARCH=amd64 go build -o=bin/$(TARGET)-linux-amd64 ./cmd/$(TARGET)/... ; \
		GOOS=windows GOARCH=amd64 go build -o=bin/$(TARGET)-windows-amd64.exe ./cmd/$(TARGET)/... ; \
		GOOS=darwin GOARCH=amd64 go build -o=bin/$(TARGET)-darwin-amd64 ./cmd/$(TARGET)/... ; \
		GOOS=darwin GOARCH=arm64 go build -o=bin/$(TARGET)-darwin-arm64 ./cmd/$(TARGET)/... ; \
	)

perm:
	chmod -R +x bin