# 服务演示地址  
# http://149.88.72.48:8000/basic-api

## 接口文档 
## https://apifox.com/apidoc/shared-a2b9478a-1782-4094-95d2-dfd9d6c883ec


# Kratos Project Template

## Install Kratos
```
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
```
## Create a service
```
# Create a template project
kratos new server

cd server
# Add a proto template
kratos proto add api/server/server.proto
# Generate the proto code
kratos proto client api/server/server.proto
# Generate the source code of service by proto file
kratos proto server api/server/server.proto -t internal/service

go generate ./...
go build -o ./bin/ ./...
./bin/server -conf ./configs
```
## Generate other auxiliary files by Makefile
```
# Download and update dependencies
make init
# Generate API files (include: pb.go, http, grpc, validate, swagger) by proto file
make api
# Generate all files
make all
```
## Automated Initialization (wire)
```
# install wire
go get github.com/google/wire/cmd/wire

# generate wire
cd cmd/base-server
wire
```

## Docker
```bash
# build
docker build -t <your-docker-image-name> .

# save image
docker save <your-docker-image-name> | gzip > <your-docker-image-name>.tar.gz

# load image
gunzip -c <your-docker-image-name>.tar.gz | docker load

# run
docker run --rm -p 8000:8000 -p 9000:9000 -v </path/to/your/configs>:/data/conf <your-docker-image-name>
```

## ent
```bash
base-server/internal/data$ ent generate ./schema --target ./ent
```