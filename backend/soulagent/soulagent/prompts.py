from __future__ import annotations

import json
from typing import Any

from .schema import Traits


BASE_SYSTEM_PROMPT = """你是 SoulAgent。请把问卷 traits 写成细腻、自然、带一点文学气息的人物画像。
你需要输出三个部分：灵魂签名、个人画像、理想共振。
禁止套用 MBTI，也不要凭空编造不存在的人设，所有判断都必须回到输入 traits。
除非输入 traits 本身就是英文专有名词，否则所有输出必须使用自然、完整的中文，不要中英混杂。"""


def _top_items(values: list[str], limit: int = 3) -> list[str]:
    return values[:limit]


def _compress_traits(traits: Traits) -> dict[str, Any]:
    data = traits.to_dict()
    return {
        "relationship_goal": data["relationship_goal"],
        "stage_group": data["stage_group"],
        "location_stability": data["location_stability"],
        "social_energy": data["social_energy"],
        "attraction_tone": data["attraction_tone"],
        "expression_preference": data["expression_preference"],
        "life_pace": data["life_pace"],
        "meeting_frequency": data["meeting_frequency"],
        "autonomy_need": data["autonomy_need"],
        "reassurance_need": data["reassurance_need"],
        "conflict_style": data["conflict_style"],
        "intimacy_timing": data["intimacy_timing"],
        "career_drive": data["career_drive"],
        "future_lifestyle": data["future_lifestyle"],
        "attraction_keywords": _top_items(data["attraction_keywords"]),
        "self_keywords": _top_items(data["self_keywords"]),
        "spark_keywords": _top_items(data["spark_keywords"]),
        "core_interests": _top_items(data["core_interests"]),
        "security_sources": _top_items(data["security_sources"], 2),
        "dealbreakers": _top_items(data["dealbreakers"], 2),
    }


def build_system_prompt(skill: dict[str, Any]) -> str:
    campus_codes = "、".join(skill.get("campus_code_bank", [])[:6])
    signature_styles = "、".join(skill.get("signature_styles", [])[:4])
    return "\n".join(
        [
            BASE_SYSTEM_PROMPT,
            str(skill.get("system_identity", "")),
            "灵魂签名控制在 10 个汉字以内，最好 6-10 个字。",
            f"可以轻微借用校园意象，但重点仍然是 traits。本次可参考的校园语汇：{campus_codes}。",
            f"签名风格可参考这些方向，但不要复用固定短句：{signature_styles}。",
        ]
    )


def build_user_prompt(traits: Traits, skill: dict[str, Any]) -> str:
    output_format = skill.get("output_format", {})
    compressed_traits = _compress_traits(traits)

    return "\n".join(
        [
            "请严格按以下顺序输出三个标题：",
            output_format.get("soul_signature", "灵魂签名"),
            output_format.get("personal_profile", "个人画像"),
            output_format.get("ideal_resonance", "理想共振"),
            "",
            "写作要求：",
            "- 所有输出必须是中文；不要出现英文句子、英文形容词、英文小标题或中英混杂表达。",
            "- 灵魂签名只写 1 行短句，不要超过 10 个汉字。",
            "- 灵魂签名必须基于当前这份 traits 单独生成，不要复用示例或套用固定短句。",
            "- 个人画像要紧扣 traits，重点写相处方式、节奏偏好、边界感、情绪处理和吸引力来源。",
            "- 理想共振要紧扣 traits，重点写适合什么样的人、什么样的关系氛围。",
            "- 在个人画像和理想共振中，各自自然挑选 2-4 个短词或短语，用 **加粗标记** 包起来，方便前端高亮。",
            "- 加粗的词语也必须是中文，不要把英文词放进 ** ** 里。",
            "- 不要编造 traits 中没有出现的具体生活事实。",
            "- 可以有一点校园气息，但不要盖过人物本身。",
            "",
            "traits 摘要：",
            json.dumps(compressed_traits, ensure_ascii=False, indent=2),
        ]
    )
