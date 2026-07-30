# Soulink Matching System

基于问卷的高质量 1:1 匹配系统。通过七大维度评分、质量门控和稳定匹配算法，为用户找到最合适的匹配对象。

## 快速开始

### 1. 准备环境

```bash
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### 2. 准备数据

将问卷导出的 CSV 文件放入 `data/` 目录：

```
data/
└── export.csv        ← 问卷导出数据（必需）
```

`data/` 目录已被 `.gitignore` 排除，不会进入版本控制。

**CSV 格式要求**：

- 编码：UTF-8（支持 BOM）
- 必需列：`user_id`, `email`, `nickname`, `department`, `grade`, `completed`, `participating`
- 问卷列：`{题号}_value`, `{题号}_importance`（题号 1-52）
- `completed` 和 `participating` 均为 `true` 的用户才会参与匹配
- 多选题值用英文分号 `;` 分隔（如 `真诚;幽默;温柔`）
- 范围题值为 JSON（如 `{"max": 25, "min": 18}` 或 `{"unlimited": true}`）

### 3. 运行匹配

```bash
python run_matching.py
```

可选参数：

```bash
python run_matching.py --csv data/export.csv --min-score 0.40 --output data/match_results.csv
```

| 参数 | 默认值 | 说明 |
|---|---|---|
| `--csv` | `data/export.csv` | 输入 CSV 路径 |
| `--min-score` | `0.40` | 最低可接受分数阈值（0-1） |
| `--output` | `data/match_results.csv` | 输出结果 CSV 路径 |

### 4. 查看结果

匹配结果输出到 `data/match_results.csv`，包含以下字段：

| 字段 | 说明 |
|---|---|
| `pool` | 匹配池（MF / MM / FF） |
| `user_a_id` / `user_b_id` | 配对双方 ID |
| `user_a_nickname` / `user_b_nickname` | 配对双方昵称 |
| `user_a_email` / `user_b_email` | 配对双方邮箱 |
| `score_a_to_b` / `score_b_to_a` | 双方偏好得分 |
| `match_quality` | 匹配质量分 = min(A→B, B→A) |
| `grade` | 质量等级（S ≥ 0.70 / A ≥ 0.55 / B ≥ 0.40） |
| `asymmetry_warning` | 若双方得分差 > 0.20，标记 `yes` |

终端同时会打印详细报告，包括每对配对的分数、未匹配者名单及原因。

### 5. 生成匹配解释

先确保已经生成：

- `data/export.csv`
- `data/match_results.csv`

启发式模式：

```bash
python generate_match_explanations.py --mode heuristic --output data/match_explanations.csv
```

LLM 模式：

```bash
python generate_match_explanations.py --mode llm --output data/match_explanations.csv
```

如需保留 `build_prompt` 对应的 JSON 结构，建议直接输出：

```bash
python generate_match_explanations.py --mode llm --output data/match_explanations.jsonl
```

常用参数：

| 参数 | 默认值 | 说明 |
|---|---|---|
| `--matches` | `data/match_results.csv` | 匹配结果 CSV |
| `--profiles` | `data/export.csv` | 问卷导出 CSV |
| `--output` | `data/match_explanations.csv` | 解释输出文件，支持 `.csv` / `.json` / `.jsonl` / `.xlsx` |
| `--mode` | `llm` | `llm` 或 `heuristic` |
| `--model` | 环境变量或默认模型 | LLM 模型名 |
| `--limit` | 空 | 只处理前 N 条匹配 |

当前解释输出：

- `match_group_no`
- `group_name`
- `explanation_text`
- `explanation_json`（仅 `.csv` / `.xlsx`，保存原始 JSON 字符串）

`.json` / `.jsonl` 会直接保留嵌套的 `explanation` 对象。

推荐执行顺序：

```bash
python run_matching.py --csv data/export.csv --output data/match_results.csv
python generate_match_explanations.py --mode llm --matches data/match_results.csv --profiles data/export.csv --output data/match_explanations.csv
```

## 系统设计

### 匹配流程

```
数据加载 → 池子分配 → 硬性筛选 → 七维评分 → 婚育观软否决
    → 质量截断 → 匹配算法 → 结果分级输出
```

### 三个匹配池

| 池子 | 分配条件 | 匹配算法 |
|---|---|---|
| 男女 (MF) | 男→女 且 女→男 | Gale-Shapley（双向运行取更优） |
| 男男 (MM) | 男→男 | Irving 稳定室友 → MWM 回退 |
| 女女 (FF) | 女→女 | Irving 稳定室友 → MWM 回退 |

### 七大评分维度

| 维度 | 权重 | 涵盖题目 |
|---|---|---|
| D1 目的与阶段 | 8% | Q3, Q4-5, Q6 |
| D2 地理与距离 | 7% | Q11, Q12 |
| D3 吸引与气质 | 15% | Q13-20 |
| D4 生活方式 | 20% | Q21-31 |
| D5 兴趣爱好 | 10% | Q32, Q33 |
| D6 依恋与亲密 | 20% | Q34-43 |
| D7 未来愿景 | 20% | Q44-52 |

### 质量控制

- **硬性筛选**：年龄 / 身高范围互通、极端吸烟不兼容 → 直接排除
- **婚育观软否决**：Q47（婚姻）或 Q48（生育）分歧 ≥ 5 → 总分 ×0.5
- **质量截断**：低于 `MIN_ACCEPTABLE_SCORE` 的候选人不进入偏好列表
- **宁缺毋滥**：不强制匹配，匹配不到高质量对象则留空

## 调参指南

编辑 `matching/config.py` 可调整：

- `MIN_ACCEPTABLE_SCORE`：质量门控阈值（↑ 更严格，↓ 更宽松）
- `IMPORTANCE_MULTIPLIER`：用户标记"对我更重要"时的权重放大系数
- `DIMENSION_WEIGHTS`：七个维度的权重分配
- 各兼容矩阵（校区距离、目的匹配、关系模式等）

详细设计文档见 [docs/matching_score_design.md](docs/matching_score_design.md)。

## 项目结构

```
Soulink-Web-Matching/
├── data/                       ← 数据目录（不入 git）
│   ├── export.csv              ← 输入：问卷导出数据
│   └── match_results.csv       ← 输出：匹配结果
├── matching/                   ← 核心模块
│   ├── config.py               ← 配置常量与兼容矩阵
│   ├── models.py               ← 数据模型
│   ├── parser.py               ← CSV 解析
│   ├── pools.py                ← 池子分配
│   ├── filters.py              ← 硬性筛选
│   ├── scoring.py              ← 七维评分系统
│   ├── algorithms.py           ← 匹配算法（GS + Irving + MWM）
│   └── report.py               ← 报告与结果导出
├── docs/
│   └── matching_score_design.md ← 评分系统设计文档
├── run_matching.py             ← 入口
├── requirements.txt
└── .gitignore
```
