# SoulAgent MVP

本地可跑的“画像整理师”最小版本。

它做四件事：

1. 把问卷答案映射成稳定的 traits
2. 通过 LLM 生成三段式报告
   - 灵魂签名
   - 个人画像
   - 理想同频
3. 支持调用主流 LLM API
   - DeepSeek
   - SiliconFlow
   - Paratera / 并行智算云
   - 智谱
   - OpenRouter
   - Groq
4. 支持从本地 `secrets.json` 或环境变量读取 API key
5. 支持通过 skill 文件控制 LLM 的报告风格

## 目录

- `soulagent/`
  - `schema.py`：traits schema 与问卷字段定义
  - `simulator.py`：10 份模拟答卷生成器
  - `traits_engine.py`：问卷答案 -> traits 的映射规则
  - `prompts.py`：LLM 系统提示词与 few-shot 示例
  - `llm_client.py`：DeepSeek / SiliconFlow / 智谱 / OpenRouter / Groq 调用
  - `pipeline.py`：主流程
  - `cli.py`：命令行入口
- `data/`
  - `simulated_responses.json`
- `outputs/`
  - 批量生成的 markdown 报告
- `thu_questionnaire.md`
  - 当前问卷草案
- `skills/`
  - LLM 风格技能文件

## 快速开始

### 1. 先生成 10 份模拟答卷

```bash
python -m soulagent.cli simulate
```

### 2. 用 LLM 模式生成报告

推荐方式是在 `soulagent` 目录下创建 `secrets.json`。示例：

```json
{
  "provider": "deepseek",
  "api_keys": {
    "deepseek": "你的key"
  },
  "models": {
    "deepseek": "deepseek-chat"
  }
}
```

然后直接运行：

```bash
python -m soulagent.cli batch --skill tsinghua_soulink
```

也支持环境变量方式。注意：`SOULAGENT_PROVIDER` 填 provider 名称，不要填 base URL。

DeepSeek 直连：

```bash
set SOULAGENT_PROVIDER=deepseek
set SOULAGENT_API_KEY=你的key
set SOULAGENT_MODEL=deepseek-chat
```

硅基流动：

```bash
set SOULAGENT_PROVIDER=siliconflow
set SOULAGENT_API_KEY=你的key
set SOULAGENT_MODEL=deepseek-ai/DeepSeek-V3
```

并行智算云：

```bash
set SOULAGENT_PROVIDER=paratera
set SOULAGENT_API_KEY=你的key
set SOULAGENT_API_BASE=控制台提供的OpenAI兼容base_url
set SOULAGENT_MODEL=DeepSeek-V3
```

智谱：

```bash
set SOULAGENT_PROVIDER=zhipu
set SOULAGENT_API_KEY=你的key
set SOULAGENT_MODEL=glm-4.7-flash
```

OpenRouter：

```bash
set SOULAGENT_PROVIDER=openrouter
set SOULAGENT_API_KEY=你的key
set SOULAGENT_MODEL=openrouter/auto
```

Groq：

```bash
set SOULAGENT_PROVIDER=groq
set SOULAGENT_API_KEY=你的key
set SOULAGENT_MODEL=llama-3.1-8b-instant
```

如果你有多个 key，希望轮询容错，也可以改为：

```bash
set SOULAGENT_PROVIDER=deepseek
set SOULAGENT_API_KEYS=key-1,key-2,key-3
set SOULAGENT_MODEL=deepseek-chat
```

如果你想按 provider + apiKey + model 三元组轮询，也可以改为：

```bash
set "SOULAGENT_MODEL_POOL=[{"provider":"deepseek","apiKey":"key-1","model":"deepseek-chat"},{"provider":"siliconflow","apiKey":"key-2","model":"deepseek-ai/DeepSeek-V3"}]"
```

配置 `SOULAGENT_MODEL_POOL` 后，会优先于单 provider 配置生效；每次调用都会取当前元素并把指针加一。

然后运行：

```bash
python -m soulagent.cli batch
```

## 推荐模型

- DeepSeek 直连：`deepseek-chat`
- SiliconFlow：`deepseek-ai/DeepSeek-V3`
- 智谱：`glm-4.7-flash`

## Skill

当前已内置一个 skill：

- `tsinghua_soulink`

它定义了：

- 画像整理师的角色身份
- 清华 Soulink 的语气和风格规则
- 灵魂签名的命名原则
- 输出结构
- few-shot 示例

你可以这样指定：

```bash
python -m soulagent.cli single sim-01 --skill tsinghua_soulink
```

## 输出原则

这个 MVP 不做：

- 心理诊断
- 创伤推断
- 宿命论匹配
- 超出问卷证据的夸张解读

它只做：

- 信息压缩
- 风格化表达
- 可解释的偏好总结
- 对理想同频对象的边界描述
