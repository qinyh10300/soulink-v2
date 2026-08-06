# SouLink 本地部署与验收

更新时间：2026-08-07

本文是 SouLink V2 当前版本的本地部署入口。默认 Compose 只监听本机回环地址，会自动迁移数据库、生成清华演示工作区，并在没有外部模型密钥时使用可复现的本地规则。

## 1. 环境要求

- Docker Desktop，或带 Docker Compose v2 的 Docker Engine；
- 建议至少 4 GB 可用内存；
- 本机端口 `18080`、`18081`、`8025`、`15432`、`16379` 未被占用。

## 2. 一键启动

在仓库根目录执行：

```bash
docker compose up -d --build
docker compose ps
```

全部服务应显示为 `healthy`。首次构建时间取决于 Docker 镜像和 npm、Go 依赖缓存。

| 服务 | 地址 | 用途 |
| --- | --- | --- |
| SouLink Web | http://localhost:18080 | 完整产品界面 |
| API 健康检查 | http://localhost:18081/healthz | API 与数据库状态 |
| Mailpit | http://localhost:8025 | 本地验证码与通知邮件 |
| PostgreSQL | `127.0.0.1:15432` | 本地数据库 |
| Redis | `127.0.0.1:16379` | 缓存与运行时状态 |

## 3. 本地账号

普通演示账号：

```text
邮箱：demo@mails.tsinghua.edu.cn
密码：Soulink2026!
```

Hack 管理员账号：

```text
邮箱：hack@mails.tsinghua.edu.cn
密码：SoulinkHack2026!
```

这两个账号仅由根目录的开发 Compose 创建。API 在 `APP_ENV=production` 时会拒绝启用演示 seed；生产镜像也不会显示一键演示登录。

## 4. 健康检查

```bash
docker compose ps
curl --noproxy '*' http://127.0.0.1:18081/healthz
docker compose logs --tail=100 api
```

API 输出逐行 JSON 日志。每个 HTTP 响应包含 `X-Request-ID`，日志只记录路由模板、状态码、字节数和耗时，不记录正文、邮箱或查询参数。

管理员登录后可在安全与运营后台查看“运行健康”，或调用：

```text
GET /api/v1/admin/operational-metrics
```

该接口只返回进程内聚合值，重启后重新计数，且仅允许 `ADMIN_EMAILS` 或本地 Hack 管理员访问。

## 5. 常用操作

查看日志：

```bash
docker compose logs -f api
docker compose logs -f web
```

停止但保留数据：

```bash
docker compose down
```

重新构建应用：

```bash
docker compose up -d --build api web
```

彻底重置本地工作区：

```bash
docker compose down -v
docker compose up -d --build
```

`down -v` 会永久删除当前 Compose 项目的 PostgreSQL 数据卷，只应用于确认不再需要数据的本地环境。

## 6. 空数据库验收基线

2026-08-07 已从全新 Compose 数据卷验证：

- 迁移自动执行到 `000031_notifications.sql`；
- API、Web、PostgreSQL、Redis 和 Mailpit 均进入 `healthy`；
- 自动生成 6 名演示用户、8 个连接意图、1 个匹配房间、1 个群聊和 3 条树洞帖子；
- demo 与 Hack 账号均可登录；
- 演示匹配房间明确使用 `deterministic / local-rules / demo-seed-v1`，无模型密钥也能运行；
- 重启 seed 幂等，不重复创建核心记录。

## 7. 生产部署边界

生产配置从 [`deploy/.env.example`](../deploy/.env.example) 开始，并使用 [`deploy/docker-compose.yml`](../deploy/docker-compose.yml)。至少需要：

- 设置强随机 `POSTGRES_PASSWORD`、`DB_DSN` 和 `JWT_SECRET`；
- 保持 `APP_ENV=production` 与 `DEMO_SEED_ENABLED=false`；
- 在 `ADMIN_EMAILS` 中显式配置实际审核员，不能使用本地 Hack 账号；
- 配置 HTTPS、真实 SMTP、受限的 `ALLOWED_ORIGINS` 和数据库备份；
- 将 AI 密钥仅放入服务器环境变量，不写入前端构建参数或仓库；
- 建立举报审核、数据删除、密钥轮换和恢复演练流程。

外部模型是可选增强。未配置 `AI_API_KEY` 或 SoulAgent 时，意图解释、推荐、群组撮合与聊天建议仍使用确定性降级路径。
