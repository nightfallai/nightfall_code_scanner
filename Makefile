# Let Go know that our modules are private
export GOPRIVATE=github.com/watchtowerai

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

SERVICE_NAME=nightfall_dlp
BINARY_NAME=./$(SERVICE_NAME)

# docker parameters
NAME=watchtowerai/$(SERVICE_NAME)
TAG=$(shell git log -1 --pretty=format:"%H")
VERSION=$(NAME):$(TAG)
LATEST=$(NAME):latest
GIT_USER?=""
GIT_PASS?=""
GO_TEST_ENV?=test
GO_GIN_MODE?=release

all: clean build start
build:
	$(GOBUILD) -o $(BINARY_NAME) -v
test:
	GIN_MODE=$(GO_GIN_MODE) GO_ENV=$(GO_TEST_ENV) $(GOTEST) ./... -count=1 -coverprofile=coverage.out
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
start:
	./$(BINARY_NAME)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
deps:
	go mod download
generate:
	go generate ./...
