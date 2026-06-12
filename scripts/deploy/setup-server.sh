#!/bin/bash
# 在 Linux 服务器上一次性执行，初始化 Git push 自动部署环境。
# 用法: sudo bash setup-server.sh

set -euo pipefail

APP_USER="aisearch"
APP_DIR="/opt/aisearch"
GIT_DIR="/var/git/aisearch.git"
SERVICE_NAME="aisearch"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ $EUID -ne 0 ]]; then
	echo "请使用 root 或 sudo 运行此脚本"
	exit 1
fi

log() {
	echo "[setup $(date '+%Y-%m-%d %H:%M:%S')] $*"
}

log "创建运行用户 $APP_USER"
if ! id "$APP_USER" &>/dev/null; then
	useradd --system --home-dir "$APP_DIR" --shell /bin/bash "$APP_USER"
fi
mkdir -p "$APP_DIR/.ssh"
chmod 700 "$APP_DIR/.ssh"
chown -R "$APP_USER:$APP_USER" "$APP_DIR/.ssh"

log "创建目录"
mkdir -p "$APP_DIR/log" "$GIT_DIR"
chown -R "$APP_USER:$APP_USER" "$APP_DIR"

log "初始化 bare 仓库"
if [[ ! -f "$GIT_DIR/HEAD" ]]; then
	git init --bare "$GIT_DIR"
fi

log "安装 post-receive hook"
install -m 755 "$SCRIPT_DIR/post-receive" "$GIT_DIR/hooks/post-receive"
chown -R "$APP_USER:$APP_USER" "$GIT_DIR"

log "安装 systemd 服务"
install -m 644 "$SCRIPT_DIR/aisearch.service" "/etc/systemd/system/$SERVICE_NAME.service"
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"

log "配置 sudoers：允许 $APP_USER 无密码重启服务"
SUDOERS_FILE="/etc/sudoers.d/aisearch-deploy"
cat > "$SUDOERS_FILE" <<EOF
$APP_USER ALL=(root) NOPASSWD: /bin/systemctl restart $SERVICE_NAME, /bin/systemctl status $SERVICE_NAME
EOF
chmod 440 "$SUDOERS_FILE"

log "设置 bare 仓库归属"
chown -R "$APP_USER:$APP_USER" "$GIT_DIR"

if [[ ! -f "$APP_DIR/.env" ]]; then
	log "创建 $APP_DIR/.env 模板（请编辑填入真实密钥）"
	cat > "$APP_DIR/.env" <<'EOF'
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
echo "  1. 确保服务器已安装 Go 1.25+: go version"
echo "  2. 将本机 SSH 公钥写入服务器 $APP_DIR/.ssh/authorized_keys"
echo "     示例: ssh-copy-id -i ~/.ssh/id_ed25519.pub $APP_USER@服务器IP"
echo "  3. 编辑 $APP_DIR/.env 填入生产环境变量"
echo "  4. 本地复制 .deploy.env.example 为 .deploy.env，DEPLOY_HOST 填 $APP_USER@服务器IP"
echo "  5. 本地执行: make deploy-init && make push"
echo "  6. 首次 push 后编辑 $APP_DIR/config.prod.yaml，再 make push 一次"
echo "  7. 启动服务: systemctl start $SERVICE_NAME"
