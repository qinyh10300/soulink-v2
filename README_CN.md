# SouLink V2

> 在清华，找到你下一位真正值得见面的人。

SouLink V2 是一个仅面向清华大学的 AI 原生校内匹配平台。用户先描述当下想认识谁、为什么想认识，系统再把自然语言整理成可匹配的意图，给出少量、可解释的推荐，并通过双向同意、匹配房间和联系方式解锁，把线上推荐推进到一次真实的校园见面。

它不是人员目录，也不是无限刷人的交友 Feed。产品当前支持：

- 饭搭子、学习搭子、运动搭子、活动搭子和聊天对象；
- 严肃、主动选择加入的关系匹配；
- 项目伙伴、联合创始人和 CTO 匹配；
- 意图解释、候选推荐、匹配原因与不确定性说明；
- 双向同意、校内见面计划、站内消息和联系方式解锁；
- 举报、拉黑、账号删除与管理员审核。

完整需求见 [V2 Web 设计需求](docs/SOULINK_V2_WEB_DESIGN_REQUIREMENTS.md)，实现覆盖情况见 [V2 实现说明](docs/SOULINK_V2_IMPLEMENTATION.md)。

## 本地启动

依赖：

- Docker Desktop 或兼容的 Docker Engine；
- Docker Compose v2。

在仓库根目录运行：

```bash
docker compose up -d --build
```

服务启动后访问：

| 服务 | 地址 |
| --- | --- |
| SouLink Web | http://localhost:18080 |
| API 健康检查 | http://localhost:18081/healthz |
| Mailpit 本地邮箱 | http://localhost:8025 |
| PostgreSQL | `127.0.0.1:15432` |
| Redis | `127.0.0.1:16379` |

本地 Compose 会自动执行数据库迁移，并生成一套可重复使用的清华校园演示数据：

```text
账号：demo@mails.tsinghua.edu.cn
密码：Soulink2026!
```

登录页可以一键填入演示账号。演示工作区包含完整资料、意图 Feed、项目 memo、候选推荐、待处理申请、已建立的匹配房间、会面计划和聊天记录。种子数据是幂等的，重复启动不会复制记录。

查看状态和日志：

```bash
docker compose ps
docker compose logs -f api
```

停止服务：

```bash
docker compose down
```

如需同时清空本地数据库：

```bash
docker compose down -v
```

## 核心产品链路

1. 使用清华邮箱注册或登录。
2. 完善匹配资料、技能、时间和边界。
3. 用一句自然语言描述想找的人。
4. 回答最多三个关键澄清问题并确认结构化意图。
5. 发布到私密搜索、候选人可见或校园机会 Feed。
6. 查看最多三个推荐及具体的契合点、不确定性和共同背景。
7. 双方表达兴趣后进入独立匹配房间。
8. 在房间内约时间地点、聊天并分别决定是否公开联系方式。
9. 会面后反馈；需要时举报或拉黑。

“找 CTO / 项目伙伴”会额外使用结构化项目 memo，包括项目阶段、每周投入、回报预期、必需与加分技能、协作方式和风险偏好。

## 技术架构

| 层级 | 技术 | 职责 |
| --- | --- | --- |
| Web | React 19、TypeScript、Vite 8、Tailwind CSS 4 | 意图创建、机会 Feed、推荐、匹配房间、资料与安全后台 |
| API | Go 1.26、Chi、JWT、WebSocket | 认证、意图生命周期、推荐、同意状态、聊天、安全治理 |
| 数据 | PostgreSQL 16 | 用户、资料、意图、兴趣、匹配房间、消息、举报与拉黑 |
| 缓存 | Redis 7 | 会话和运行时缓存 |
| 本地邮件 | Mailpit | 接收开发环境验证码邮件 |
| 交付 | Docker Compose、Nginx | 一键构建、迁移、反向代理和健康检查 |

V2 的意图解析与推荐目前采用可测试的确定性规则和显式资料信号，即使没有外部模型密钥也能完整运行。AI 输出会标注推断、匹配理由和不确定性，不代替用户承诺，也不会自动以用户身份交流。

## 仓库结构

```text
backend/      Go API、数据库迁移和测试
frontend/     React Web 应用
matching/     V1 周期匹配引擎，作为兼容能力保留
docs/         V2 产品需求、实现说明、迁移与隐私文档
deploy/       生产部署辅助文件
```

V2 的主要数据迁移位于：

- `000016_v2_intent_matching.sql`
- `000017_safety_moderation.sql`
- `000018_project_memos.sql`

## 开发验证

后端：

```bash
cd backend
go test ./...
go vet ./...
```

前端：

```bash
cd frontend
npm ci
npm run lint
npm run build
```

Compose：

```bash
docker compose config
docker compose up -d --build
docker compose ps
```

## 配置与上线

本地演示由 `DEMO_SEED_ENABLED=true` 开启。程序会在 `APP_ENV=production` 时拒绝启用演示种子，但正式部署仍必须：

- 使用强随机 `JWT_SECRET`；
- 关闭演示种子并移除演示管理员账号；
- 配置真实 SMTP 服务和 HTTPS；
- 收紧 `ALLOWED_ORIGINS`；
- 使用独立的 PostgreSQL、Redis 凭据和持久化备份；
- 明确举报处理、数据删除和隐私响应流程。

环境变量样例见 [backend/.env.example](backend/.env.example)，后端细节见 [backend/README.md](backend/README.md)。

## 产品原则

- 意图优先于浏览人；
- 少量推荐优先于无限滚动；
- 双向同意优先于提前暴露；
- AI 是牵线者，不是假扮用户的代理；
- 真实见面优先于屏幕停留时长；
- 信任、安全和可逆选择是核心功能。
