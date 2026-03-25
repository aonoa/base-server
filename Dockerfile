# Multi-service container build
FROM golang:1.25.8-bullseye AS builder

ARG SERVICE=auth

COPY . /src
WORKDIR /src

RUN apt-get update && apt-get install -y --no-install-recommends make && rm -rf /var/lib/apt/lists/*
RUN GOPROXY=https://goproxy.cn make build-${SERVICE}

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates \
	netbase \
	&& rm -rf /var/lib/apt/lists/*

ARG SERVICE=auth
WORKDIR /app

COPY --from=builder /src/bin/${SERVICE}-service /app/service
VOLUME /app/configs

CMD ["./service", "-conf", "/app/configs"]
