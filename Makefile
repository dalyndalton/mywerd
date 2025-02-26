APP_NAME := werd
DOCKER_REPO := git.dx2.dev/ddalton/$(APP_NAME)
GO_FILES := $(shell find . -type f -name '*.go')
DOCKERFILE := Dockerfile
DIST_DIR := dist

.PHONY: all build-x86 build-arm64 docker-x86 docker-arm64 push clean

all: build-x86 build-arm64 docker-x86 docker-arm64

build-x86: $(GO_FILES)
	mkdir -p $(DIST_DIR)
	GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o $(DIST_DIR)/$(APP_NAME)-x86 .

build-arm64: $(GO_FILES)
	mkdir -p $(DIST_DIR)
	GOOS=linux CGO_ENABLED=0 GOARCH=arm64 go build -o $(DIST_DIR)/$(APP_NAME)-arm64 .

docker-x86: build-x86
	docker build -t $(DOCKER_REPO):x86 --build-arg BINARY=$(DIST_DIR)/$(APP_NAME)-x86 -f $(DOCKERFILE) .

docker-arm64: build-arm64
	docker build -t $(DOCKER_REPO):arm64 --build-arg BINARY=$(DIST_DIR)/$(APP_NAME)-arm64 -f $(DOCKERFILE) .

push: docker-x86 docker-arm64
	docker push $(DOCKER_REPO):x86
	docker push $(DOCKER_REPO):arm64

	docker manifest create $(DOCKER_REPO):latest \
		--amend $(DOCKER_REPO):x86 \
		--amend $(DOCKER_REPO):arm64

	docker manifest push $(DOCKER_REPO):latest

clean:
	rm -rf $(DIST_DIR)
	docker rmi -f $(DOCKER_REPO):x86 $(DOCKER_REPO):arm64 2>/dev/null || true
