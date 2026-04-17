GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find app -name '*.proto'")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name '*.proto'")
else
	INTERNAL_PROTO_FILES=$(shell find app -name '*.proto')
	API_PROTO_FILES=$(shell find api -name '*.proto')
endif

.PHONY: init
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: config
# generate internal proto
config:
	for file in $(INTERNAL_PROTO_FILES); do \
		protoc --proto_path=. \
		       --proto_path=./third_party \
		       --go_out=paths=source_relative:. \
		       $$file || exit 1; \
	done

.PHONY: api
# generate api proto
api:
	mkdir -p api/openapi
	protoc --proto_path=./api/protos \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api/gen/go \
 	       --go-http_out=paths=source_relative:./api/gen/go \
 	       --go-grpc_out=paths=source_relative:./api/gen/go \
 	       --go-errors_out=paths=source_relative:./api/gen/go \
	       --openapi_out=fq_schema_naming=true,default_response=false:./api/openapi \
	       $(API_PROTO_FILES)

.PHONY: errors
# generate errors proto
errors:
	protoc --proto_path=./api/protos \
              --proto_path=./third_party \
              --go_out=paths=source_relative:./api/gen/go \
              --go-errors_out=paths=source_relative:./api/gen/go \
              $(API_PROTO_FILES)

.PHONY: openapi
# generate openapi proto
openapi:
	mkdir -p api/openapi
	protoc --proto_path=./api/protos \
	       --proto_path=./third_party \
	       --openapi_out=fq_schema_naming=true,default_response=false:./api/openapi \
	       $(API_PROTO_FILES)

.PHONY: frontend-api
# generate frontend api code from openapi
frontend-api:
	cp ./api/openapi/openapi.yaml ../vben-admin/openapi.yaml
	cd ../vben-admin && pnpm run generate:api
.PHONY: wire

.PHONY: wire
# wire generate code
wire:
	go generate ./...

.PHONY: ent
# generate database code
ent:
	ent generate ./pkg/data/schema \
			--template ./pkg/data/template \
			--feature sql/modifier \
			--target ./pkg/data/ent

.PHONY: build-auth
# build auth service
build-auth:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/auth-service ./app/auth/service/cmd/service

.PHONY: build-user
# build user service
build-user:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/user-service ./app/user/service/cmd/service

.PHONY: build-admin
# build admin service
build-admin:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/admin-service ./app/admin/service/cmd/service

.PHONY: build-common
# build common service
build-common:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/common-service ./app/common/service/cmd/service

.PHONY: build-gateway
# build gateway service
build-gateway:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/gateway-service ./app/gateway/service/cmd/service

.PHONY: build
# build split services
build: build-auth build-user build-admin build-common build-gateway

.PHONY: dev-env-up
# start local dependency services for IDE debugging
dev-env-up:
	docker network inspect base_networks >/dev/null 2>&1 || docker network create base_networks
	docker-compose -f ./docker-compose-env.yml up -d

.PHONY: dev-env-down
# stop local dependency services for IDE debugging
dev-env-down:
	docker-compose -f ./docker-compose-env.yml down

.PHONY: dev-env-reset
# recreate local dependency services after PostgreSQL major upgrades
dev-env-reset:
	docker-compose -f ./docker-compose-env.yml down -v
	docker network inspect base_networks >/dev/null 2>&1 || docker network create base_networks
	docker-compose -f ./docker-compose-env.yml up -d

.PHONY: run-gateway
# run gateway service with local config
run-gateway:
	go run ./app/gateway/service/cmd/service -conf ./app/gateway/service/configs

.PHONY: run-auth
# run auth service with local config
run-auth:
	go run ./app/auth/service/cmd/service -conf ./app/auth/service/configs

.PHONY: run-user
# run user service with local config
run-user:
	go run ./app/user/service/cmd/service -conf ./app/user/service/configs

.PHONY: run-admin
# run admin service with local config
run-admin:
	go run ./app/admin/service/cmd/service -conf ./app/admin/service/configs

.PHONY: run-common
# run common service with local config
run-common:
	go run ./app/common/service/cmd/service -conf ./app/common/service/configs

.PHONY: migrate-auth-permissions-to-admin
# migrate permission tables from auth DB to admin DB
migrate-auth-permissions-to-admin:
	bash ./scripts/migrate_auth_permissions_to_admin.sh

.PHONY: migrate-user-role-bindings-to-admin
# migrate user role bindings from user DB to admin DB
migrate-user-role-bindings-to-admin:
	bash ./scripts/migrate_user_role_bindings_to_admin.sh

.PHONY: drop-user-role-id-column
# drop deprecated user.role_id column from user DB
drop-user-role-id-column:
	bash ./scripts/drop_user_role_id_column.sh

.PHONY: generate
# generate
generate:
	go mod tidy
	go get github.com/google/wire/cmd/wire@latest
	go generate ./...

.PHONY: all
# generate all
all:
	make api;
	make config;
	make generate;

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
