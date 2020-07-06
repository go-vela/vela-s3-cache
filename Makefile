# Copyright (c) 2019 Target Brands, Inc. All rights reserved.
#
# Use of this source code is governed by the LICENSE file in this repository.

.PHONY: build
build: binary-build

.PHONY: run
run: build docker-build docker-run

.PHONY: test
test: build docker-build docker-example

#################################
######      Go clean       ######
#################################
.PHONY: clean
clean:

	@go mod tidy
	@go vet ./...
	@go fmt ./...
	@echo "I'm kind of the only name in clean energy right now"

#################################
######    Build Binary     ######
#################################
.PHONY: binary-build
binary-build:

	GOOS=linux CGO_ENABLED=0 go build -o release/vela-s3-cache github.com/go-vela/vela-s3-cache/cmd/vela-s3-cache

#################################
######    Docker Build     ######
#################################
.PHONY: docker-build
docker-build:

	docker build --no-cache -t s3-cache-plugin:local .

#################################
######     Docker Run      ######
#################################
.PHONY: docker-run
docker-run:

	docker run --rm \
		-e PARAMETER_SERVER \
		-e PARAMETER_ACCESS_KEY \
		-e PARAMETER_SECRET_KEY \
		-e PARAMETER_FILENAME \
		-e PARAMETER_LOG_LEVEL \
		-e PARAMETER_ROOT \
		-e PARAMETER_MOUNT \
		-e PARAMETER_ACTION \
		-e REPOSITORY_ORG \
		-e REPOSITORY_NAME \
		-v $(PWD)/example/:/home/ \
		-w /home/ \
		s3-cache-plugin:local

.PHONY: docker-example
docker-example:

	docker run --rm \
		-e PARAMETER_SERVER=http://localhost:9000 \
		-e PARAMETER_ACCESS_KEY=access_key \
		-e PARAMETER_SECRET_KEY=secret_key \
		-e PARAMETER_FILENAME=hello.tar \
		-e PARAMETER_DEBUG=false \
		-e PARAMETER_ROOT=bucket_name \
		-e REPOSITORY_ORG=vela-plugins \
		-e REPOSITORY_NAME=s3-cache \
		-e PARAMETER_FLUSH \
		-e PARAMETER_REBUILD \
		-e PARAMETER_RESTORE \
		-v $(PWD)/example/:/home/ \
		-w /home/ \
		s3-cache-plugin:local
