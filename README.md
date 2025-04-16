# 服务演示地址  
# http://149.88.72.48:30080/basic-api

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
docker build -t base-server:v1.1.0 .

# save image
docker save <your-docker-image-name> | gzip > <your-docker-image-name>.tar.gz

# load image
gunzip -c <your-docker-image-name>.tar.gz | docker load

# run
docker run --rm -p 8000:8000 -p 9000:9000 -v </path/to/your/configs>:/data/conf <your-docker-image-name>
```

## ent
使用`--template`添加更新全部字段的模板
```bash
base-server/internal/data$ ent generate ./schema --template ./ent/template --target ./ent
```


## casbin
数据表中手动导入或者迁移的id可能与新增的id冲突导致插入失败，需手动重置序列的当前值
```shell
-- 1. 查询表中当前最大 ID
SELECT MAX(id) FROM casbin_rules;

-- 2. 通过 PostgreSQL 系统表 pg_sequences 直接获取序列的当前值
SELECT last_value FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'casbin_rules_id_seq';

-- 3. 重置序列起始值为当前最大 ID + 1
ALTER SEQUENCE casbin_rules_id_seq RESTART WITH {max_id + 1};
ALTER SEQUENCE casbin_rules_id_seq RESTART WITH 10001;
```