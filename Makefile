.PHONY = build test

# Path Related
MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))
RELEASE_DIR := ${MKFILE_DIR}bin

build: go.sum pserver importer

pserver importer:
	go build -o ${RELEASE_DIR}/$@ ${MKFILE_DIR}cmd/$@/

vet:
	go vet ${MKFILE_DIR}cmd/pserver
	go vet ${MKFILE_DIR}cmd/importer
	go vet ${MKFILE_DIR}internal/database

fmt:
	go fmt ${MKFILE_DIR}cmd/pserver
	go fmt ${MKFILE_DIR}cmd/importer
	go fmt ${MKFILE_DIR}internal/database

go.sum: go.mod
	go mod tidy
