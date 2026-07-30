from __future__ import annotations

import http.client
import json
import os
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path
from typing import Callable


class LLMError(RuntimeError):
    pass


_PROVIDER_CONFIGS = {
    "deepseek": {
        "url": "https://api.deepseek.com/chat/completions",
        "default_model": "deepseek-chat",
        "headers": {},
    },
    "groq": {
        "url": "https://api.groq.com/openai/v1/chat/completions",
        "default_model": "llama-3.1-8b-instant",
        "headers": {},
    },
    "openrouter": {
        "url": "https://openrouter.ai/api/v1/chat/completions",
        "default_model": "openrouter/auto",
        "headers": {
            "HTTP-Referer": "https://local.soulagent",
            "X-Title": "SoulAgent MVP",
        },
    },
    "siliconflow": {
        "url": "https://api.siliconflow.cn/v1/chat/completions",
        "default_model": "deepseek-ai/DeepSeek-V3",
        "headers": {},
    },
    "paratera": {
        "url": "",
        "default_model": "DeepSeek-V3",
        "headers": {},
    },
    "zhipu": {
        "url": "https://open.bigmodel.cn/api/paas/v4/chat/completions",
        "default_model": "glm-4.7-flash",
        "headers": {},
    },
}

_PROVIDER_ALIASES = {
    "https://api.deepseek.com": "deepseek",
    "https://api.deepseek.com/v1": "deepseek",
    "https://api.deepseek.com/chat/completions": "deepseek",
    "https://api.siliconflow.cn": "siliconflow",
    "https://api.siliconflow.cn/v1": "siliconflow",
    "https://api.siliconflow.cn/v1/chat/completions": "siliconflow",
}

_MODEL_POOL_CURSOR_KEY = "__model_pool__"


def _project_root() -> Path:
    return Path(__file__).resolve().parent.parent


def _load_file_config() -> dict:
    config_path = _project_root() / "secrets.json"
    if not config_path.exists():
        return {}
    try:
        return json.loads(config_path.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        raise LLMError(f"Invalid JSON in {config_path}") from exc


def _normalize_provider_name(provider: str) -> str:
    normalized = str(provider).strip().lower().rstrip("/")
    if not normalized:
        return ""
    return _PROVIDER_ALIASES.get(normalized, normalized)


def _resolve_provider() -> str:
    file_config = _load_file_config()
    raw_provider = os.getenv("SOULAGENT_PROVIDER", "").strip() or str(file_config.get("provider", "")).strip()
    return _normalize_provider_name(raw_provider)


def _resolve_model(provider: str) -> str:
    file_config = _load_file_config()
    models = file_config.get("models", {})
    env_model = os.getenv("SOULAGENT_MODEL", "").strip()
    if env_model:
        return env_model
    if isinstance(models, dict):
        file_model = models.get(provider, "")
        if isinstance(file_model, str) and file_model.strip():
            return file_model.strip()
    return ""


def _resolve_provider_url(provider: str) -> str:
    config = _PROVIDER_CONFIGS.get(provider)
    if config is None:
        return ""

    if provider == "paratera":
        custom_base = os.getenv("SOULAGENT_API_BASE", "").strip().rstrip("/")
        if custom_base:
            if custom_base.endswith("/chat/completions"):
                return custom_base
            if custom_base.endswith("/v1"):
                return f"{custom_base}/chat/completions"
            return f"{custom_base}/v1/chat/completions"

    return str(config.get("url", "")).strip()


def _normalize_api_keys(raw: object) -> list[str]:
    if isinstance(raw, str):
        return [item.strip() for item in raw.split(",") if item.strip()]
    if isinstance(raw, list):
        return [str(item).strip() for item in raw if str(item).strip()]
    return []


def _resolve_api_keys(provider: str) -> list[str]:
    env_keys = _normalize_api_keys(os.getenv("SOULAGENT_API_KEYS", ""))
    if env_keys:
        return env_keys

    env_key = os.getenv("SOULAGENT_API_KEY", "").strip()
    if env_key:
        return [env_key]

    file_config = _load_file_config()
    api_keys = file_config.get("api_keys", {})
    if isinstance(api_keys, dict):
        return _normalize_api_keys(api_keys.get(provider, []))
    return []


def _normalize_model_pool_entries(raw: object) -> list[dict[str, str]]:
    if raw in (None, ""):
        return []

    if isinstance(raw, str):
        content = raw.strip()
        if not content:
            return []
        try:
            data = json.loads(content)
        except json.JSONDecodeError as exc:
            raise LLMError("Invalid JSON in SOULAGENT_MODEL_POOL") from exc
    else:
        data = raw

    if not isinstance(data, list):
        raise LLMError("SOULAGENT_MODEL_POOL must be a JSON array")

    entries: list[dict[str, str]] = []
    supported = ", ".join(sorted(_PROVIDER_CONFIGS))
    for index, item in enumerate(data, start=1):
        if not isinstance(item, dict):
            raise LLMError(f"SOULAGENT_MODEL_POOL item #{index} must be an object")

        provider = _normalize_provider_name(str(item.get("provider", "")))
        api_key = item.get("apiKey")
        if api_key in (None, ""):
            api_key = item.get("api_key")
        model = str(item.get("model", "")).strip()
        api_key_text = str(api_key).strip() if api_key is not None else ""

        if not provider or not api_key_text or not model:
            raise LLMError(
                f"SOULAGENT_MODEL_POOL item #{index} must include provider, apiKey, and model"
            )
        if provider not in _PROVIDER_CONFIGS:
            raise LLMError(
                f"Unsupported provider in SOULAGENT_MODEL_POOL item #{index}: {provider}. "
                f"Supported providers: {supported}"
            )

        entries.append(
            {
                "provider": provider,
                "api_key": api_key_text,
                "model": model,
            }
        )

    return entries


def _resolve_model_pool() -> list[dict[str, str]]:
    env_pool = os.getenv("SOULAGENT_MODEL_POOL", "").strip()
    if env_pool:
        return _normalize_model_pool_entries(env_pool)

    file_config = _load_file_config()
    return _normalize_model_pool_entries(file_config.get("model_pool", []))


def _cursor_path() -> Path:
    return _project_root() / ".api_key_cursor.json"


def _load_cursor() -> dict[str, int]:
    path = _cursor_path()
    if not path.exists():
        return {}
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        return {}
    if not isinstance(data, dict):
        return {}
    result: dict[str, int] = {}
    for key, value in data.items():
        if isinstance(value, int):
            result[str(key)] = value
    return result


def _save_cursor(cursor: dict[str, int]) -> None:
    _cursor_path().write_text(json.dumps(cursor, ensure_ascii=False), encoding="utf-8")


def _next_model_pool_entry(entries: list[dict[str, str]]) -> dict[str, str]:
    if not entries:
        raise LLMError("SOULAGENT_MODEL_POOL is empty")

    cursor = _load_cursor()
    start = cursor.get(_MODEL_POOL_CURSOR_KEY, 0) % len(entries)
    cursor[_MODEL_POOL_CURSOR_KEY] = (start + 1) % len(entries)
    _save_cursor(cursor)
    return entries[start]


def _rotate_api_keys(provider: str, keys: list[str]) -> list[str]:
    if len(keys) <= 1:
        return keys

    cursor = _load_cursor()
    start = cursor.get(provider, 0) % len(keys)
    cursor[provider] = (start + 1) % len(keys)
    _save_cursor(cursor)

    return keys[start:] + keys[:start]


def _post_json(url: str, payload: dict, headers: dict[str, str]) -> dict:
    last_error: Exception | None = None
    for attempt in range(1, 4):
        request = urllib.request.Request(
            url,
            data=json.dumps(payload).encode("utf-8"),
            headers={
                **headers,
                "Content-Type": "application/json",
                "Accept": "application/json",
                "User-Agent": "SoulAgent/0.1",
            },
            method="POST",
        )
        try:
            with urllib.request.urlopen(request, timeout=60) as response:
                return json.loads(response.read().decode("utf-8"))
        except urllib.error.HTTPError as exc:
            detail = exc.read().decode("utf-8", errors="ignore")
            raise LLMError(f"HTTP {exc.code}: {detail}") from exc
        except (urllib.error.URLError, http.client.RemoteDisconnected, TimeoutError) as exc:
            last_error = exc
            if attempt < 3:
                time.sleep(attempt)
                continue
            break

    raise LLMError(f"Network error after retries: {last_error}") from last_error


def _extract_message_content(data: dict) -> str:
    try:
        message = data["choices"][0]["message"]
    except (KeyError, IndexError, TypeError) as exc:
        raise LLMError(f"Malformed response payload: {data}") from exc

    content = message.get("content")
    if isinstance(content, str) and content.strip():
        return content
    if isinstance(content, list):
        parts: list[str] = []
        for item in content:
            if not isinstance(item, dict):
                continue
            text = item.get("text")
            if isinstance(text, str) and text:
                parts.append(text)
        if parts:
            return "".join(parts)

    reasoning = message.get("reasoning_content")
    if isinstance(reasoning, str) and reasoning.strip():
        return reasoning

    raise LLMError("Response did not include message content")


def generate_with_provider(provider: str, system_prompt: str, user_prompt: str, model: str, api_key: str) -> str:
    config = _PROVIDER_CONFIGS[provider]
    url = _resolve_provider_url(provider)
    if not url:
        if provider == "paratera":
            raise LLMError("Provider paratera requires SOULAGENT_API_BASE to be set to an OpenAI-compatible endpoint")
        raise LLMError(f"Provider {provider} has no configured API URL")
    print(f"[soulagent] provider={provider} model={model} url={url}", file=sys.stderr)
    data = _post_json(
        url,
        {
            "model": model,
            "messages": [
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_prompt},
            ],
        },
        {
            "Authorization": f"Bearer {api_key}",
            **config.get("headers", {}),
        },
    )
    return _extract_message_content(data)


def _generate_with_failover(
    provider: str,
    api_keys: list[str],
    runner: Callable[[str], str],
) -> str:
    ordered_keys = _rotate_api_keys(provider, api_keys)
    errors: list[str] = []

    for index, api_key in enumerate(ordered_keys, start=1):
        try:
            return runner(api_key)
        except LLMError as exc:
            errors.append(f"key#{index}: {exc}")
            continue

    raise LLMError("; ".join(errors) if errors else "All API keys failed")


def generate_text(system_prompt: str, user_prompt: str) -> str:
    expected_path = _project_root() / "secrets.json"
    model_pool = _resolve_model_pool()
    if model_pool:
        selected = _next_model_pool_entry(model_pool)
        return generate_with_provider(
            selected["provider"],
            system_prompt,
            user_prompt,
            selected["model"],
            selected["api_key"],
        )

    provider = _resolve_provider()
    api_keys = _resolve_api_keys(provider)

    if not provider or not api_keys:
        raise LLMError(
            f"Missing provider or api key. Set environment variables or create {expected_path}"
        )

    config = _PROVIDER_CONFIGS.get(provider)
    if config is None:
        supported = ", ".join(sorted(_PROVIDER_CONFIGS))
        raise LLMError(f"Unsupported provider: {provider}. Supported providers: {supported}")

    model = _resolve_model(provider) or str(config["default_model"])
    return _generate_with_failover(
        provider,
        api_keys,
        lambda api_key: generate_with_provider(
            provider,
            system_prompt,
            user_prompt,
            model,
            api_key,
        ),
    )
