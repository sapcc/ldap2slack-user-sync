IMAGE   ?= sapcc/ldap2slack
VERSION = $(shell git rev-parse --verify HEAD | head -c 8)
GOOS    ?= $(shell go env | grep GOOS | cut -d'"' -f2)
BINARY  := ldap2slack

LDFLAGS := -X github.com/sapcc/ldap2slack/ldap2slack.VERSION=$(VERSION)
GOFLAGS := -ldflags "$(LDFLAGS)"

SRCDIRS  := .
PACKAGES := $(shell find $(SRCDIRS) -type d)
GOFILES  := $(addsuffix /*.go,$(PACKAGES))
GOFILES  := $(wildcard $(GOFILES))

#GLIDE := $(shell command -v glide 2> /dev/null)

.PHONY: all clean vendor tests static-check

all: test build

build:
	bin/$(GOOS)/$(BINARY) 
	bin/%/$(BINARY): $(GOFILES) Makefile
		GOOS=$* GOARCH=amd64 go build $(GOFLAGS) -v -i -o bin/$*/$(BINARY) .

run: 
	go -o $(BINARY_NAME) -v ./...  
	./$(BINARY_NAME)

static-check:
	@if s="$$(gofmt -s -l *.go pkg 2>/dev/null)"                            && test -n "$$s"; then printf ' => %s\n%s\n' gofmt  "$$s"; false; fi
	@if s="$$(golint . && find pkg -type d -exec golint {} \; 2>/dev/null)" && test -n "$$s"; then printf ' => %s\n%s\n' golint "$$s"; false; fi

tests: build static-check
	go test -v github.com/sapcc/ldap2slack/...

docker-build: tests bin/linux/$(BINARY)
	docker build -t $(IMAGE):$(VERSION) .

docker-push: build
	docker push $(IMAGE):$(VERSION)

helm-latest: push helm-values
	docker tag $(IMAGE):$(VERSION) $(IMAGE):latest
	docker push $(IMAGE):latest

helm-values:
	sed -i '' -e 's/tag:.*/tag: $(VERSION)/g' ./helm/values.yaml

clean:
	rm -rf bin/*

vendor:
	dep ensure
