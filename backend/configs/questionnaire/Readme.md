# 问卷版本迁移说明

## 适用场景

当你发布新版本问卷，并希望保留旧版本中一部分重复题目的答案时，可以使用问卷迁移脚本。脚本会按照映射表，把旧问卷中的指定题目答案迁移到新问卷版本，其余题目保持为空，并按新问卷重新计算完成状态。

本文档默认你使用仓库内的发布脚本：

[`scripts/deploy_questionnaire_migrate.sh`](/e:/Desktop/QYfile/program/Soulink/src/Soulink-Web-Backend/scripts/deploy_questionnaire_migrate.sh)

这个脚本适用于以下场景：

- 服务器上只有 Docker 环境
- 服务器上没有项目源码
- 你希望在本地或 CI 上完成编译，再上传到服务器
- 你不想开放 PostgreSQL 端口到宿主机或公网

## 脚本会做什么

执行 `deploy_questionnaire_migrate.sh` 后，脚本会自动完成：

1. 在本地编译 `questionnairemigrate`
2. 创建服务器上的迁移目录
3. 上传迁移二进制、问卷文件和迁移配置
4. 在服务器上修复历史遗留的问卷答案唯一约束问题：
   - 删除旧约束 `user_questionnaire_answers_user_id_questionnaire_key_questi_key`
   - 删除旧约束 `user_questionnaire_answers_user_id_questionnaire_key_question_id_key`
   - 保留或补建新约束 `user_questionnaire_answers_user_key_version_question_key`
5. 在远程宿主机上通过 Docker CLI 获取 PostgreSQL 容器信息：
   - 使用 `docker inspect` 获取容器 IP
   - 使用 `docker exec printenv` 读取容器内数据库密码
6. 在远程宿主机上直接执行 `questionnairemigrate`，通过 Docker 内部网络访问数据库，不依赖数据库端口映射

也就是说，现在整套流程已经自动覆盖：

- Linux 二进制编译
- 文件上传
- 线上旧约束清理
- 通过 Docker 内网执行正式迁移

## 为什么不需要开放数据库端口

脚本不会从你的本地电脑直接连接服务器上的 PostgreSQL。

它的工作方式是：

1. 先 SSH 登录远程宿主机
2. 在远程宿主机上使用 Docker CLI 获取 `soulink-postgres` 容器的内网 IP
3. 在远程宿主机上读取容器内的 `POSTGRES_PASSWORD`
4. 在远程宿主机上运行 `questionnairemigrate`，让这个二进制直接连接容器 IP 的 `5432`

因此：

- 不需要把 PostgreSQL 映射到宿主机端口
- 不需要对公网开放数据库端口
- 只要求远程宿主机本身能访问 Docker 网络中的 PostgreSQL 容器

## 迁移脚本本身的行为

`questionnairemigrate` 在执行时会：

- 保留旧版本问卷数据，不会删除旧答案
- 清空该用户在目标版本下已有的答案、状态和画像，再重新写入
- 只迁移 `mappings` 中声明的题目
- 如果旧答案与新题型或新选项不兼容，该题会被跳过
- 按新问卷重新计算：
  - 已回答题数
  - 必答题已回答数
  - 进度百分比
  - `has_completed_questionnaire`
  - 是否继续参与匹配

因此，迁移完成后用户通常会看到：

- 重复题已经自动带出
- 新增题或未映射题仍为空
- 如果新版本必答题没有全部补齐，状态会变成未完成

## 迁移前准备

在运行脚本前，请先确认以下内容：

1. 你本地已经安装 Go，并且能执行 `go build`
2. 你已经准备好迁移配置文件 `configs/questionnaire/answer_migration.local.json`
3. 你本地的 `configs/questionnaire/` 目录下，已经放好了迁移所需的新旧问卷文件
4. 你知道目标服务器地址，并且服务器上能正常执行 `docker inspect`、`docker exec` 和 `psql`
5. PostgreSQL 容器内存在以下环境或等价默认值：
   - 用户名：默认 `solink`
   - 数据库名：默认 `soulink-web`
   - 密码环境变量名：默认 `POSTGRES_PASSWORD`

## 迁移配置文件

先复制模板：

```bash
cp configs/questionnaire/answer_migration.example.json configs/questionnaire/answer_migration.local.json
```

然后编辑 `configs/questionnaire/answer_migration.local.json`。

主要字段如下：

- `sourceCatalogFile`：旧问卷 Markdown 文件名
- `sourceQuestionnaireKey`：旧问卷标识，用于校验配置是否写错
- `sourceQuestionnaireVersion`：旧问卷版本号，用于校验配置是否写错
- `targetCatalogFile`：新问卷 Markdown 文件名
- `targetQuestionnaireKey`：新问卷标识，用于校验配置是否写错
- `targetQuestionnaireVersion`：新问卷版本号，用于校验配置是否写错
- `mappings`：题目映射关系，只会迁移这里列出的题目

示例：

```json
{
  "sourceCatalogFile": "thu_questionnaire_template_v0.1.md",
  "sourceQuestionnaireKey": "thu-date",
  "sourceQuestionnaireVersion": "v0.1",
  "targetCatalogFile": "thu_questionnaire_template.md",
  "targetQuestionnaireKey": "thu-date",
  "targetQuestionnaireVersion": "v0.2",
  "mappings": [
    {
      "sourceQuestionId": "1",
      "targetQuestionId": "1"
    },
    {
      "sourceQuestionId": "2",
      "targetQuestionId": "2"
    }
  ]
}
```

注意：

- 这里填写的是文件名，不是绝对路径
- 编译时会把 `configs/questionnaire/*.md` 打包进二进制
- 因此参与迁移的问卷文件，必须在编译前就已经存在于本地仓库的 `configs/questionnaire/` 目录下

## 一键上传并执行迁移

在项目根目录下执行：

```bash
cd /path/to/Soulink-Web-Backend
chmod +x ./scripts/deploy_questionnaire_migrate.sh
SERVER_HOST=203.0.113.10 \
SOURCE_CATALOG_FILE=thu_questionnaire_template_v0.5.md \
TARGET_CATALOG_FILE=thu_questionnaire_template.md \
./scripts/deploy_questionnaire_migrate.sh
```

这会使用脚本默认值：

- `SERVER_USER=root`
- `SERVER_PORT=22`
- `REMOTE_DIR=/opt/soulink-migrate`
- `GOARCH_TARGET=amd64`
- `MIGRATION_CONFIG_FILE=configs/questionnaire/answer_migration.local.json`
- `REMOTE_POSTGRES_CONTAINER=soulink-postgres`
- `POSTGRES_DB=soulink-web`
- `POSTGRES_USER=solink`
- `POSTGRES_PASSWORD_ENV=POSTGRES_PASSWORD`

如果你的环境不同，可以在命令前覆盖变量。

例如密码环境变量名不是 `POSTGRES_PASSWORD`：

dev 服务器
203.0.113.10
deploy服务器
203.0.113.10

```bash
SERVER_HOST=203.0.113.10 \
SOURCE_CATALOG_FILE=thu_questionnaire_template_v0.3.md \
TARGET_CATALOG_FILE=thu_questionnaire_template.md \
./scripts/deploy_questionnaire_migrate.sh
```

## 成功输出示例

脚本成功后，通常会在本地终端看到类似输出：

```text
source=thu-date@v0.1 target=thu-date@v0.2 users=1 migratedAnswers=48 skippedAnswers=0 completedUsers=0 incompleteUsers=1
```

字段含义如下：

- `users`：处理的用户数量
- `migratedAnswers`：成功迁移的答案数量
- `skippedAnswers`：因题型或选项不兼容而跳过的答案数量
- `completedUsers`：迁移后在新版本下已完成的用户数量
- `incompleteUsers`：迁移后在新版本下仍未完成的用户数量

## 推荐执行顺序

1. 准备好 `answer_migration.local.json`
2. 确认本地 `configs/questionnaire/` 中的新旧问卷文件齐全
3. 运行 `scripts/deploy_questionnaire_migrate.sh`
4. 查看脚本输出，确认迁移成功
5. 切换 API 的 `QUESTIONNAIRE_CATALOG_FILE`
6. 重启或重建 API 服务
7. 验证新问卷是否按预期展示

## 迁移检查

```bash
docker exec -i soulink-postgres psql -U solink -d soulink-web -P pager=off -x -c "SELECT u.id AS user_id, u.email, u.nickname, s.questionnaire_key, s.questionnaire_version, s.answered_count, s.required_answered_count, s.has_completed_questionnaire, s.participating_in_match, s.last_saved_at FROM users u JOIN user_questionnaire_states s ON s.user_id = u.id WHERE u.email = 'ztl22@mails.tsinghua.edu.cn' ORDER BY s.questionnaire_key, s.questionnaire_version;"
```

## 与后端切换版本的关系

迁移脚本只负责迁移数据库里的问卷答案数据，不会自动让前端开始读取新问卷。

如果你希望前端真正切到新版本，还需要把 API 的环境变量：

```env
QUESTIONNAIRE_CATALOG_FILE=旧问卷文件名
```

改成：

```env
QUESTIONNAIRE_CATALOG_FILE=新问卷文件名
```

然后重启或重新部署 API 服务。

也就是说：

- 只执行迁移脚本：数据库里已有新版本答案，但前端仍可能显示旧问卷
- 修改 `QUESTIONNAIRE_CATALOG_FILE` 并重启 API：前端才会开始读取新版本问卷

## 常见问题

### 1. 为什么现在不再需要上传 `.env`？

因为迁移脚本不再依赖服务器上的 `.env` 去读取 `DB_DSN`。它会在远程宿主机上临时拼出一个只用于本次迁移的 `DB_DSN`，然后直接执行迁移程序。

### 2. 为什么问卷文件必须在本地仓库里准备好？

因为问卷 Markdown 会在编译 `questionnairemigrate` 时通过 `embed` 打包进二进制。上传到服务器的 Markdown 文件主要用于留档和排查，不是迁移程序运行时的主要读取来源。

### 3. 为什么脚本要顺手修复数据库约束？

因为历史数据库里可能同时残留旧约束和新约束。旧约束会阻止同一用户在 `v0.1` 和 `v0.2` 下保存同一题号的答案，导致迁移时报重复键错误。

### 4. 迁移后，旧版本问卷数据还在吗？

在。迁移脚本不会删除旧版本答案，只会写入目标版本，并重算目标版本状态。

## 额外建议

- 迁移前先备份数据库
- 先在测试环境跑同样的迁移配置
- 正式执行时尽量选择低峰期
- 保留本次使用的 `answer_migration.local.json`，便于审计和重复执行