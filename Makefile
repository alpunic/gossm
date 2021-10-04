# ---------------------------------------------------------
# -- Variables
# ---------------------------------------------------------
GOVERSIONTAG = golang:1.17.1-alpine
GOCACHE = $(shell go env GOCACHE)
GOPATH  = $(shell go env GOPATH)
ACRHOST?=ACRHOST

# make any target given at the command line PHONY
.PHONY: default $(MAKECMDGOALS)
default: help
help:
	@echo TODO: display help

container-acr-check:
	@echo "Check current login user for ${ACRHOST} registry"
	@docker login --get-login ${ACRHOST} >/dev/null

container-acr-login:
	@docker login -u "${ACRUSER}" -p "${ACRPASS}" ${ACRHOST}

container-acr-logout:
	@docker logout ${ACRHOST}

build-zombie:
	docker build -t ${ACRHOST}/xi-zombie-mon .

push-zombie: container-acr-check build-zombie
	docker push ${ACRHOST}/xi-zombie-mon

run-zombie:
	go run cmd/gossm/main.go
