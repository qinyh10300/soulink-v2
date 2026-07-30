# Soulink Web Backend

Soulink Web Backend 是 Soulink Web 的 Go 服务端，负责认证、资料管理、问卷收集、匹配结果查询、邮件发送以及运行时网关控制等能力。当前项目使用 `Go + PostgreSQL + Redis + Nginx`，并通过 HTTP API 对接前端。

## 当前状态

- `go test ./...` 已验证通过
- API 入口位于 `cmd/api`
- 当前包含受登录保护的 `POST /api/v1/feedback` 接口
- 本地运行依赖 PostgreSQL，Redis 为可选，邮件默认可使用 `log` 驱动

## 主要功能

- 用户注册、登录、登出、查看当前登录用户
- 注册验证码、忘记密码验证码、修改密码验证码、注销账号验证码
- 个人资料读写
- 联系方式读写
- 当前问卷读取、进度读取、单题保存、批量保存
- 匹配结果查询、历史查询、周期状态查询
- 反馈提交接口
- 邮件发送：`log`、单账号 SMTP、多账号轮询 SMTP
- Redis 缓存：验证码、匹配结果、注册人数统计
- 运行时写接口封闭控制
- 自动数据库迁移

## 目录结构

- `cmd/api`：API 服务启动入口
- `cmd/gatectl`：控制写接口封闭时间窗
- `cmd/mailprobe`：逐账号验证 SMTP 配置
- `cmd/migrate`：手动执行数据库迁移
- `configs/mail`：SMTP 账号文件示例
- `configs/match`：匹配结果导入配置
- `configs/questionnaire`：问卷模板与目录配置
- `db/migrations`：数据库迁移
- `deploy/nginx`：Nginx 反向代理配置
- `docs/frontend_api.md`：前端接口文档
- `internal/auth`：认证与验证码逻辑
- `internal/feedback`：反馈提交逻辑
- `internal/gate`：运行时网关控制
- `internal/match`：匹配查询与缓存逻辑
- `internal/meta`：元数据与注册人数统计
- `internal/profile`：资料与联系方式逻辑
- `internal/questionnaire`：问卷解析与保存逻辑
- `internal/store`：数据库访问层

## 环境要求

- Go `1.26.1+`
- Docker Desktop
- PowerShell

## 快速启动

### 1. 准备环境变量

```powershell
Copy-Item .env.example .env
```

如需本地 SMTP 多账号轮询，建议另外复制一份账号文件：

```powershell
Copy-Item configs\mail\smtp_accounts.example.json configs\mail\smtp_accounts.local.json
```

然后在 `.env` 中把：

```env
SMTP_ACCOUNTS_FILE=configs/mail/smtp_accounts.local.json
```

如需启用 `soulagent` 画像生成，可继续在 `.env` 中配置。`provider` 请填名称，不要填 URL，例如：

```env
SOULAGENT_PROVIDER=deepseek
SOULAGENT_API_KEY=your-api-key
SOULAGENT_MODEL=deepseek-chat
```

如果使用硅基流动，可以改为：

```env
SOULAGENT_PROVIDER=siliconflow
SOULAGENT_API_KEY=your-api-key
SOULAGENT_MODEL=deepseek-ai/DeepSeek-V3
```

如果你有多个 key，希望轮询容错，也可以改为：

```env
SOULAGENT_PROVIDER=deepseek
SOULAGENT_API_KEYS=key-1,key-2,key-3
SOULAGENT_MODEL=deepseek-chat
```

如果你想按 provider + apiKey + model 三元组轮询，也可以改为：

```env
SOULAGENT_MODEL_POOL=[{"provider":"deepseek","apiKey":"key-1","model":"deepseek-chat"},{"provider":"siliconflow","apiKey":"key-2","model":"deepseek-ai/DeepSeek-V3"}]
```

配置 `SOULAGENT_MODEL_POOL` 后，会优先于单 provider 配置生效；每次画像调用都会取当前元素并把指针加一。

### 2. 启动依赖

```powershell
docker compose up -d postgres redis mailpit
```

默认端口：

- PostgreSQL：`127.0.0.1:15432`
- Redis：`127.0.0.1:16379`
- Mailpit SMTP：`127.0.0.1:1025`
- Mailpit Web：`http://localhost:8025`

### 3. 启动 API

```powershell
go run .\cmd\api
```

启动后可访问：

- 健康检查：`http://localhost:8080/healthz`
- API 前缀：`http://localhost:8080/api/v1`

## Docker 一体化运行

```powershell
docker compose up -d --build
```

启动后可访问：

- Nginx：`http://localhost:18080`
- 健康检查：`http://localhost:18080/healthz`
- API 前缀：`http://localhost:18080/api/v1`

## 关键配置

### 基础配置

```env
APP_ENV=development
HTTP_ADDR=:8080
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

### 数据库

```env
DB_DSN=postgres://solink:soulink@127.0.0.1:15432/soulink-web?sslmode=disable
DB_AUTO_MIGRATE=true
```

### Redis

```env
REDIS_ENABLED=false
REDIS_ADDR=127.0.0.1:16379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_PREFIX=soulink:
```

### 鉴权

```env
JWT_SECRET=change-me-in-production
JWT_TTL=72h
VERIFICATION_CODE_TTL=5m
VERIFICATION_CODE_COOLDOWN=60s
BCRYPT_COST=12
```

### 邮件

单账号 SMTP：

```env
MAIL_DRIVER=smtp
MAIL_FROM_NAME=Soulink
MAIL_FROM_ADDRESS=noreply@example.com
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=noreply@example.com
SMTP_PASSWORD=your-app-password
SMTP_MAX_ATTEMPTS=3
```

### 3. 多账号轮换 SMTP

在 `configs/mail/` 下新建 `smtp_accounts.local.json`（勿提交 Git），并在 `.env` 中指定：

```env
MAIL_DRIVER=smtp
SMTP_ACCOUNTS_FILE=configs/mail/smtp_accounts.local.json
SMTP_MAX_ATTEMPTS=3
```

JSON 格式示例见下文 `accounts` 数组。默认开发环境也可先用单账号或：

```env
MAIL_DRIVER=log
MAIL_FROM_NAME=Soulink
MAIL_FROM_ADDRESS=noreply@soulink.local
```

### 问卷与匹配结果公布

```env
QUESTIONNAIRE_CATALOG_FILE=thu_questionnaire_template.md
MATCH_RESULT_CACHE_TTL=24h
```

如果你切换到新的问卷 Markdown，并希望把旧版本中的重复题答案迁移到新版本：

1. 将 `.env` 中的 `QUESTIONNAIRE_CATALOG_FILE` 改为新的问卷文件名。
2. 复制 [configs/questionnaire/answer_migration.example.json](/e:/Desktop/QYfile/program/Soulink/src/Soulink-Web-Backend/configs/questionnaire/answer_migration.example.json) 为你自己的配置文件。
3. 在迁移配置里填写旧问卷文件、新问卷文件，以及需要保留的重复题映射关系。
4. 执行迁移命令：

```powershell
go run .\cmd\questionnairemigrate -config .\configs\questionnaire\answer_migration.local.json
```

迁移脚本会：
- 保留旧版本问卷数据不动
- 只把映射表中的题目答案迁移到新版本
- 迁移前清空该用户在目标版本下已有的答案、状态和画像，支持重复执行
- 按新版本必答题重新计算完成状态，因此用户切到新版本后可能会看到“未完成”

### SoulAgent 画像生成

```env
SOULAGENT_PROVIDER=deepseek
SOULAGENT_API_KEY=your-api-key
SOULAGENT_MODEL=deepseek-chat
```

或使用硅基流动：

```env
SOULAGENT_PROVIDER=siliconflow
SOULAGENT_API_KEY=your-api-key
SOULAGENT_MODEL=deepseek-ai/DeepSeek-V3
```

如果需要多个 key 轮询容错：

```env
SOULAGENT_PROVIDER=deepseek
SOULAGENT_API_KEYS=key-1,key-2,key-3
SOULAGENT_MODEL=deepseek-chat
```

如果需要按 provider + apiKey + model 三元组轮询：

```env
SOULAGENT_MODEL_POOL=[{"provider":"deepseek","apiKey":"key-1","model":"deepseek-chat"},{"provider":"siliconflow","apiKey":"key-2","model":"deepseek-ai/DeepSeek-V3"}]
```

说明：

- `SOULAGENT_PROVIDER` 填 provider 名称，不要填 base URL
- 当前支持 `deepseek`、`siliconflow`、`zhipu`、`openrouter`、`groq`
- `SOULAGENT_MODEL_POOL` 优先级高于 `SOULAGENT_PROVIDER`、`SOULAGENT_API_KEY(S)`、`SOULAGENT_MODEL`
- 每次画像调用都会从 `SOULAGENT_MODEL_POOL` 取当前元素，并把轮询指针加一
- 未配置 `SOULAGENT_MODEL_POOL` 时，`SOULAGENT_API_KEYS` 优先级高于 `SOULAGENT_API_KEY`
- `go run .\cmd\api` 会继承 `.env` 中这些变量
- `docker compose up -d --build` 会把这些变量注入 `api` 容器
- 不再需要额外创建 `soulagent/secrets.json`

## 常用命令

```powershell
go test ./...
```

```powershell
go run .\cmd\migrate
```

```powershell
go run .\cmd\mailprobe -to your_mail@example.com
```

```powershell
go run .\cmd\gatectl -action status -gate matching_write
```

## 主要接口

认证相关：

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/register/send-code`
- `POST /api/v1/auth/register/verify-code`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/password/forgot/send-code`
- `POST /api/v1/auth/password/forgot/reset`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

用户相关：

- `GET /api/v1/users/me/profile`
- `PUT /api/v1/users/me/profile`
- `GET /api/v1/users/me/contacts`
- `PUT /api/v1/users/me/contacts/{contactType}`
- `DELETE /api/v1/users/me/contacts/{contactType}`
- `POST /api/v1/users/me/password/send-code`
- `POST /api/v1/users/me/password`
- `POST /api/v1/users/me/delete/send-code`
- `DELETE /api/v1/users/me`

元数据：

- `GET /api/v1/meta/departments`
- `GET /api/v1/meta/grades`
- `GET /api/v1/meta/contact-types`
- `GET /api/v1/meta/registered-user-count`

问卷：

- `GET /api/v1/questionnaires/current`
- `GET /api/v1/questionnaires/current/progress`
- `PUT /api/v1/questionnaires/current/answers`
- `PUT /api/v1/questionnaires/current/answers/{questionId}`

匹配：

- `GET /api/v1/matches/current`
- `GET /api/v1/matches/history`
- `GET /api/v1/matches/{matchId}`
- `GET /api/v1/matches/cycle/status`

反馈：

- `POST /api/v1/feedback`

更多接口说明见 `docs/frontend_api.md`。
