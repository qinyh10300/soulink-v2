#!/bin/bash
set -e

# ─────────────────────────────────────────────────
# Soulink 全栈部署（Soulink-Web + Soulink-Web-Backend）
#
# 服务器不能 docker pull 时：本地构建镜像并上传 tar；小配置用 pack-config 一次上传。
#
# 本地（有网络）:
#   pack-web / pack-api [user@host]   分别构建并导出 soulink-web.tar、soulink-api.tar（大文件，可分开传）
#   pack-config [user@host]           一次上传所有小配置文件：compose、deploy.sh、.env、configs/mail
#   pack [user@host]                  快捷：web+api 两 tar + pack-config
#   pack-infra [user@host]            postgres+redis → soulink-infra.tar
#   remote-update [user@host]         SSH 在服务器执行 update
#
# 服务器（无需外网拉镜像）:
#   bash deploy.sh setup              首次：load 镜像 + 检查 SSL + 启动
#   bash deploy.sh update             加载新 tar + 切换容器
#   bash deploy.sh status | logs | stop | restart
#
# 交叉架构：服务器 amd64、本机 ARM 时打包前执行:
#   export DOCKER_DEFAULT_PLATFORM=linux/amd64
# ─────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WEB_DIR="$REPO_ROOT/frontend"
BACKEND_DIR="$REPO_ROOT/backend"

IMAGE_WEB="soulink-web:latest"
IMAGE_API="soulink-api:latest"
IMAGE_POSTGRES="postgres:16"
IMAGE_REDIS="redis:7-alpine"

# 应用镜像（可分开发布）；基础镜像（Postgres/Redis 一般只需上传一次）
WEB_TAR="soulink-web.tar"
API_TAR="soulink-api.tar"
APP_TAR="soulink-app.tar"
INFRA_TAR="soulink-infra.tar"
LEGACY_STACK_TAR="soulink-stack.tar"

REMOTE_DIR="/opt/soulink"
SSL_DIR="/root/ssl"
# pack 上传生产环境变量（与 docker-compose 同目录的 .env）
ENV_FILE="$SCRIPT_DIR/.env"

# 默认不保存任何服务器地址。通过参数或 DEPLOY_REMOTE_HOST 显式传入。
DEFAULT_REMOTE_HOST=""

ACCESS_HOST=""

resolve_remote_host() {
    echo "${1:-${DEPLOY_REMOTE_HOST:-$DEFAULT_REMOTE_HOST}}"
}

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
err()  { echo -e "${RED}[✗]${NC} $1"; exit 1; }
info() { echo -e "${CYAN}[i]${NC} $1"; }

check_docker() {
    command -v docker >/dev/null 2>&1 || err "Docker 未安装，请先安装: https://docs.docker.com/engine/install/"
}

compose() {
    if docker compose version >/dev/null 2>&1; then
        docker compose "$@"
    elif command -v docker-compose >/dev/null 2>&1; then
        docker-compose "$@"
    else
        err "未检测到 Docker Compose。请安装 docker compose 插件或 docker-compose"
    fi
}

compose_is_v2_plugin() {
    docker compose version >/dev/null 2>&1
}

# Compose V1（Python docker-compose 1.29 等）对新镜像执行「原地重建」会 KeyError: ContainerConfig；
# 非 V2 时一律先 down 再 up。V2 可把额外参数传给 `docker compose up -d`。
compose_bring_up() {
    if compose_is_v2_plugin; then
        compose up -d "$@"
    else
        warn "当前为 docker-compose 1.x：将先 down 再 up，避免 ContainerConfig 错误（建议安装 docker-compose-plugin，使用 docker compose）。"
        compose down
        compose up -d
    fi
}

access_display_host() {
    if [ -n "$ACCESS_HOST" ]; then
        echo "$ACCESS_HOST"
        return
    fi
    hostname -I | awk '{print $1}'
}

show_access_urls() {
    local host
    host=$(access_display_host)
    info "访问地址:"
    echo -e "  站点 HTTPS: ${GREEN}https://$host/${NC}"
    echo -e "  API 同域:   ${GREEN}https://$host/api/...${NC}"
    echo -e "  HTTP:       ${GREEN}http://$host/${NC}（会跳转 HTTPS，/health 除外）"
    if [[ "$host" =~ ^10\.|^172\.(1[6-9]|2[0-9]|3[0-1])\.|^192\.168\. ]]; then
        info "当前为内网 IP；外网请用公网 IP/域名，或在脚本顶部设置 ACCESS_HOST"
    fi
}

check_ssl() {
    if [ ! -f "$SSL_DIR/cert.pem" ] || [ ! -f "$SSL_DIR/key.pem" ]; then
        err "SSL 证书未找到，请将证书放到 $SSL_DIR/ 目录下:\n  $SSL_DIR/cert.pem  (证书文件)\n  $SSL_DIR/key.pem   (私钥文件)"
    fi
    log "SSL 证书已就绪"
}

check_web_paths() {
    [ -d "$WEB_DIR" ] || err "未找到前端目录: $WEB_DIR"
    [ -f "$WEB_DIR/Dockerfile" ] || err "未找到 $WEB_DIR/Dockerfile"
}

check_backend_paths() {
    [ -d "$BACKEND_DIR" ] || err "未找到后端目录: $BACKEND_DIR"
    [ -f "$BACKEND_DIR/Dockerfile" ] || err "未找到 $BACKEND_DIR/Dockerfile"
}

resolve_mail_src() {
    if [ -d "$SCRIPT_DIR/configs/mail" ]; then
        local n
        n=$(find "$SCRIPT_DIR/configs/mail" -mindepth 1 -maxdepth 1 2>/dev/null | wc -l)
        if [ "${n// /}" -gt 0 ]; then
            echo "$SCRIPT_DIR/configs/mail"
            return
        fi
    fi
    [ -d "$BACKEND_DIR/configs/mail" ] || err "未找到邮件配置目录: $BACKEND_DIR/configs/mail"
    echo "$BACKEND_DIR/configs/mail"
}

# ─── 小配置文件：一个命令全部上传 ─────────────────────────────────────

upload_all_config() {
    local REMOTE_HOST="$1"
    if [ ! -f "$ENV_FILE" ]; then
        err "未找到 $ENV_FILE（可从 deploy/.env.example 复制）"
    fi
    check_backend_paths
    local MAIL_SRC
    MAIL_SRC=$(resolve_mail_src)
    if [[ "$MAIL_SRC" == "$SCRIPT_DIR/configs/mail" ]]; then
        log "邮件配置来源: $MAIL_SRC"
    else
        info "邮件配置来源: backend/configs/mail（示例 JSON 非生产密钥）"
    fi

    ssh "$REMOTE_HOST" "mkdir -p $REMOTE_DIR/configs"
    scp "$SCRIPT_DIR/docker-compose.yml" "$REMOTE_HOST:$REMOTE_DIR/docker-compose.yml"
    scp "$SCRIPT_DIR/deploy.sh" "$REMOTE_HOST:$REMOTE_DIR/deploy.sh"
    scp "$ENV_FILE" "$REMOTE_HOST:$REMOTE_DIR/.env"
    scp -r "$MAIL_SRC" "$REMOTE_HOST:$REMOTE_DIR/configs/"
    log "已上传：docker-compose.yml、deploy.sh、.env、configs/mail"
}

cmd_pack_web() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════${NC}"
    echo -e "${CYAN}   构建并导出前端镜像${NC}"
    echo -e "${CYAN}═══════════════════════════════════════${NC}"
    echo ""

    check_docker
    check_web_paths
    [ -n "${DOCKER_DEFAULT_PLATFORM:-}" ] && info "使用平台: $DOCKER_DEFAULT_PLATFORM（交叉构建）"

    log "构建 $IMAGE_WEB ..."
    docker build -t "$IMAGE_WEB" "$WEB_DIR"
    log "导出 $WEB_TAR ..."
    docker save -o "$WEB_TAR" "$IMAGE_WEB"
    log "完成: $REPO_ROOT/$WEB_TAR ($(du -h "$WEB_TAR" | cut -f1))"

    REMOTE_HOST=$(resolve_remote_host "$1")
    if [ -n "$REMOTE_HOST" ]; then
        scp "$WEB_TAR" "$REMOTE_HOST:$REMOTE_DIR/$WEB_TAR"
        log "已上传 $WEB_TAR → $REMOTE_HOST:$REMOTE_DIR/"
    fi
}

cmd_pack_api() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════${NC}"
    echo -e "${CYAN}   构建并导出 API 镜像${NC}"
    echo -e "${CYAN}═══════════════════════════════════════${NC}"
    echo ""

    check_docker
    check_backend_paths
    [ -n "${DOCKER_DEFAULT_PLATFORM:-}" ] && info "使用平台: $DOCKER_DEFAULT_PLATFORM（交叉构建）"

    log "构建 $IMAGE_API ..."
    docker build -t "$IMAGE_API" "$BACKEND_DIR"
    log "导出 $API_TAR ..."
    docker save -o "$API_TAR" "$IMAGE_API"
    log "完成: $REPO_ROOT/$API_TAR ($(du -h "$API_TAR" | cut -f1))"

    REMOTE_HOST=$(resolve_remote_host "$1")
    if [ -n "$REMOTE_HOST" ]; then
        scp "$API_TAR" "$REMOTE_HOST:$REMOTE_DIR/$API_TAR"
        log "已上传 $API_TAR → $REMOTE_HOST:$REMOTE_DIR/"
    fi
}

cmd_pack_config() {
    REMOTE_HOST=$(resolve_remote_host "$1")
    [ -n "$REMOTE_HOST" ] || err "请指定 user@host，例如: bash deploy.sh pack-config root@服务器"
    echo ""
    echo -e "${CYAN}上传小配置文件（compose、deploy.sh、.env、configs/mail）${NC}"
    echo ""
    upload_all_config "$REMOTE_HOST"
}

# 快捷：两镜像 + 全套配置（等价于依次 pack-web、pack-api、pack-config）
cmd_pack() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════${NC}"
    echo -e "${CYAN}   快捷：web + api 镜像 + 全套配置${NC}"
    echo -e "${CYAN}═══════════════════════════════════════${NC}"
    echo ""

    check_docker
    check_web_paths
    check_backend_paths
    [ -n "${DOCKER_DEFAULT_PLATFORM:-}" ] && info "使用平台: $DOCKER_DEFAULT_PLATFORM（交叉构建）"

    log "构建 $IMAGE_WEB ..."
    docker build -t "$IMAGE_WEB" "$WEB_DIR"
    log "构建 $IMAGE_API ..."
    docker build -t "$IMAGE_API" "$BACKEND_DIR"
    log "导出 $WEB_TAR ..."
    docker save -o "$WEB_TAR" "$IMAGE_WEB"
    log "导出 $API_TAR ..."
    docker save -o "$API_TAR" "$IMAGE_API"
    log "本地产物: $WEB_TAR、$API_TAR（可选再合并旧版单包: docker save -o $APP_TAR $IMAGE_WEB $IMAGE_API）"

    REMOTE_HOST=$(resolve_remote_host "$1")
    if [ -n "$REMOTE_HOST" ]; then
        echo ""
        log "上传到 $REMOTE_HOST:$REMOTE_DIR ..."
        ssh "$REMOTE_HOST" "mkdir -p $REMOTE_DIR"
        scp "$WEB_TAR" "$REMOTE_HOST:$REMOTE_DIR/$WEB_TAR"
        scp "$API_TAR" "$REMOTE_HOST:$REMOTE_DIR/$API_TAR"
        upload_all_config "$REMOTE_HOST"
        log "上传完成。"
        echo ""
        info "首次部署还需: ${CYAN}bash deploy.sh pack-infra [user@host]${NC} → $INFRA_TAR"
        info "服务器: ${CYAN}cd $REMOTE_DIR && bash deploy.sh setup${NC}"
    else
        echo ""
        info "未指定上传目标。请 scp $WEB_TAR、$API_TAR，小配置可: bash deploy.sh pack-config user@host"
    fi
}

# ─── 本地：基础镜像包（postgres + redis），通常只需打包/上传一次 ─────────────

cmd_pack_infra() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════${NC}"
    echo -e "${CYAN}   拉取并打包基础镜像（postgres + redis）${NC}"
    echo -e "${CYAN}═══════════════════════════════════════${NC}"
    echo ""

    check_docker

    log "拉取 $IMAGE_POSTGRES 与 $IMAGE_REDIS ..."
    docker pull "$IMAGE_POSTGRES"
    docker pull "$IMAGE_REDIS"

    log "导出到 $INFRA_TAR ..."
    docker save -o "$INFRA_TAR" "$IMAGE_POSTGRES" "$IMAGE_REDIS"

    SIZE=$(du -h "$INFRA_TAR" | cut -f1)
    log "打包完成: $REPO_ROOT/$INFRA_TAR ($SIZE)"

    REMOTE_HOST=$(resolve_remote_host "$1")
    if [ -n "$REMOTE_HOST" ]; then
        echo ""
        log "上传 $INFRA_TAR 到 $REMOTE_HOST:$REMOTE_DIR ..."
        ssh "$REMOTE_HOST" "mkdir -p $REMOTE_DIR"
        scp "$INFRA_TAR" "$REMOTE_HOST:$REMOTE_DIR/$INFRA_TAR"
        log "上传完成。${CYAN}bash deploy.sh setup${NC} 会自动 load 同目录下的 $INFRA_TAR。"
    else
        echo ""
        info "请手动将 $INFRA_TAR 传到服务器 $REMOTE_DIR/（与 $WEB_TAR / $API_TAR 同目录）。"
    fi
}

cmd_remote_update() {
    local REMOTE
    REMOTE=$(resolve_remote_host "$1")
    if [ -z "$REMOTE" ]; then
        err "未指定服务器。设置 DEFAULT_REMOTE_HOST、DEPLOY_REMOTE_HOST，或: bash deploy.sh remote-update user@host"
    fi
    echo ""
    log "SSH 到 $REMOTE 执行 $REMOTE_DIR/deploy.sh update ..."
    ssh -t "$REMOTE" "cd $REMOTE_DIR && bash deploy.sh update"
}

# ─── 服务器命令 ───────────────────────────────────

cmd_setup() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════${NC}"
    echo -e "${CYAN}   Soulink 全栈首次部署${NC}"
    echo -e "${CYAN}═══════════════════════════════════════${NC}"
    echo ""

    check_docker
    check_ssl

    log "创建日志目录..."
    mkdir -p /var/log/soulink-web

    load_tar_if_present() {
        local f="$1"
        if [ -f "$f" ]; then
            log "从 $f 加载镜像..."
            docker load -i "$f"
        fi
    }

    load_tar_if_present "$INFRA_TAR"
    load_tar_if_present "$WEB_TAR"
    load_tar_if_present "$API_TAR"
    load_tar_if_present "$APP_TAR"
    load_tar_if_present "$LEGACY_STACK_TAR"

    image_ok() { docker image inspect "$1" >/dev/null 2>&1; }

    if ! image_ok "$IMAGE_POSTGRES" || ! image_ok "$IMAGE_REDIS"; then
        err "缺少 $IMAGE_POSTGRES 或 $IMAGE_REDIS。\n  请上传 $INFRA_TAR 到本目录后重新 setup，或在有网机器执行: bash deploy.sh pack-infra 再上传。"
    fi
    if ! image_ok "$IMAGE_WEB" || ! image_ok "$IMAGE_API"; then
        err "缺少 $IMAGE_WEB 或 $IMAGE_API。\n  请上传 $WEB_TAR 与 $API_TAR（或合并包 $APP_TAR / 旧版 $LEGACY_STACK_TAR）。"
    fi

    # compose 使用相对路径 ./configs/mail；请在包含 docker-compose.yml 的目录下执行本脚本
    if [ ! -d "configs/mail" ]; then
        warn "未找到 ./configs/mail，API 挂载可能失败。请从仓库上传 backend/configs/mail"
    fi

    log "启动容器..."
    compose_bring_up

    echo ""
    log "部署完成，等待健康检查..."
    sleep 5

    if compose ps | grep -q "Up"; then
        log "服务已启动"
        echo ""
        show_access_urls
    else
        err "服务启动失败，请: bash deploy.sh logs"
    fi
}

cmd_update() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════${NC}"
    echo -e "${CYAN}   Soulink 全栈更新${NC}"
    echo -e "${CYAN}═══════════════════════════════════════${NC}"
    echo ""

    check_docker

    local loaded=0
    if [ -f "$WEB_TAR" ]; then
        log "从 $WEB_TAR 加载..."
        docker load -i "$WEB_TAR"
        loaded=1
    fi
    if [ -f "$API_TAR" ]; then
        log "从 $API_TAR 加载..."
        docker load -i "$API_TAR"
        loaded=1
    fi
    if [ -f "$APP_TAR" ]; then
        log "从 $APP_TAR 加载（web+api 合并包）..."
        docker load -i "$APP_TAR"
        loaded=1
    fi
    if [ -f "$LEGACY_STACK_TAR" ]; then
        log "从 $LEGACY_STACK_TAR 加载（旧版单包）..."
        docker load -i "$LEGACY_STACK_TAR"
        loaded=1
    fi
    if [ "$loaded" -eq 0 ]; then
        err "未找到任何应用镜像包（$WEB_TAR / $API_TAR / $APP_TAR / $LEGACY_STACK_TAR）。本地执行 pack-web、pack-api 或 pack 后上传。"
    fi

    log "切换到新版本..."
    compose_bring_up --force-recreate --remove-orphans

    log "清理悬空镜像..."
    docker image prune -f > /dev/null 2>&1 || true

    sleep 5
    if compose ps | grep -q "Up"; then
        log "更新完成"
        echo ""
        show_access_urls
        cmd_status
    else
        err "更新后异常，请: bash deploy.sh logs"
    fi
}

cmd_status() {
    echo ""
    info "容器状态:"
    compose ps
    echo ""
    info "最近日志（各服务末 3 行）:"
    compose logs --tail 3
}

cmd_logs() {
    LOG_DIR="/var/log/soulink-web"
    if [ -f "$LOG_DIR/access.log" ]; then
        info "Nginx 访问/错误日志（Ctrl+C 退出）:"
        tail -f "$LOG_DIR/access.log" "$LOG_DIR/error.log" 2>/dev/null || tail -f "$LOG_DIR/access.log"
    else
        info "容器日志（Ctrl+C 退出）:"
        compose logs -f
    fi
}

cmd_stop() {
    warn "停止全栈服务..."
    compose down
    log "已停止"
}

cmd_restart() {
    log "重启服务..."
    compose restart
    sleep 3
    cmd_status
}

cmd_help() {
    echo ""
    echo "Soulink 全栈部署（前端 + API + Postgres + Redis）"
    echo ""
    echo "用法: bash deploy.sh <命令>"
    echo ""
    echo -e "${CYAN}本地（有网络）:${NC}"
    echo "  pack-web [user@host]       仅前端 → $WEB_TAR"
    echo "  pack-api [user@host]       仅 API → $API_TAR"
    echo "  pack-config [user@host]    一次上传小文件：compose、deploy.sh、.env、configs/mail"
    echo "  pack [user@host]           快捷：web+api 两 tar + pack-config"
    echo "  pack-infra [user@host]     postgres+redis → $INFRA_TAR（一般只需一次）"
    echo "  remote-update [user@host]  SSH 在服务器执行 update"
    echo ""
    echo -e "${CYAN}默认 SSH 目标:${NC}"
    if [ -n "${DEPLOY_REMOTE_HOST:-}" ]; then
        echo -e "  ${GREEN}$DEPLOY_REMOTE_HOST${NC}（环境变量 DEPLOY_REMOTE_HOST）"
    elif [ -n "$DEFAULT_REMOTE_HOST" ]; then
        echo -e "  ${GREEN}$DEFAULT_REMOTE_HOST${NC}（脚本内 DEFAULT_REMOTE_HOST）"
    else
        echo "  （未设置）仅生成本地 tar，不上传"
    fi
    echo ""
    echo -e "${CYAN}服务器:${NC}"
    echo "  setup / update / status / logs / stop / restart"
    echo ""
    echo -e "${CYAN}说明:${NC}"
    echo "  - 证书路径: $SSL_DIR/{cert.pem,key.pem}"
    echo "  - 镜像: $WEB_TAR + $API_TAR（或合并 $APP_TAR / 旧版 $LEGACY_STACK_TAR）+ $INFRA_TAR"
    echo "  - 若遇 KeyError ContainerConfig：请安装 Docker Compose V2（如 apt install docker-compose-plugin）"
    echo "  - 修改数据库密码时同时改 POSTGRES_PASSWORD 与 DB_DSN"
    echo ""
}

# 在服务器上应从 REMOTE_DIR 运行，与各 *.tar 同目录
cd "$SCRIPT_DIR" 2>/dev/null || true

case "${1:-}" in
    pack)           cmd_pack "$2"           ;;
    pack-web)       cmd_pack_web "$2"       ;;
    pack-api)       cmd_pack_api "$2"       ;;
    pack-config)    cmd_pack_config "$2"    ;;
    pack-infra)     cmd_pack_infra "$2"     ;;
    remote-update)  cmd_remote_update "$2"  ;;
    setup)          cmd_setup               ;;
    update)         cmd_update              ;;
    status)         cmd_status              ;;
    logs)           cmd_logs                ;;
    stop)           cmd_stop                ;;
    restart)        cmd_restart             ;;
    *)              cmd_help                ;;
esac
