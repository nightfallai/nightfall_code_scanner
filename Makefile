.PHONY: test

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

SERVICE_NAME=nightfall_dlp
BINARY_NAME=./$(SERVICE_NAME)
GO_TEST_ENV?=test

NAME=nightfallai/nightfall_dlp
TAG=$(shell git log -1 --pretty=format:"%H")
VERSION=$(NAME):$(TAG)
LATEST=$(NAME):latest

all: clean build start
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/nightfalldlp/main.go
test:
	GO_ENV=$(GO_TEST_ENV) $(GOTEST) ./... -count=1 -coverprofile=coverage.out
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
start:
	./$(BINARY_NAME)
dockertag:
	docker tag nightfall_dlp:latest $(VERSION)
	docker tag nightfall_dlp:latest $(LATEST)
dockerbuild:
	docker build -t $(VERSION) -t $(LATEST) .
dockerpush:
	docker push $(VERSION)
	docker push $(LATEST)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
deps:
	go mod download
generate:
	go generate ./...
