from __future__ import annotations

import json
import random
from pathlib import Path
from typing import Any

from .schema import QUESTIONNAIRE_VERSION


TRAIT_POOL = [
    "真诚",
    "聪明",
    "幽默",
    "温柔",
    "自律",
    "稳定",
    "有野心",
    "有审美",
    "有担当",
    "好奇心强",
    "善于倾听",
    "有自己的热爱",
]

SPARK_POOL = [
    "被认真倾听",
    "被坚定选择",
    "被温柔照顾",
    "被欣赏和夸奖",
    "一起做成一件事",
    "一起安静待着也不尴尬",
]

INTEREST_POOL = [
    "阅读",
    "电影 / 剧集",
    "音乐 / 演出",
    "健身 / 跑步",
    "球类运动",
    "户外 / 徒步",
    "游戏",
    "摄影 / 视频",
    "展览 / 博物馆",
    "做饭 / 探店",
    "编程 / 科技",
    "学术讨论",
    "创作 / 写作",
    "旅行",
]

SECURITY_POOL = [
    "稳定回复和出现",
    "明确表达喜欢和在意",
    "主动规划见面",
    "冲突时不失联",
    "愿意介绍自己的朋友和生活",
    "尊重我的边界",
]

AFFECTION_POOL = [
    "主动找话题聊天",
    "记住细节并回应",
    "安排见面",
    "直接表达欣赏",
    "默默付出和帮忙",
    "分享我的真实生活",
]

DEALBREAKER_POOL = [
    "不真诚",
    "边界感差",
    "情绪不稳定且拒绝沟通",
    "不忠诚",
    "对未来完全没有规划",
    "价值观冲突过大",
    "不尊重他人",
]

DIETARY_POOL = [
    "无",
    "清真",
    "素食",
    "不吃辣",
    "对酒精敏感",
    "其他",
]

IDENTITY_POOL = ["本科生", "硕士生", "博士生", "已毕业校友", "教职工/研究人员"]
CAMPUS_POOL = ["校本部", "五道口金融学院", "深圳国际研究生院", "其他"]
TARGET_GENDER_BY_SELF = {
    "男": ["女", "男", "非二元性别"],
    "女": ["男", "女", "非二元性别"],
    "非二元性别": ["男", "女", "非二元性别"],
}


def sample_scale(center: int, spread: int = 2) -> int:
    return max(1, min(7, center + random.randint(-spread, spread)))


def choose_many(pool: list[str], low: int, high: int) -> list[str]:
    size = random.randint(low, high)
    return random.sample(pool, size)


def choose_age(identity: str, stage: str) -> int:
    if identity == "本科生":
        return random.randint(18, 23)
    if identity == "硕士生":
        return random.randint(22, 27)
    if identity == "博士生":
        return random.randint(24, 32)
    if identity == "已毕业校友":
        return random.randint(23, 34)
    if stage == "已毕业/工作":
        return random.randint(23, 34)
    return random.randint(22, 35)


def choose_height(self_gender: str) -> int:
    if self_gender == "男":
        return random.randint(165, 192)
    if self_gender == "女":
        return random.randint(152, 178)
    return random.randint(155, 185)


def choose_dietary() -> list[str]:
    picked = choose_many(DIETARY_POOL, 1, 2)
    if "无" in picked:
        return ["无"]
    return picked


def make_response(seed: int) -> dict[str, Any]:
    random.seed(seed)

    archetypes = [
        "slow_stable",
        "warm_listener",
        "urban_ambitious",
        "cool_boundaries",
        "gentle_romantic",
        "academic_partner",
        "independent_explorer",
        "high_aesthetic",
        "secure_builder",
        "thoughtful_late_bloomer",
    ]
    archetype = archetypes[seed % len(archetypes)]

    self_gender = random.choices(["男", "女", "非二元性别"], weights=[45, 45, 10])[0]
    target_gender = random.choice(TARGET_GENDER_BY_SELF[self_gender])
    identity = random.choices(IDENTITY_POOL, weights=[35, 35, 15, 10, 5])[0]
    stage = random.choice(["本科低年级（大一/大二）", "本科高年级（大三/大四）", "硕士阶段", "博士阶段", "已毕业/工作"])
    if identity == "本科生":
        stage = random.choice(["本科低年级（大一/大二）", "本科高年级（大三/大四）"])
    elif identity == "硕士生":
        stage = "硕士阶段"
    elif identity == "博士生":
        stage = "博士阶段"
    elif identity in ["已毕业校友", "教职工/研究人员"]:
        stage = "已毕业/工作"

    age = choose_age(identity, stage)
    age_range = {
        "min": max(18, age - random.randint(1, 4)),
        "max": min(40, age + random.randint(1, 5)),
    }
    height = choose_height(self_gender)
    height_range = {
        "min": max(140, height - random.randint(5, 20)),
        "max": min(210, height + random.randint(5, 20)),
    }

    base = {
        "meta": {
            "respondent_id": f"sim-{seed:02d}",
            "questionnaire_version": QUESTIONNAIRE_VERSION,
            "archetype": archetype,
        },
        "important_question_ids": random.sample(
            [12, 17, 19, 24, 25, 34, 37, 39, 42, 44, 46, 47, 48, 52],
            k=random.randint(4, 8),
        ),
        "answers": {
            "q1": self_gender,
            "q2": target_gender,
            "q3": random.choice(["寻找认真恋爱对象", "先认真认识，再决定是否发展关系", "寻找高质量新朋友"]),
            "q4": identity,
            "q5": stage,
            "q6": random.choice(CAMPUS_POOL),
            "q7": age,
            "q8": age_range,
            "q9": height,
            "q10": height_range,
            "q11": random.choice(["大概率继续留在北京", "可能去其他一线城市", "可能出国/出境", "暂时不确定"]),
            "q12": sample_scale(4),
            "q13": sample_scale(4),
            "q14": sample_scale(4),
            "q15": sample_scale(4),
            "q16": sample_scale(4),
            "q17": choose_many(TRAIT_POOL, 3, 5),
            "q18": choose_many(TRAIT_POOL, 3, 5),
            "q19": sample_scale(4),
            "q20": choose_many(SPARK_POOL, 2, 3),
            "q21": sample_scale(4),
            "q22": sample_scale(4),
            "q23": random.choice(["校园散步 / 自习 / 吃饭", "咖啡馆 / 展览 / City Walk", "运动 / 户外活动", "去市区认真约会"]),
            "q24": random.choice(["几乎每天", "一周 2-3 次", "一周 1 次", "两周及以上 1 次"]),
            "q25": sample_scale(4),
            "q26": sample_scale(4),
            "q27": sample_scale(4),
            "q28": choose_dietary(),
            "q29": random.choice(["从不", "偶尔", "经常"]),
            "q30": sample_scale(4),
            "q31": random.choice(["基本不喝", "社交场合偶尔喝", "比较常喝"]),
            "q32": choose_many(INTEREST_POOL, 3, 6),
            "q33": sample_scale(4),
            "q34": sample_scale(4),
            "q35": sample_scale(4),
            "q36": sample_scale(4),
            "q37": sample_scale(4),
            "q38": sample_scale(4),
            "q39": sample_scale(4),
            "q40": sample_scale(4),
            "q41": sample_scale(4),
            "q42": choose_many(SECURITY_POOL, 2, 3),
            "q43": choose_many(AFFECTION_POOL, 2, 3),
            "q44": sample_scale(4),
            "q45": sample_scale(4),
            "q46": sample_scale(5),
            "q47": sample_scale(4),
            "q48": sample_scale(4),
            "q49": sample_scale(4),
            "q50": random.choice(["两个独立的人并肩同行", "可以互相依赖、深度绑定", "视具体阶段动态调整"]),
            "q51": random.choice(["尊重彼此独立决定", "一起讨论后共同决定", "关键时刻以关系整体优先"]),
            "q52": choose_many(DEALBREAKER_POOL, 2, 3),
        },
    }

    if archetype == "slow_stable":
        base["answers"].update({"q34": 5, "q35": 5, "q37": 2, "q39": 6, "q44": 5, "q47": 6})
    elif archetype == "warm_listener":
        base["answers"].update({"q13": 6, "q17": ["温柔", "善于倾听", "真诚", "稳定"], "q42": ["稳定回复和出现", "尊重我的边界", "主动规划见面"]})
    elif archetype == "urban_ambitious":
        base["answers"].update({"q22": 6, "q44": 6, "q45": 6, "q46": 7, "q49": 3})
    elif archetype == "cool_boundaries":
        base["answers"].update({"q14": 6, "q16": 2, "q25": 2, "q34": 2, "q41": 6})
    elif archetype == "gentle_romantic":
        base["answers"].update({"q19": 6, "q20": ["被坚定选择", "被温柔照顾", "被欣赏和夸奖"], "q39": 5})
    elif archetype == "academic_partner":
        base["answers"].update({"q15": 6, "q17": ["聪明", "真诚", "有担当", "有自己的热爱"], "q32": ["阅读", "学术讨论", "编程 / 科技", "展览 / 博物馆"]})
    elif archetype == "independent_explorer":
        base["answers"].update({"q23": "运动 / 户外活动", "q25": 2, "q32": ["旅行", "户外 / 徒步", "摄影 / 视频", "创作 / 写作"], "q50": "两个独立的人并肩同行"})
    elif archetype == "high_aesthetic":
        base["answers"].update({"q19": 7, "q17": ["有审美", "聪明", "温柔", "有自己的热爱"], "q23": "咖啡馆 / 展览 / City Walk"})
    elif archetype == "secure_builder":
        base["answers"].update({"q34": 2, "q35": 3, "q38": 6, "q42": ["稳定回复和出现", "冲突时不失联", "尊重我的边界"]})
    elif archetype == "thoughtful_late_bloomer":
        base["answers"].update({"q13": 5, "q36": 3, "q39": 7, "q43": ["记住细节并回应", "分享我的真实生活", "默默付出和帮忙"]})

    return base


def generate_simulated_responses(count: int = 10) -> list[dict[str, Any]]:
    return [make_response(seed) for seed in range(1, count + 1)]


def save_simulated_responses(path: Path, count: int = 10) -> None:
    responses = generate_simulated_responses(count=count)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(responses, ensure_ascii=False, indent=2), encoding="utf-8")

