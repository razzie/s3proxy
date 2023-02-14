VERSION ?= `git describe --tags`
BUILDFLAGS := -ldflags="-s -w" -gcflags=-trimpath=$(CURDIR)
IMAGE_NAME := s3proxy
IMAGE_REGISTRY ?= ghcr.io/razzie
FULL_IMAGE_NAME := $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(VERSION)

.PHONY: all
all: s3proxy encrypt decrypt

.PHONY: s3proxy
s3proxy:
	go build $(BUILDFLAGS) ./cmd/s3proxy

.PHONY: encrypt
encrypt:
	go build $(BUILDFLAGS) ./cmd/encrypt

.PHONY: decrypt
decrypt:
	go build $(BUILDFLAGS) ./cmd/decrypt

.PHONY: docker-build
docker-build:
	docker build . -t $(FULL_IMAGE_NAME)

.PHONY: docker-push
docker-push: docker-build
	docker push $(FULL_IMAGE_NAME)
