from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any


QUESTIONNAIRE_VERSION = "thu_v1"


QUESTION_IDS = {
    "self_gender": 1,
    "target_gender": 2,
    "goal": 3,
    "identity": 4,
    "stage": 5,
    "campus": 6,
    "age": 7,
    "target_age_range": 8,
    "height": 9,
    "target_height_range": 10,
    "future_location": 11,
    "long_distance": 12,
    "social_style": 13,
    "attraction_tone": 14,
    "serious_doer": 15,
    "expressive_partner": 16,
    "attracted_traits": 17,
    "self_traits": 18,
    "aesthetic_need": 19,
    "spark_moments": 20,
    "schedule": 21,
    "life_pace": 22,
    "date_style": 23,
    "meeting_frequency": 24,
    "space_need": 25,
    "orderliness": 26,
    "spending_style": 27,
    "smoking_acceptance": 30,
    "interests": 32,
    "interest_overlap": 33,
    "reply_anxiety": 34,
    "insecurity_response": 35,
    "vulnerability": 36,
    "conflict_response": 37,
    "repair_style": 38,
    "sexual_timing": 39,
    "relationship_role": 40,
    "opposite_sex_friend_acceptance": 41,
    "security_sources": 42,
    "love_expression": 43,
    "career_orientation": 44,
    "partner_career_orientation": 45,
    "life_after_graduation": 46,
    "marriage_attitude": 47,
    "children_attitude": 48,
    "principle_vs_interest": 49,
    "relationship_model": 50,
    "major_decisions": 51,
    "dealbreakers": 52,
}


CORE_TRAIT_QUESTION_IDS = [
    3,
    5,
    11,
    12,
    13,
    14,
    16,
    17,
    18,
    19,
    20,
    22,
    23,
    24,
    25,
    26,
    27,
    30,
    32,
    33,
    34,
    35,
    36,
    37,
    38,
    39,
    40,
    41,
    42,
    43,
    44,
    45,
    46,
    47,
    48,
    49,
    50,
    51,
    52,
]


@dataclass
class Traits:
    relationship_goal: str
    stage_group: str
    location_stability: str
    long_distance_tolerance: str
    social_energy: str
    attraction_tone: str
    expression_preference: str
    attraction_keywords: list[str] = field(default_factory=list)
    self_keywords: list[str] = field(default_factory=list)
    spark_keywords: list[str] = field(default_factory=list)
    aesthetic_importance: str = "medium"
    life_pace: str = "balanced"
    date_style: str = "balanced"
    meeting_frequency: str = "weekly"
    autonomy_need: str = "balanced"
    order_need: str = "balanced"
    spending_style: str = "balanced"
    smoking_boundary: str = "neutral"
    core_interests: list[str] = field(default_factory=list)
    interest_overlap_need: str = "balanced"
    reassurance_need: str = "balanced"
    insecurity_pattern: str = "balanced"
    vulnerability_style: str = "gradual"
    conflict_style: str = "balanced"
    repair_style: str = "balanced"
    intimacy_timing: str = "medium"
    relationship_role: str = "balanced"
    jealousy_boundary: str = "balanced"
    security_sources: list[str] = field(default_factory=list)
    affection_style: list[str] = field(default_factory=list)
    career_drive: str = "balanced"
    preferred_partner_drive: str = "balanced"
    future_lifestyle: str = "balanced"
    marriage_orientation: str = "open"
    children_orientation: str = "open"
    principle_orientation: str = "balanced"
    relationship_model: str = "balanced"
    decision_model: str = "balanced"
    dealbreakers: list[str] = field(default_factory=list)
    important_question_ids: list[int] = field(default_factory=list)

    def to_dict(self) -> dict[str, Any]:
        return self.__dict__.copy()


TRAITS_SCHEMA = {
    "relationship_goal": ["serious_relationship", "slow_exploration", "high_quality_friendship"],
    "long_distance_tolerance": ["low", "medium", "high"],
    "social_energy": ["introverted", "balanced", "extroverted"],
    "attraction_tone": ["warm", "balanced", "cool"],
    "aesthetic_importance": ["low", "medium", "high"],
    "life_pace": ["relaxed", "balanced", "intense"],
    "autonomy_need": ["high_space", "balanced", "high_closeness"],
    "reassurance_need": ["low", "balanced", "high"],
    "conflict_style": ["cool_down_first", "balanced", "solve_immediately"],
    "intimacy_timing": ["early", "medium", "late"],
    "career_drive": ["life_first", "balanced", "career_first"],
    "future_lifestyle": ["stable", "balanced", "high_density_city"],
}
