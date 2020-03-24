APP_NAME ?= ldap-exporter
PROJECT_NAME ?= ldap-exporter
EXPORTER_VERSION=0.0.1
VERSION=${EXPORTER_VERSION}-5
ARCH=amd64
TEMPDIR := $(shell mktemp -d)
SHELL := /bin/bash -ex

export DOCKER_HUB ?= bo01-vm-nexus01.node.bo01.noroutine.me:5000
export GO_VERSION ?= 1.14
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

CI_JOB_ID ?= $(shell openssl rand -hex 4)

DOCKER_ID := ${APP_NAME}-build-${CI_JOB_ID}
DOCKER_TAG := ${APP_NAME}-build:${CI_JOB_ID}

DOCKER_DIST_TAG := noroutine/${APP_NAME}:${VERSION}-${BUILD_TYPE}

.PHONY: all

all: clean build

clean:
	rm -rf ${DIST_DIR}

include Makefile.tools

compile = GOOS=${1} GOARCH=amd64 go build ${BUILD_FLAGS} -o ${OUTPUT_DIR}/${APP_NAME}-${1}-${VERSION} ./app/main

format:
	@echo "Running goimports ..."
	@goimports -w -local github.com/tomcz/openldap_exporter $(shell find . -type f -name '*.go' | grep -v '/vendor/')

build:
	mkdir -p dist/${APP_NAME}_build ${OUTPUT_DIR}
	rsync -av go.mod go.sum Makefile.tools gitlab/build_container/ dist/${APP_NAME}_build
	tar -z -c -f dist/${APP_NAME}_build/src.tar.gz \
		--exclude './dist/*' \
		--exclude './vendor/*' \
		--exclude './.git/*' \
		--exclude './_tools/*' \
		--exclude './.idea/*' \
		.

	docker build \
		--tag ${DOCKER_TAG} \
		--build-arg GO_VERSION=${GO_VERSION} \
		--build-arg DOCKER_HUB=${DOCKER_HUB} \
		dist/${APP_NAME}_build

	docker rm ${DOCKER_ID} || true
	docker run --name ${DOCKER_ID} \
		-e APP_NAME=${APP_NAME} \
		-e USER_EMAIL=${USER_EMAIL} \
		-e BUILD_HASH=${BUILD_HASH} \
		-e BUILD_TYPE=${BUILD_TYPE} \
		-e BUILD_HOST=${BUILD_HOST} \
		-e OUTPUT_DIR=/dist \
		-w /go/src/nrtn.io/ldap_exporter ${DOCKER_TAG} make build-local

	docker cp ${DOCKER_ID}:/dist/. ${OUTPUT_DIR}/
	docker rm ${DOCKER_ID}

dist: build
	mkdir -p dist/${APP_NAME}_dist
	rsync -av gitlab/dist_container/ dist/${APP_NAME}_dist
	rsync -av ${OUTPUT_DIR}/${APP_NAME}-linux-${VERSION} dist/${APP_NAME}_dist/${APP_NAME}-linux

	docker build \
		--tag ${DOCKER_DIST_TAG} \
		--build-arg DOCKER_HUB=${DOCKER_HUB} \
		dist/${APP_NAME}_dist

build-local:
	mkdir -p ${OUTPUT_DIR}
	$(call compile,linux)
	$(call compile,darwin)

upload-nexus:
#	curl -u "${NEXUS_USERNAME}:${NEXUS_PASSWORD}" \
#		--upload-file "${OUTPUT_DIR}/${APP_NAME}" "${NEXUS_REPO_URL}/repository/raw/witnessd/${BUILD_TYPE}/${APP_NAME}"

upload-nfs:
#	mkdir -p /ci/witnessd/${BUILD_TYPE}/
#	cp ${OUTPUT_DIR}/${APP_NAME} /ci/witnessd/${BUILD_TYPE}/


