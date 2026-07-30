# Soulink-Web

SouLink 官网前端，包含落地页、用户注册/登录、问卷、个人中心、控制面板等功能。

## 技术栈

- **React 19** + **TypeScript**
- **Vite 8** 构建
- **Tailwind CSS 4** 样式
- **React Router** 路由
- **Framer Motion** 动画
- **React Hook Form** + **Zod** 表单验证

## 本地开发

```bash
# 安装依赖
npm install

# Default proxy target: http://127.0.0.1:18080
# For a host-run API, set:
# PowerShell: $env:VITE_DEV_API_PROXY_TARGET="http://127.0.0.1:8080"
#
# 启动开发服务器
npm run dev

# 构建生产版本
npm run build

# 代码检查
npm run lint
```

## 项目结构

```
src/
├── app/            # 路由配置
├── assets/         # 静态资源
├── components/     # UI 组件
│   ├── auth/       # 认证相关
│   ├── dashboard/  # 控制面板
│   ├── home/       # 首页
│   ├── layout/     # 布局（导航栏、页脚）
│   ├── profile/    # 个人中心
│   ├── survey/     # 问卷
│   └── ui/         # 通用 UI 组件
├── lib/            # 工具函数
├── mock/           # Mock 数据
├── pages/          # 页面
└── types/          # TypeScript 类型
```

## Docker 部署

服务器**无需外网**：在本地（有网络的机器）用 Docker 构建镜像，导出为 `soulink-web.tar` 后与 `docker-compose.yml`、`deploy.sh` 一并上传到服务器，在服务器上 `docker load` 并启动。

`docker-compose.yml` 使用固定镜像名 `soulink-web:latest`，**不在服务器目录构建**，因此 `/opt/soulink-web` 下不需要 Dockerfile 或前端源码。

### 服务器依赖

- **Docker Engine**
- **Compose 任选其一**：
  - 推荐：**Docker Compose V2 插件**（`docker compose`，如 Ubuntu 22.04 需添加 [Docker 官方 apt 源](https://docs.docker.com/engine/install/ubuntu/) 后安装 `docker-compose-plugin`）
  - 或：旧版 **`docker-compose`**（Python 1.x，`apt install docker-compose`）

脚本会自动优先使用 `docker compose`，否则使用 `docker-compose`。

**说明：** 旧版 `docker-compose` 在 `update` 时若使用 `--force-recreate` 可能对新镜像报 `KeyError: 'ContainerConfig'`。脚本在检测到非 V2 时会改为 **`down` + `up`** 切换版本（约有数秒中断）；安装 Compose V2 插件后可走原地重建，更接近零停机。

### 配置 `deploy.sh`（可选但建议）

在仓库里的 `deploy.sh` 顶部可改：

| 变量 | 作用 |
|------|------|
| `DEFAULT_REMOTE_HOST` | `pack` 不传 `user@host` 时的默认 SSH 上传目标 |
| `ACCESS_HOST` | `setup` / `update` 成功后在终端打印的访问主机（公网 IP 或域名，**不要**带 `https://`）。留空时会在阿里云 ECS 上尝试读元数据里的弹性公网 IP，否则用本机首个网卡 IP（常为内网） |

单次指定其它服务器上传：

```bash
DEPLOY_REMOTE_HOST=user@其它机器 bash deploy.sh pack
```

查看当前默认上传目标等信息：

```bash
bash deploy.sh
```

云厂商安全组需放行 **80 / 443**。

### 首次部署

```bash
# 1. 本地：构建镜像、导出 tar、上传到服务器（可省略参数，使用 DEFAULT_REMOTE_HOST）
bash deploy.sh pack
# 或显式指定：bash deploy.sh pack root@你的服务器

# 2. 服务器：加载镜像 + 检查 SSL + 启动
cd /opt/soulink-web && bash deploy.sh setup
```

### 更新版本

```bash
# 1. 本地：打包并上传
git pull
bash deploy.sh pack

# 2. 任选其一
# 2a. 本地通过 SSH 替你在服务器上执行 update（与 pack 共用 DEFAULT_REMOTE_HOST / DEPLOY_REMOTE_HOST）
bash deploy.sh remote-update

# 2b. 或登录服务器后执行
cd /opt/soulink-web && bash deploy.sh update
```

`setup` / `update` 成功后脚本会打印 HTTPS / HTTP 访问地址（由 `ACCESS_HOST` 或自动探测决定）。

### 本地仅用 Docker 验证（可选）

与服务器一致，需先构建出镜像再启动：

```bash
docker build -t soulink-web:latest .
docker compose up -d
# 若只有 docker-compose：docker-compose up -d
```

### 部署工具命令一览

| 命令 | 环境 | 说明 |
|------|------|------|
| `bash deploy.sh pack [user@host]` | 本地 | 构建并导出镜像；省略参数则用 `DEFAULT_REMOTE_HOST` 或环境变量 `DEPLOY_REMOTE_HOST` 上传 |
| `bash deploy.sh remote-update [user@host]` | 本地 | SSH 到服务器执行 `update`（需已上传新 `soulink-web.tar`） |
| `bash deploy.sh setup` | 服务器 | 首次部署（`docker load` + SSL 检查 + 启动） |
| `bash deploy.sh update` | 服务器 | `docker load` 新 tar 后切换运行中的容器 |
| `bash deploy.sh status` | 服务器 | 查看运行状态与最近日志 |
| `bash deploy.sh logs` | 服务器 | 查看实时日志 |
| `bash deploy.sh restart` | 服务器 | 重启服务 |
| `bash deploy.sh stop` | 服务器 | 停止服务 |

### SSL 配置

将证书文件放在服务器 `/root/ssl/` 目录下：

```
/root/ssl/cert.pem    # 证书文件
/root/ssl/key.pem     # 私钥文件
```

部署强制要求 HTTPS，未配置证书将无法启动。HTTP 80 端口自动 301 跳转到 HTTPS。
