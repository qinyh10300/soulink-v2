# 匹配结果导入说明

## 匹配结果 CSV 导入脚本

当你已经在线下匹配仓库生成了 `match_results.csv`，希望把结果回写到后端数据库中的 `match_cycles` 和 `user_match_results` 时，可以使用：

```bash
go run ./cmd/matchresultsimport -config ./configs/match/match_results_import.local.json -csv ../match_results.csv
```

这个命令会：

- 读取 `match_results.csv` 中的配对双方 ID 和分数
- 为每个用户生成一条 `matched` 结果
- 自动根据数据库中的用户资料补齐：
  - 对方昵称
  - 对方院系 / 年级
  - 对方最后保存的联系方式；如果未设置，则回退邮箱
- 基于 `match_quality`、`score_a_to_b`、`score_b_to_a`、`grade` 和 `asymmetry_warning` 生成 `match_reason` 作为相处提示
- 当 `markUnmatchedUsers=true` 时，自动把同一问卷版本下未出现在 CSV 里的参与者写成 `failed`

### 配置文件

示例配置文件：

- `configs/match/match_results_import.example.json`

建议先复制：

```bash
cp ./configs/match/match_results_import.example.json ./configs/match/match_results_import.local.json
```

主要配置项说明：

- `questionnaireKey`
  本轮结果对应的问卷标识

- `questionnaireVersion`
  本轮结果对应的问卷版本

- `cycleKey`
  本轮匹配周期唯一标识；如果重复使用同一个 `cycleKey`，脚本会覆盖该周期下已不存在于本次结果中的旧记录

- `title`
  本轮匹配周期标题

- `revealStatus`
  周期公布状态，通常为 `revealed`

- `revealAt`
  本轮结果的公布时间

- `dryRun`
  `true` 时只校验 CSV 和参与者，不落库

- `matchedMessage`
  匹配成功时写入的结果消息

- `unmatchedMessage`
  未匹配成功时写入的结果消息

- `reasonTemplate`
  相处提示前缀，会和分数说明及建议文案拼接成最终 `match_reason`

- `markUnmatchedUsers`
  是否自动把未出现在 CSV 中的参与者写成 `failed`

### 远程服务器导入

如果你希望像现有问卷导出 / 迁移脚本一样，通过 SSH 登录服务器并连接 Docker 中的 PostgreSQL，可以使用：

```bash
SERVER_HOST=203.0.113.10 \
RESULTS_CSV_FILE=results/match_results_r4.csv \
EXPLANATIONS_CSV_FILE=results/match_explanations_r4c.csv \
IMPORT_CONFIG_FILE=configs/match/match_results_import.local.json \
bash ./scripts/deploy_match_results_import.sh
```

这个部署脚本会：

1. 本地编译 `matchresultsimport`
2. 上传二进制、导入配置、`match_results.csv` 和揭晓时间配置
3. 在远程宿主机上通过 `docker inspect` 获取 `soulink-postgres` 容器 IP
4. 通过 `docker exec` 读取容器内的 `POSTGRES_PASSWORD`
5. 用宿主机上的导入命令直接连接容器数据库并完成写入

## 批量自动加好友脚本（基于匹配 CSV）

如果你希望根据匹配 CSV 将每对匹配对象直接设为好友（`accepted`），可以使用：

```bash
SERVER_HOST=203.0.113.10 \
RESULTS_CSV_FILE=match_results_r2.csv \
bash ./scripts/deploy_match_friendships_import.sh
```

可选参数：

- `DRY_RUN=true`：只校验并输出统计，不写库
- `ENSURE_CONVERSATIONS=true`：写好友关系时同步创建会话

该脚本会：

1. 读取 `user_a_id` / `user_b_id` 形成唯一配对
2. 跳过无效用户和重复配对
3. 将配对双方写成好友（已有 `accepted` 会跳过）
4. 跳过 `blocked` 关系并在统计里输出数量
