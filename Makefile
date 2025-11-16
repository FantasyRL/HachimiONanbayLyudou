
.DEFAULT_GOAL := help


MODULE = github.com/FantasyRL/go-mcp-demo
REMOTE_REPOSITORY ?= fantasyrl/go-mcp-demo

DIR = $(CURDIR)
CMD = $(DIR)/cmd
CONFIG_PATH = $(DIR)/config
IDL_PATH = $(DIR)/idl
OUTPUT_PATH = $(DIR)/output
API_PATH= $(DIR)/cmd/api
GEN_CONFIG_PATH ?= $(DIR)/pkg/gorm-gen/generator/etc/config.yaml

DOCKER_NET := docker_mcp_net

IMAGE_PREFIX ?= hachimi
TAG          ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)


SERVICES := host mcp_local mcp_remote
service = $(word 1, $@)

.PHONY: hertz-gen-api
hertz-gen-api:
	hz update -idl ${IDL_PATH}/api.thrift; \
	rm -rf $(DIR)/swagger; \
    thriftgo -g go -p http-swagger $(IDL_PATH)/api.thrift; \
    rm -rf $(DIR)/gen-go

.PHONY: $(SERVICES)
$(SERVICES):
	go run $(CMD)/$(service) -cfg $(CONFIG_PATH)/config.yaml

.PHONY: model
model:
	@echo "Generating database models..."
	go run $(DIR)/pkg/gorm-gen/generator -f $(GEN_CONFIG_PATH)

.PHONY: vendor
vendor:
	@echo ">> go mod tidy && go mod vendor"
	go mod tidy
	go mod vendor

.PHONY: docker-build-%
docker-build-%: vendor
	@echo ">> Building image for service: $* (tag: $(TAG))"
	docker build \
	  --build-arg SERVICE=$* \
	  -f docker/Dockerfile \
	  -t $(IMAGE_PREFIX)/$*:$(TAG) \
	  .

.PHONY: pull-run-%
pull-run-%:
ifeq ($(OS),Windows_NT)
		@echo ">> Pulling and running docker (STRICT config - Windows): $*"
		@docker pull $(REMOTE_REPOSITORY):$*
		@powershell -NoProfile -ExecutionPolicy Bypass -File "$(DIR)\scripts\docker-run.ps1" -Service "$*" -Image "$(REMOTE_REPOSITORY):$*" -ConfigPath "$(CONFIG_PATH)\config.yaml"
else
		@echo ">> Pulling and running docker (STRICT config - Linux): $*"
		@docker pull $(REMOTE_REPOSITORY):$*
		@CFG_SRC="$(CONFIG_PATH)/config.yaml"; \
		if [ ! -f "$$CFG_SRC" ]; then \
			echo "ERROR: $$CFG_SRC not found. Please create it." >&2; \
			exit 2; \
		fi; \
		docker rm -f $* >/dev/null 2>&1 || true; \
		if [ "$*" = "host" ]; then \
			docker run -itd \
				--name $* \
				--network ${DOCKER_NET} \
				-e SERVICE=$* \
				-e TZ=Asia/Shanghai \
				-v "$$CFG_SRC":/app/config/config.yaml:ro \
				-p 10001:10001 \
				$(REMOTE_REPOSITORY):$*; \
		else \
			docker run -itd \
				--name $* \
				--network ${DOCKER_NET} \
				-e SERVICE=$* \
				-e TZ=Asia/Shanghai \
				-v "$$CFG_SRC":/app/config/config.yaml:ro \
				$(REMOTE_REPOSITORY):$*; \
		fi
endif


# 帮助信息
.PHONY: help
help:
	@echo "Available targets:"; \
	echo "  host                 - go run cmd/host with config.yaml"; \
	echo "  mcp_local           - go run cmd/mcp_local with config.yaml"; \
	echo "  vendor               - go mod tidy && vendor"; \
	echo "  docker-build-<svc>   - build image for service (host|mcp_local)"; \
	echo "  docker-run-<svc>     - run container (Windows自动映射端口, Linux使用--network host)"; \
	echo "  pull-run-<svc>       - pull and run container (同上)"; \
	echo "  stdio                - build mcp_local and run host with stdio config"; \
	echo "  push-<svc>           - push image to remote repo"


.PHONY: stdio
stdio:
	go build -o bin/mcp_local ./cmd/mcp_local 
	go run ./cmd/host -cfg $(CONFIG_PATH)/config.stdio.yaml

.PHONY: push-%
push-%:
	@read -p "Confirm service name to push (type '$*' to confirm): " CONFIRM_SERVICE; \
	if [ "$$CONFIRM_SERVICE" != "$*" ]; then \
		echo "Confirmation failed. Expected '$*', but got '$$CONFIRM_SERVICE'."; \
		exit 1; \
	fi; \
	if echo "$(SERVICES)" | grep -wq "$*"; then \
		if [ "$(ARCH)" = "x86_64" ] || [ "$(ARCH)" = "amd64" ]; then \
			echo "Building and pushing $* for amd64 architecture..."; \
			docker build --build-arg SERVICE=$* -t $(REMOTE_REPOSITORY):$* -f docker/Dockerfile .; \
			docker push $(REMOTE_REPOSITORY):$*; \
		else \
			echo "Building and pushing $* using buildx for amd64 architecture..."; \
			docker buildx build --platform linux/amd64 --build-arg SERVICE=$* -t $(REMOTE_REPOSITORY):$* -f docker/Dockerfile --push .; \
		fi; \
	else \
		echo "Service '$*' is not a valid service. Available: [$(SERVICES)]"; \
		exit 1; \
	fi

.PHONY: env
env:
	rm -rf $(DIR)/docker/data/consul ; \
	cd $(DIR)/docker && docker-compose up -d

.PHONY: push-cd-%
push-cd-%: vendor
	@if echo "$(SERVICES)" | grep -wq "$*"; then \
		if [ "$(ARCH)" = "x86_64" ] || [ "$(ARCH)" = "amd64" ]; then \
			echo "Building and pushing $* for amd64 architecture..."; \
			docker build --build-arg SERVICE=$* -t $(REMOTE_REPOSITORY):$* -f docker/Dockerfile .; \
			docker push $(REMOTE_REPOSITORY):$*; \
		else \
			echo "Building and pushing $* using buildx for amd64 architecture..."; \
			docker buildx build --platform linux/amd64 --build-arg SERVICE=$* -t $(REMOTE_REPOSITORY):$* -f docker/Dockerfile --push .; \
		fi; \
	else \
		echo "Service '$*' is not a valid service. Available: [$(SERVICES)]"; \
		exit 1; \
	fi