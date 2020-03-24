ARG DOCKER_HUB

FROM ${DOCKER_HUB}/debian:10.3-slim

MAINTAINER Noroutine UG <info@noroutine.me>

COPY ldap-exporter-linux /ldap-exporter

EXPOSE 9330
ENTRYPOINT [ "/ldap-exporter" ]