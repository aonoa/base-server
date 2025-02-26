FROM golang:1.23.1-bullseye AS builder

COPY . /src
WORKDIR /src

RUN apt-get update && apt-get install make
#FROM golang:1.21.8-alpine AS builder
#
#COPY . /src
#WORKDIR /src
#
#RUN apk add make
RUN GOPROXY=https://goproxy.cn make build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app
#COPY authconf /app/authconf

WORKDIR /app

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["./base-server", "-conf", "/data/conf"]
