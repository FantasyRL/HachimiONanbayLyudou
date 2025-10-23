# 辅助工具安装列表
# go install github.com/cloudwego/hertz/cmd/hz@latest
# go install github.com/cloudwego/kitex/tool/cmd/kitex@latest
# go install github.com/hertz-contrib/swagger-generate/thrift-gen-http-swagger@latest

# 默认输出帮助信息
.DEFAULT_GOAL := help

# 纯 Windows 环境：不强制依赖 WSL/Git Bash，默认使用 cmd，必要逻辑用 PowerShell 执行
# 项目 MODULE 名
MODULE = github.com/FantasyRL/HachimiONanbayLyudou
REMOTE_REPOSITORY ?= fantasyrl/hachimionanbaylyudou
# 目录相关（避免在 Windows 下调用 pwd 失败，使用内置 CURDIR）
DIR = $(CURDIR)
CMD = $(DIR)/cmd
CONFIG_PATH = $(DIR)/config
IDL_PATH = $(DIR)/idl
OUTPUT_PATH = $(DIR)/output
API_PATH= $(DIR)/cmd/api
# Docker 网络名称
DOCKER_NET := go-mcp-net
# Docker 镜像前缀和标签
IMAGE_PREFIX ?= hachimi
TAG          ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)

# 服务名
SERVICES := host mcp_local mcp_remote
service = $(word 1, $@)

# hertz HTTP脚手架
# init: hz new -idl ./idl/api.thrift -mod github.com/FantasyRL/HachimiONanbayLyudou -handler_dir ./api/handler -model_dir ./api/model -router_dir ./api/router
.PHONY: hertz-gen-api
hertz-gen-api:
	hz update -idl ${IDL_PATH}/api.thrift; \
	rm -rf $(DIR)/swagger; \
    thriftgo -g go -p http-swagger $(IDL_PATH)/api.thrift; \
    rm -rf $(DIR)/gen-go

.PHONY: $(SERVICES)
$(SERVICES):
	go run $(CMD)/$(service) -cfg $(CONFIG_PATH)/config.yaml

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

# 创建 Docker 网络供容器间HTTP通信
.PHONY: docker-net
docker-net:
ifeq ($(OS),Windows_NT)
	@powershell -NoProfile -ExecutionPolicy Bypass -Command "docker network inspect $(DOCKER_NET) *> $null; if ($$LASTEXITCODE -ne 0) { docker network create $(DOCKER_NET) | Out-Null }"
else
	@docker network inspect $(DOCKER_NET) >/dev/null 2>&1 || docker network create $(DOCKER_NET)
endif

.PHONY: docker-run-%
docker-run-%: docker-build-% docker-net
ifeq ($(OS),Windows_NT)
	@echo ">> Running docker (STRICT config - Windows) on network $(DOCKER_NET)"
	@powershell -NoProfile -ExecutionPolicy Bypass -Command ^
		"$$cfg='$(CONFIG_PATH)\config.yaml'; if (!(Test-Path $$cfg)) { Write-Error 'ERROR: config.yaml not found'; exit 2 } ;" ^
		"docker rm -f $* *> $null ;" ^
		"$${portFlags}=''; if ('$*' -eq 'host') { $${portFlags}='-p 10001:10001' } elseif ('$*' -eq 'mcp_server') { $${portFlags}='-p 10002:10002' } ;" ^
		"docker run --rm -itd --name $* --network $(DOCKER_NET) --network-alias $* $${portFlags} -e SERVICE=$* -e TZ=Asia/Shanghai -v $$cfg:/app/config/config.yaml:ro $(IMAGE_PREFIX)/$*:$(TAG)"
else
	@echo ">> Running docker (STRICT config - Linux) on network $(DOCKER_NET)"
	@CFG_SRC="$(CONFIG_PATH)/config.yaml"; \
	if [ ! -f "$$CFG_SRC" ]; then \
		echo "ERROR: $$CFG_SRC not found. Please create it." >&2; \
		exit 2; \
	fi; \
	docker rm -f $* >/dev/null 2>&1 || true; \
	case "$*" in \
		host) PORT_FLAGS="-p 10001:10001" ;; \
		mcp_local) PORT_FLAGS="-p 10002:10002" ;; \
		mcp_remote) PORT_FLAGS="-p 10003:10003" ;; \
		*) PORT_FLAGS="" ;; \
	esac; \
	docker run --rm -itd \
		--name $* \
		--network host \
		$$PORT_FLAGS \
		-e SERVICE=$* \
		-e TZ=Asia/Shanghai \
		-v "$$CFG_SRC":/app/config/config.yaml:ro \
		$(IMAGE_PREFIX)/$*:$(TAG)
endif


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
		docker run --rm -itd \
			--name $* \
			--network host \
			-e SERVICE=$* \
			-e TZ=Asia/Shanghai \
			-v "$$CFG_SRC":/app/config/config.yaml:ro \
			$(REMOTE_REPOSITORY):$*
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
	go build -o bin/mcp_local ./cmd/mcp_local # windows的output需要是.exe，并且在config.stdio.yaml中修改，bin/mcp-server.exe
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
	cd $(DIR)/docker && docker-compose up -d
