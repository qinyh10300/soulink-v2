"""
Seven-dimension compatibility scoring system.

All dimension scorers return a value in [0, 1].
Functions named score_*_asym(a, b) compute A's preference for B (may ≠ B's for A).
Functions named score_*_sym(a, b) compute a symmetric score.
"""
from __future__ import annotations

from .models import User
from .config import (
    IMPORTANCE_MULTIPLIER,
    DIMENSION_WEIGHTS,
    PURPOSE_MATRIX,
    UNIVERSITY_GROUPS,
    STAGE_ORDER,
    FUTURE_LOC_ORDER,
    FREQ_ORDER,
    SMOKING_INTENSITY,
    DRINK_ORDER,
    REL_MODEL_MATRIX,
    DECISION_MATRIX,
    SECURITY_EXPRESSION_MAP,
)


# ---------------------------------------------------------------------------
# Primitive scoring functions
# ---------------------------------------------------------------------------

def likert_sim(a, b) -> float:
    if a is None or b is None:
        return 0.5
    return 1.0 - abs(a - b) / 6.0


def likert_comp(a, b) -> float:
    if a is None or b is None:
        return 0.5
    return 1.0 - abs(a + b - 8) / 6.0


def jaccard(set_a: set | None, set_b: set | None) -> float:
    if not set_a or not set_b:
        return 0.0
    union = set_a | set_b
    if not union:
        return 0.0
    return len(set_a & set_b) / len(union)


def cross_match(want: set | None, offer: set | None) -> float:
    if not want or not offer:
        return 0.0
    matched = want & offer
    return len(matched) / len(want)


def cat_match(a, b, matrix: dict) -> float:
    if a is None or b is None:
        return 0.5
    row = matrix.get(a)
    if row is None:
        return 0.3
    return row.get(b, 0.3)


def _imp(user: User, q: int) -> float:
    return IMPORTANCE_MULTIPLIER if user.imp(q) else 1.0


def _weighted_avg(items: list[tuple[float, float]]) -> float:
    """Weighted average from list of (score, weight) tuples."""
    total_w = sum(w for _, w in items)
    if total_w == 0:
        return 0.5
    return sum(s * w for s, w in items) / total_w


# ---------------------------------------------------------------------------
# D1: Purpose & Stage (symmetric)
# ---------------------------------------------------------------------------

def _university_preference_score(a_unis: set | None, b_unis: set | None) -> float:
    """
    Score university preference compatibility using jaccard overlap,
    with a bonus for universities in the same geographic group.
    """
    if not a_unis or not b_unis:
        return 0.5
    base = jaccard(a_unis, b_unis)
    a_groups = {UNIVERSITY_GROUPS.get(u, u) for u in a_unis}
    b_groups = {UNIVERSITY_GROUPS.get(u, u) for u in b_unis}
    group_overlap = jaccard(a_groups, b_groups)
    return 0.6 * base + 0.4 * group_overlap


def score_d1(a: User, b: User) -> float:
    items = []

    # Q3 — purpose (w=0.50)
    s_q3 = cat_match(a.q(3), b.q(3), PURPOSE_MATRIX)
    items.append((s_q3, 0.50 * max(_imp(a, 3), _imp(b, 3))))

    # Q4/Q5 — stage (w=0.35)
    a_stage = STAGE_ORDER.get(a.q(5))
    b_stage = STAGE_ORDER.get(b.q(5))
    if a_stage is not None and b_stage is not None:
        s_stage = 1.0 - abs(a_stage - b_stage) / 4.0
    else:
        s_stage = 0.5
    items.append((s_stage, 0.35 * max(_imp(a, 5), _imp(b, 5))))

    # Q6 — university preferences (w=0.15, reduced since university is now a hard filter)
    s_q6 = _university_preference_score(a.q(6), b.q(6))
    items.append((s_q6, 0.15 * max(_imp(a, 6), _imp(b, 6))))

    return _weighted_avg(items)


# ---------------------------------------------------------------------------
# D2: Geography & Distance (symmetric)
# ---------------------------------------------------------------------------

def score_d2(a: User, b: User) -> float:
    items = []

    # Q11 — future location (w=0.50)
    a_loc = FUTURE_LOC_ORDER.get(a.q(11))
    b_loc = FUTURE_LOC_ORDER.get(b.q(11))
    same_q11 = (a.q(11) == b.q(11)) and a.q(11) is not None

    if a_loc is None or b_loc is None:
        s_q11 = 0.5
    else:
        s_q11 = 1.0 - abs(a_loc - b_loc) / 4.0
    items.append((s_q11, 0.50 * max(_imp(a, 11), _imp(b, 11))))

    # Q12 — long distance tolerance (w=0.50)
    if same_q11:
        s_q12 = 1.0
    else:
        a_ld = a.q(12) or 4
        b_ld = b.q(12) or 4
        s_q12 = min(a_ld, b_ld) / 7.0
    items.append((s_q12, 0.50 * max(_imp(a, 12), _imp(b, 12))))

    return _weighted_avg(items)


# ---------------------------------------------------------------------------
# D3: Attraction & Temperament (asymmetric via Q17↔Q18)
# ---------------------------------------------------------------------------

def score_d3(a: User, b: User) -> float:
    """Score from A's perspective: how attractive B is to A."""
    items = []

    # Q13 — social style similarity (w=0.10)
    items.append((likert_sim(a.q(13), b.q(13)), 0.10 * _imp(a, 13)))

    # Q14 — temperament preference similarity (w=0.10)
    items.append((likert_sim(a.q(14), b.q(14)), 0.10 * _imp(a, 14)))

    # Q15 — "earnest worker" attraction (w=0.05)
    items.append((likert_sim(a.q(15), b.q(15)), 0.05 * _imp(a, 15)))

    # Q16 — expressiveness preference (w=0.05)
    items.append((likert_sim(a.q(16), b.q(16)), 0.05 * _imp(a, 16)))

    # Q17 ↔ Q18 — trait cross-match (w=0.35) ⭐
    # What A wants (Q17) vs what B offers (Q18)
    s_trait = cross_match(a.q(17), b.q(18))
    items.append((s_trait, 0.35 * _imp(a, 17)))

    # Q19 — appearance importance similarity (w=0.10)
    items.append((likert_sim(a.q(19), b.q(19)), 0.10 * _imp(a, 19)))

    # Q20 — heartfelt moments (w=0.25)
    items.append((jaccard(a.q(20), b.q(20)), 0.25 * _imp(a, 20)))

    return _weighted_avg(items)


# ---------------------------------------------------------------------------
# D4: Lifestyle (asymmetric via Q29↔Q30)
# ---------------------------------------------------------------------------

def score_d4(a: User, b: User) -> float:
    """Score from A's perspective: can A accept B's lifestyle."""
    items = []

    # Q21 — schedule (w=0.10)
    items.append((likert_sim(a.q(21), b.q(21)), 0.10 * _imp(a, 21)))

    # Q22 — life pace (w=0.08)
    items.append((likert_sim(a.q(22), b.q(22)), 0.08 * _imp(a, 22)))

    # Q23 — date activities (w=0.12)
    items.append((jaccard(a.q(23), b.q(23)), 0.12 * _imp(a, 23)))

    # Q24 — meeting frequency (w=0.10)
    a_freq = FREQ_ORDER.get(a.q(24))
    b_freq = FREQ_ORDER.get(b.q(24))
    if a_freq is not None and b_freq is not None:
        s_q24 = 1.0 - abs(a_freq - b_freq) / 3.0
    else:
        s_q24 = 0.5
    items.append((s_q24, 0.10 * _imp(a, 24)))

    # Q25 — personal space (w=0.10)
    items.append((likert_sim(a.q(25), b.q(25)), 0.10 * _imp(a, 25)))

    # Q26 — tidiness (w=0.08)
    items.append((likert_sim(a.q(26), b.q(26)), 0.08 * _imp(a, 26)))

    # Q27 — spending (w=0.10)
    items.append((likert_sim(a.q(27), b.q(27)), 0.10 * _imp(a, 27)))

    # Q28 — diet (w=0.05)
    a_diet = a.q(28)
    b_diet = b.q(28)
    if not a_diet or not b_diet:
        s_q28 = 0.5
    elif {"无"} == a_diet or {"无"} == b_diet:
        s_q28 = 1.0
    elif "无" in a_diet or "无" in b_diet:
        s_q28 = 1.0
    else:
        s_q28 = jaccard(a_diet, b_diet)
    items.append((s_q28, 0.05 * _imp(a, 28)))

    # Q29 ↔ Q30 — smoking cross-match (w=0.12) ⭐
    # Can A accept B's smoking?
    b_smoke_level = SMOKING_INTENSITY.get(b.q(29), 0)
    if b_smoke_level == 0:
        s_smoke = 1.0
    else:
        a_tol = a.q(30)
        s_smoke = (a_tol / 7.0) if a_tol is not None else 0.5
    items.append((s_smoke, 0.12 * _imp(a, 30)))

    # Q31 — drinking (w=0.05)
    a_drink = DRINK_ORDER.get(a.q(31))
    b_drink = DRINK_ORDER.get(b.q(31))
    if a_drink is not None and b_drink is not None:
        s_q31 = 1.0 - abs(a_drink - b_drink) / 2.0
    else:
        s_q31 = 0.5
    items.append((s_q31, 0.05 * _imp(a, 31)))

    # Q24 + Q25 consistency bonus (w=0.10)
    items.append(((s_q24 + likert_sim(a.q(25), b.q(25))) / 2.0, 0.10))

    return _weighted_avg(items)


# ---------------------------------------------------------------------------
# D5: Interests (symmetric, modulated by Q33)
# ---------------------------------------------------------------------------

def score_d5(a: User, b: User) -> float:
    s_q32 = jaccard(a.q(32), b.q(32))

    a_desire = a.q(33) or 4
    b_desire = b.q(33) or 4
    interest_desire = (a_desire + b_desire) / 14.0

    return s_q32 * (0.5 + 0.5 * interest_desire)


# ---------------------------------------------------------------------------
# D6: Attachment & Intimacy (asymmetric via Q42↔Q43)
# ---------------------------------------------------------------------------

def _security_cross_score(need_set: set | None, expr_set: set | None) -> float:
    """
    For each item A needs (Q42), find best mapping to B's expression (Q43).
    Average of those best scores.
    """
    if not need_set or not expr_set:
        return 0.3

    scores = []
    for need_item in need_set:
        mapping = SECURITY_EXPRESSION_MAP.get(need_item, {})
        best = 0.0
        for expr_item in expr_set:
            best = max(best, mapping.get(expr_item, 0.1))
        scores.append(best)

    return sum(scores) / len(scores) if scores else 0.3


def score_d6(a: User, b: User) -> float:
    """Score from A's perspective."""
    items = []

    # Q34 — reply anxiety (w=0.08)
    items.append((likert_sim(a.q(34), b.q(34)), 0.08 * _imp(a, 34)))

    # Q35 — when anxious (w=0.10)
    items.append((likert_sim(a.q(35), b.q(35)), 0.10 * _imp(a, 35)))

    # Q36 — vulnerability (w=0.08)
    items.append((likert_sim(a.q(36), b.q(36)), 0.08 * _imp(a, 36)))

    # Q37 — conflict style with floor (w=0.10)
    s_q37 = max(likert_sim(a.q(37), b.q(37)), 0.4)
    items.append((s_q37, 0.10 * _imp(a, 37)))

    # Q38 — reconciliation willingness (w=0.08) ⚠️ special
    a38 = a.q(38)
    b38 = b.q(38)
    if a38 is not None and b38 is not None:
        if a38 <= 2 and b38 <= 2:
            s_q38 = 0.2
        else:
            s_q38 = max(a38, b38) / 7.0
    else:
        s_q38 = 0.5
    items.append((s_q38, 0.08 * _imp(a, 38)))

    # Q39 — intimacy timeline (w=0.08)
    items.append((likert_sim(a.q(39), b.q(39)), 0.08 * _imp(a, 39)))

    # Q40 — caretaking role (w=0.10) ⚠️ complementary
    items.append((likert_comp(a.q(40), b.q(40)), 0.10 * _imp(a, 40)))

    # Q41 — opposite-sex friend tolerance (w=0.08)
    items.append((likert_sim(a.q(41), b.q(41)), 0.08 * _imp(a, 41)))

    # Q42 ↔ Q43 — security-expression cross (w=0.30) ⭐
    # What A needs (Q42) vs how B expresses (Q43)
    s_sec = _security_cross_score(a.q(42), b.q(43))
    items.append((s_sec, 0.30 * _imp(a, 42)))

    return _weighted_avg(items)


# ---------------------------------------------------------------------------
# D7: Future Vision (asymmetric via Q45↔Q44)
# ---------------------------------------------------------------------------

def score_d7(a: User, b: User) -> float:
    """Score from A's perspective."""
    items = []

    # Q44 — work-life balance similarity (w=0.08)
    items.append((likert_sim(a.q(44), b.q(44)), 0.08 * _imp(a, 44)))

    # Q45 ↔ Q44 — A's ambition expectation vs B's actual (w=0.10) ⭐
    s_q45 = likert_sim(a.q(45), b.q(44))
    items.append((s_q45, 0.10 * _imp(a, 45)))

    # Q46 — future lifestyle (w=0.08)
    items.append((likert_sim(a.q(46), b.q(46)), 0.08 * _imp(a, 46)))

    # Q47 — marriage attitude (w=0.18) ⭐
    items.append((likert_sim(a.q(47), b.q(47)), 0.18 * _imp(a, 47)))

    # Q48 — children attitude (w=0.18) ⭐
    items.append((likert_sim(a.q(48), b.q(48)), 0.18 * _imp(a, 48)))

    # Q49 — pragmatism vs principles (w=0.08)
    items.append((likert_sim(a.q(49), b.q(49)), 0.08 * _imp(a, 49)))

    # Q50 — relationship model (w=0.10)
    s_q50 = cat_match(a.q(50), b.q(50), REL_MODEL_MATRIX)
    items.append((s_q50, 0.10 * _imp(a, 50)))

    # Q51 — decision style (w=0.08)
    s_q51 = cat_match(a.q(51), b.q(51), DECISION_MATRIX)
    items.append((s_q51, 0.08 * _imp(a, 51)))

    # Q52 — dealbreaker consensus (w=0.12)
    s_q52 = jaccard(a.q(52), b.q(52))
    items.append((s_q52, 0.12 * _imp(a, 52)))

    return _weighted_avg(items)


# ---------------------------------------------------------------------------
# Marriage & fertility soft veto
# ---------------------------------------------------------------------------

def _marriage_fertility_penalty(a: User, b: User) -> float:
    penalty = 1.0
    a47, b47 = a.q(47), b.q(47)
    if a47 is not None and b47 is not None and abs(a47 - b47) >= 5:
        penalty *= 0.5
    a48, b48 = a.q(48), b.q(48)
    if a48 is not None and b48 is not None and abs(a48 - b48) >= 5:
        penalty *= 0.5
    return penalty


# ---------------------------------------------------------------------------
# Score boost for previously unmatched users
# ---------------------------------------------------------------------------

_score_boosts: dict[int, float] = {}


def set_score_boosts(boosts: dict[int, float]):
    """Set per-user score multipliers (e.g. 1.05 for one-round unmatched)."""
    global _score_boosts
    _score_boosts = boosts


# ---------------------------------------------------------------------------
# Total score: A's preference for B
# ---------------------------------------------------------------------------

def compute_score(a: User, b: User) -> float:
    W = DIMENSION_WEIGHTS
    raw = (
        W["D1"] * score_d1(a, b)
        + W["D2"] * score_d2(a, b)
        + W["D3"] * score_d3(a, b)
        + W["D4"] * score_d4(a, b)
        + W["D5"] * score_d5(a, b)
        + W["D6"] * score_d6(a, b)
        + W["D7"] * score_d7(a, b)
    )
    score = raw * _marriage_fertility_penalty(a, b)
    boost = max(_score_boosts.get(a.user_id, 1.0), _score_boosts.get(b.user_id, 1.0))
    return score * boost


def compute_score_breakdown(a: User, b: User) -> dict:
    W = DIMENSION_WEIGHTS
    d_scores = {
        "D1_purpose_stage": score_d1(a, b),
        "D2_geography": score_d2(a, b),
        "D3_attraction": score_d3(a, b),
        "D4_lifestyle": score_d4(a, b),
        "D5_interests": score_d5(a, b),
        "D6_attachment": score_d6(a, b),
        "D7_future": score_d7(a, b),
    }
    penalty = _marriage_fertility_penalty(a, b)
    raw = sum(W[k[:2]] * v for k, v in d_scores.items())
    return {
        "dimensions": d_scores,
        "penalty": penalty,
        "raw_score": raw,
        "final_score": raw * penalty,
    }
