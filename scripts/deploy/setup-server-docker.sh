#!/bin/bash
# 在 Linux 服务器上一次性执行，初始化 Docker + Git push 自动部署。
# 用法: sudo bash setup-server-docker.sh

set -euo pipefail

APP_USER="aisearch"
APP_DIR="/opt/aisearch"
GIT_DIR="/var/git/aisearch.git"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ $EUID -ne 0 ]]; then
	echo "请使用 root 或 sudo 运行此脚本"
	exit 1
fi

log() {
	echo "[setup $(date '+%Y-%m-%d %H:%M:%S')] $*"
}

log "检查 Docker"
if ! command -v docker &>/dev/null; then
	log "安装 Docker（官方脚本）"
	curl -fsSL https://get.docker.com | sh
fi

if ! docker compose version &>/dev/null; then
	log "ERROR: 需要 Docker Compose 插件，请升级 Docker"
	exit 1
fi

log "创建运行用户 $APP_USER"
if ! id "$APP_USER" &>/dev/null; then
	useradd --system --home-dir "$APP_DIR" --shell /bin/bash "$APP_USER"
fi
usermod -aG docker "$APP_USER"

mkdir -p "$APP_DIR/.ssh" "$APP_DIR/log" "$GIT_DIR"
chmod 700 "$APP_DIR/.ssh"
chown -R "$APP_USER:$APP_USER" "$APP_DIR" "$GIT_DIR"

log "初始化 bare 仓库"
if [[ ! -f "$GIT_DIR/HEAD" ]]; then
	git init --bare "$GIT_DIR"
fi

log "安装 Docker 版 post-receive hook"
install -m 755 "$SCRIPT_DIR/post-receive.docker" "$GIT_DIR/hooks/post-receive"
chown -R "$APP_USER:$APP_USER" "$GIT_DIR"

if [[ ! -f "$APP_DIR/.env" ]]; then
	log "创建 $APP_DIR/.env 模板"
	cat > "$APP_DIR/.env" <<'EOF'
APP_PORT=8081
MYSQL_PASSWORD=your_mysql_password
REDIS_PASSWORD=your_redis_password
OPENAI_API_KEY=your_openai_key
DEEPSEEK_API_KEY=your_deepseek_key
PASSWORD_PEPPER=change-me-to-random-string
SNOWFLAKE_WORKER_ID=1
EOF
	chown "$APP_USER:$APP_USER" "$APP_DIR/.env"
	chmod 600 "$APP_DIR/.env"
fi

log "完成。接下来："
echo "  1. 将本机 SSH 公钥写入 $APP_DIR/.ssh/authorized_keys"
echo "     示例: ssh-copy-id -i ~/.ssh/id_ed25519.pub $APP_USER@服务器IP"
echo "  2. 编辑 $APP_DIR/.env 填入生产环境变量"
echo "  3. 本地 .deploy.env 设置 DEPLOY_MODE=docker"
echo "  4. 本地执行: make deploy-init && make push"
echo "  5. 首次 push 后编辑 $APP_DIR/manifest/config/config.prod.yaml（MySQL/Redis 地址）"
echo "     若 MySQL/Redis 在宿主机，host 填 host.docker.internal"
echo "  6. 再 make push 一次完成部署"
