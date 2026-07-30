from __future__ import annotations

import json
from pathlib import Path
from typing import Any, Callable

from .llm_client import generate_text
from .prompts import build_system_prompt, build_user_prompt
from .skill_loader import load_skill
from .simulator import save_simulated_responses
from .traits_engine import derive_traits


SECTION_TITLES = {
    "soul_signature": "灵魂签名",
    "personal_profile": "个人画像",
    "ideal_resonance": "理想共振",
}


def load_responses(path: Path) -> list[dict[str, Any]]:
    return json.loads(path.read_text(encoding="utf-8"))


def save_markdown_report(
    output_path: Path,
    response: dict[str, Any],
    traits: dict[str, Any],
    report: dict[str, str],
    mode: str,
) -> None:
    output_path.parent.mkdir(parents=True, exist_ok=True)
    content = [
        f"# SoulAgent Report: {response['meta']['respondent_id']}",
        "",
        f"- mode: `{mode}`",
        f"- archetype: `{response['meta'].get('archetype', 'unknown')}`",
        "",
        f"## {SECTION_TITLES['soul_signature']}",
        "",
        report["soul_signature"].strip(),
        "",
        f"## {SECTION_TITLES['personal_profile']}",
        "",
        report["personal_profile"].strip(),
        "",
        f"## {SECTION_TITLES['ideal_resonance']}",
        "",
        report["ideal_resonance"].strip(),
        "",
        "## Traits",
        "",
        "```json",
        json.dumps(traits, ensure_ascii=False, indent=2),
        "```",
    ]
    output_path.write_text("\n".join(content), encoding="utf-8")


def parse_llm_report(text: str) -> dict[str, str]:
    sections = {key: "" for key in SECTION_TITLES}
    current: str | None = None

    for line in text.splitlines():
        stripped = line.strip()
        if not stripped:
            continue

        matched = False
        for key, title in SECTION_TITLES.items():
            if stripped == title or stripped.startswith(f"{title}:") or stripped.startswith(f"{title}："):
                parts = stripped.replace("：", ":", 1).split(":", 1)
                sections[key] = parts[1].strip() if len(parts) > 1 else ""
                current = key
                matched = True
                break

        if matched:
            continue

        if current:
            sections[current] = (sections[current] + "\n" + stripped).strip()

    return sections


def build_report(
    response: dict[str, Any],
    skill_id: str = "tsinghua_soulink",
) -> tuple[dict[str, Any], dict[str, str]]:
    traits = derive_traits(response)
    skill = load_skill(skill_id)
    system_prompt = build_system_prompt(skill)
    user_prompt = build_user_prompt(traits, skill)
    raw = generate_text(system_prompt, user_prompt)
    parsed = parse_llm_report(raw)
    return traits.to_dict(), parsed


def ensure_simulated_data(data_path: Path, count: int = 10) -> None:
    if not data_path.exists():
        save_simulated_responses(data_path, count=count)


def run_batch(
    project_root: Path,
    limit: int | None = None,
    progress_callback: Callable[[int, int, str], None] | None = None,
    skill_id: str = "tsinghua_soulink",
) -> list[Path]:
    data_path = project_root / "data" / "simulated_responses.json"
    output_dir = project_root / "outputs" / "llm"
    ensure_simulated_data(data_path)
    responses = load_responses(data_path)
    if limit is not None:
        responses = responses[:limit]

    written: list[Path] = []
    total = len(responses)
    for index, response in enumerate(responses, start=1):
        if progress_callback:
            progress_callback(index, total, response["meta"]["respondent_id"])
        traits, report = build_report(response, skill_id=skill_id)
        out = output_dir / f"{response['meta']['respondent_id']}.md"
        save_markdown_report(out, response, traits, report, mode="llm")
        written.append(out)
    return written


def run_single(
    project_root: Path,
    respondent_id: str,
    skill_id: str = "tsinghua_soulink",
) -> Path:
    data_path = project_root / "data" / "simulated_responses.json"
    ensure_simulated_data(data_path)
    responses = load_responses(data_path)
    target = next(item for item in responses if item["meta"]["respondent_id"] == respondent_id)
    traits, report = build_report(target, skill_id=skill_id)
    out = project_root / "outputs" / "llm" / f"{respondent_id}.md"
    save_markdown_report(out, target, traits, report, mode="llm")
    return out
