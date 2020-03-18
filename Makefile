APP_NAME ?= ldap-exporter
PROJECT_NAME ?= ldap-exporter
EXPORTER_VERSION=0.0.1
VERSION=${EXPORTER_VERSION}-1
ARCH=amd64
TEMPDIR := $(shell mktemp -d)
SHELL := /bin/bash -ex

export DOCKER_HUB ?= bo01-vm-nexus01.node.bo01.noroutine.me:5000
export GO_VERSION ?= 1.13.4
export NEXUS_REPO_URL ?= https://bo01-vm-nexus01.node.bo01.noroutine.me

BUILD_TYPE ?= dev
ifeq ($(filter $(BUILD_TYPE),debug dev release),)
$(error invalid `BUILD_TYPE` value)
endif

USER_EMAIL ?= info@noroutine.me
BUILD_HASH ?= $(shell git rev-parse HEAD)
BUILD_HOST ?= $(shell hostname)

BUILD_STAMP = `python -c 'from datetime import datetime; print(datetime.now().isoformat())'`

BUILD_IMPORT := nrtn.io/ldap_exporter/app/build
BUILD_FLAGS = $(shell echo "-tags ${BUILD_TYPE} -ldflags \"\
	-X ${BUILD_IMPORT}.Stamp=${BUILD_STAMP} \
	-X ${BUILD_IMPORT}.Hash=${BUILD_HASH} \
	-X ${BUILD_IMPORT}.Type=${BUILD_TYPE} \
	-X ${BUILD_IMPORT}.Email=${USER_EMAIL} \
	-X ${BUILD_IMPORT}.Host=${BUILD_HOST} \
	-X ${BUILD_IMPORT}.Name=${APP_NAME} \
	-X ${BUILD_IMPORT}.Version=${VERSION}\"")

GC_FLAGS = -gcflags "all=-N -l"

DIST_DIR ?= ${PWD}/dist
OUTPUT_DIR ?= ${DIST_DIR}/${BUILD_TYPE}

DOCKER_ID := ${APP_NAME}-build-${CI_JOB_ID}
DOCKER_TAG := ${APP_NAME}-build:${CI_JOB_ID}

.PHONY: all

all: clean build

clean:
	rm -rf ${DIST_DIR}

compile = GOOS=${1} GOARCH=amd64 go build ${BUILD_FLAGS} -o ${OUTPUT_DIR}/${APP_NAME}-${1}-${VERSION} ./app/main

format:
	@echo "Running goimports ..."
	@goimports -w -local github.com/tomcz/openldap_exporter $(shell find . -type f -name '*.go' | grep -v '/vendor/')

build:
	mkdir -p ${OUTPUT_DIR}
	$(call compile,linux)
	$(call compile,darwin)

upload-nexus:
#	curl -u "${NEXUS_USERNAME}:${NEXUS_PASSWORD}" \
#		--upload-file "${OUTPUT_DIR}/${APP_NAME}" "${NEXUS_REPO_URL}/repository/raw/witnessd/${BUILD_TYPE}/${APP_NAME}"

upload-nfs:
#	mkdir -p /ci/witnessd/${BUILD_TYPE}/
#	cp ${OUTPUT_DIR}/${APP_NAME} /ci/witnessd/${BUILD_TYPE}/


