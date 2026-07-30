from __future__ import annotations

from typing import Any

from .schema import Traits


def bucket_scale(value: int, low: str, mid: str, high: str) -> str:
    if value <= 3:
        return low
    if value >= 5:
        return high
    return mid


def derive_traits(response: dict[str, Any]) -> Traits:
    answers = response["answers"]
    important = response.get("important_question_ids", [])

    return Traits(
        relationship_goal={
            "ç€µç»˜å£˜ç’ã‚‡æ¹¡éŽ­å¬¬åŸç€µç¡…è–„": "serious_relationship",
            "éå £î…»éªç†»î…»ç’‡å—­ç´éå¶…å–…ç€¹æ°­æ§¸éšï¹€å½‚çžæ›žå§ç»¯?": "slow_exploration",
            "ç€µç»˜å£˜æ¥‚æ¨¿å·é–²å¿”æŸŠéˆå¬ªå¼¸": "high_quality_friendship",
        }.get(answers["q3"], "slow_exploration"),
        stage_group=answers["q5"],
        location_stability=answers["q11"],
        long_distance_tolerance=bucket_scale(answers["q12"], "low", "medium", "high"),
        social_energy=bucket_scale(8 - answers["q13"], "introverted", "balanced", "extroverted"),
        attraction_tone=bucket_scale(8 - answers["q14"], "cool", "balanced", "warm"),
        expression_preference=bucket_scale(answers["q16"], "reserved", "balanced", "expressive"),
        attraction_keywords=answers["q17"],
        self_keywords=answers["q18"],
        spark_keywords=answers["q20"],
        aesthetic_importance=bucket_scale(answers["q19"], "low", "medium", "high"),
        life_pace=bucket_scale(answers["q22"], "relaxed", "balanced", "intense"),
        date_style=answers["q23"],
        meeting_frequency=answers["q24"],
        autonomy_need=bucket_scale(8 - answers["q25"], "high_closeness", "balanced", "high_space"),
        order_need=bucket_scale(answers["q26"], "low", "balanced", "high"),
        spending_style=bucket_scale(answers["q27"], "frugal", "balanced", "experience_oriented"),
        smoking_boundary=bucket_scale(8 - answers["q30"], "strict", "neutral", "open"),
        core_interests=answers["q32"],
        interest_overlap_need=bucket_scale(answers["q33"], "low", "balanced", "high"),
        reassurance_need=bucket_scale(answers["q34"], "low", "balanced", "high"),
        insecurity_pattern=bucket_scale(answers["q35"], "self_process", "balanced", "seek_reassurance"),
        vulnerability_style=bucket_scale(answers["q36"], "guarded", "gradual", "open"),
        conflict_style=bucket_scale(answers["q37"], "cool_down_first", "balanced", "solve_immediately"),
        repair_style=bucket_scale(answers["q38"], "wait_other", "balanced", "initiate_repair"),
        intimacy_timing=bucket_scale(answers["q39"], "early", "medium", "late"),
        relationship_role=bucket_scale(answers["q40"], "care_for_other", "balanced", "want_to_be_cared_for"),
        jealousy_boundary=bucket_scale(answers["q41"], "strict", "balanced", "trusting"),
        security_sources=answers["q42"],
        affection_style=answers["q43"],
        career_drive=bucket_scale(answers["q44"], "life_first", "balanced", "career_first"),
        preferred_partner_drive=bucket_scale(answers["q45"], "life_first", "balanced", "career_first"),
        future_lifestyle=bucket_scale(answers["q46"], "stable", "balanced", "high_density_city"),
        marriage_orientation=bucket_scale(answers["q47"], "open", "considering", "clear_yes"),
        children_orientation=bucket_scale(answers["q48"], "open", "considering", "clear_yes"),
        principle_orientation=bucket_scale(answers["q49"], "result_first", "balanced", "principle_first"),
        relationship_model={
            "æ¶“ã‚„é‡œé™î„‚ç›é¨å‹ªæ±‰éªžæƒ°åéšå²ƒî”‘": "parallel_independent",
            "é™îˆ™äº’æµœæŽ”æµ‰æ¸šæ¿Šç¦†éŠ†ä½¹ç¹æ´ï¸¾ç²¦ç€¹?": "deeply_interdependent",
            "ç‘™å——å¿æµ£æ’»æ¨å¨ˆé›å§©éŽ¬ä½½çšŸé?": "adaptive",
        }.get(answers["q50"], "adaptive"),
        decision_model={
            "çå©‡å™¸è¤°å…¼î„é™î„‚ç›éå†²ç•¾": "independent",
            "æ¶“â‚¬ç’§ç–¯î…¿ç’å“„æ‚—éåžæ‚“éå†²ç•¾": "collaborative",
            "éæŠ½æ•­éƒè·ºåŸ¢æµ ãƒ¥å§ç»¯ç»˜æš£æµ£æ’²ç´­é?": "relationship_first",
        }.get(answers["q51"], "collaborative"),
        dealbreakers=answers["q52"],
        important_question_ids=important,
    )
