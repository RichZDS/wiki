.PHONY: build run test clean push deploy-init deploy-status deploy-logs

APP_NAME = wiki
BUILD_DIR = build

# 部署配置，复制 .deploy.env.example 为 .deploy.env 后修改
-include .deploy.env
DEPLOY_HOST   ?= deploy@your-server.com
DEPLOY_REPO   ?= /var/git/wiki.git
DEPLOY_BRANCH ?= main
DEPLOY_REMOTE ?= deploy
DEPLOY_MODE   ?= native

build:
	@echo "Building $(APP_NAME)..."
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME).exe .

build-linux:
	@echo "Building $(APP_NAME) for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) .

run:
	go run main.go

run-prod:
	go run main.go prod

test:
	go test ./tests/...

test-verbose:
	go test -v ./tests/...

clean:
	rm -rf $(BUILD_DIR)

# --- 部署（Git push + 服务器 post-receive hook，无需 Jenkins）---

# 添加 deploy 远程仓库（只需执行一次）
deploy-init:
	@if git remote get-url $(DEPLOY_REMOTE) >/dev/null 2>&1; then \
		echo "remote $(DEPLOY_REMOTE) 已存在: $$(git remote get-url $(DEPLOY_REMOTE))"; \
	else \
		git remote add $(DEPLOY_REMOTE) $(DEPLOY_HOST):$(DEPLOY_REPO); \
		echo "已添加 remote $(DEPLOY_REMOTE) -> $(DEPLOY_HOST):$(DEPLOY_REPO)"; \
	fi

# 推送代码到服务器，触发自动编译和重启
push: deploy-init
	git push $(DEPLOY_REMOTE) HEAD:$(DEPLOY_BRANCH)

# 查看服务器上服务状态
deploy-status:
ifeq ($(DEPLOY_MODE),docker)
	ssh $(DEPLOY_HOST) "cd /opt/wiki && docker compose -f docker-compose.prod.yml ps"
else
	ssh $(DEPLOY_HOST) "systemctl status wiki --no-pager"
endif

# 查看服务器上最近日志
deploy-logs:
ifeq ($(DEPLOY_MODE),docker)
	ssh $(DEPLOY_HOST) "cd /opt/wiki && docker compose -f docker-compose.prod.yml logs --tail=50 app"
else
	ssh $(DEPLOY_HOST) "journalctl -u wiki -n 50 --no-pager"
endif
