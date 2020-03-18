ARG DOCKER_HUB
ARG GO_VERSION

FROM ${DOCKER_HUB}/golang:${GO_VERSION}-buster

MAINTAINER Noroutine UG <info@noroutine.me>

ADD sources.list /etc/apt/sources.list

RUN apt-get -qqy update && DEBIAN_FRONTEND=noninteractive apt-get install -qqy netcat unzip \
    protobuf-compiler rsync \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /go/src/nrtn.io/ldap_exporter
COPY go.sum /go/src/nrtn.io/ldap_exporter/go.sum
COPY go.mod /go/src/nrtn.io/ldap_exporter/go.mod
COPY Makefile.tools /go/src/nrtn.io/ldap_exporter/Makefile.tools

WORKDIR /go/src/nrtn.io/ldap_exporter

RUN make -f Makefile.tools install-tools

COPY src.tar.gz /src.tar.gz
RUN tar -zxf /src.tar.gz
