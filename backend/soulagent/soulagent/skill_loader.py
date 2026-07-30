from __future__ import annotations

import json
from pathlib import Path
from typing import Any


def skill_root() -> Path:
    return Path(__file__).resolve().parent.parent / "skills"


def load_skill(skill_id: str) -> dict[str, Any]:
    path = skill_root() / f"{skill_id}.json"
    if not path.exists():
        raise FileNotFoundError(f"Skill not found: {path}")
    return json.loads(path.read_text(encoding="utf-8"))
