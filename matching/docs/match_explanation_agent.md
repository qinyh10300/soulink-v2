# Match Explanation Agent

This repo now includes `generate_match_explanations.py` for turning matched pairs into an explanation file.

## What It Does

It reads:

- `data/match_results.csv`
- `data/export.csv`

For each pair, it:

1. finds both users' questionnaire records by `user_id`
2. computes evidence from the existing scoring logic
3. sends the pair summary to an LLM Agent, or uses a local heuristic mode
4. exports a new file such as `data/match_explanations.csv`, `data/match_explanations.jsonl`, or `data/match_explanations.xlsx`

## Output Columns

The generated output includes:

- original pair fields from `match_results.csv`
- `pair_key` in the form `min_user_id:max_user_id`
- `source_match_user_a_id` / `source_match_user_b_id`
- both users' department and grade
- strongest and weakest scoring dimensions
- shared interests / shared activities / mutual trait hits
- `match_summary`
- `match_reasons`
- `date_topics`
- `cautions`
- generation status and mode

## Install

```bash
pip install -r requirements.txt
```

## Quick Start

Heuristic mode, no API required:

```bash
python generate_match_explanations.py --mode heuristic --limit 5
```

LLM mode with environment variables:

```bash
set OPENAI_API_KEY=your_api_key
set OPENAI_API_BASE=https://api.openai.com/v1
set OPENAI_MODEL=gpt-4.1-mini
set MATCH_EXPLANATION_MODE=llm
python generate_match_explanations.py
```

If you use another OpenAI-compatible provider, only replace `OPENAI_API_BASE` and `OPENAI_MODEL`.

SiliconFlow example:

```bash
set SOULAGENT_PROVIDER=siliconflow
set SOULAGENT_API_KEY=your_siliconflow_api_key
set SOULAGENT_MODEL=Qwen/Qwen2.5-7B-Instruct
set MATCH_EXPLANATION_MODE=llm
python generate_match_explanations.py
```

The script also loads variables from a local `.env` file automatically.

You can still override them on the command line:

```bash
python generate_match_explanations.py ^
  --api-key your_api_key ^
  --api-base https://api.openai.com/v1 ^
  --model gpt-4.1-mini
```

## Useful Options

```bash
python generate_match_explanations.py --help
```

Important flags:

- `--mode llm|heuristic`
- `--matches PATH`
- `--profiles PATH`
- `--output PATH`
- `--api-key-env OPENAI_API_KEY`
- `--api-base URL`
- `--model MODEL_NAME`
- `--limit N`
- `--sleep-seconds FLOAT`
- `--include-raw-json`

Important environment variables:

- `OPENAI_API_KEY`
- `OPENAI_API_BASE`
- `OPENAI_MODEL`
- `SOULAGENT_PROVIDER`
- `SOULAGENT_API_KEY`
- `SOULAGENT_MODEL`
- `MATCH_EXPLANATION_MODE`

## Suggested Workflow

1. Run `--mode heuristic --limit 3` first to verify the output layout.
2. After you provide the API key, switch to `--mode llm`.
3. Inspect a few rows manually and tighten the prompt if you want a different tone.
