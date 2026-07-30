"""
Match quality reporting and result output.
"""
from __future__ import annotations
import csv
import math
from .models import User
from .scoring import compute_score, compute_score_breakdown


def display_score(raw: float) -> float:
    """Transform raw score to display score via sqrt for nicer-looking numbers."""
    return math.sqrt(raw) if raw >= 0 else 0.0


def grade_match(quality: float) -> str:
    dq = display_score(quality)
    if dq >= 0.837:
        return "S"
    elif dq >= 0.742:
        return "A"
    elif dq >= 0.632:
        return "B"
    return "C"


def analyze_matches(
    matches: list[tuple[int, int, float, float]],
    pool_name: str,
) -> dict:
    if not matches:
        return {
            "pool": pool_name,
            "total_pairs": 0,
            "avg_quality": 0,
            "median_quality": 0,
            "grade_dist": {},
        }

    qualities = [display_score(min(ps, as_)) for _, _, ps, as_ in matches]
    qualities.sort()

    grades = {}
    for q in qualities:
        if q >= 0.837:
            g = "S"
        elif q >= 0.742:
            g = "A"
        elif q >= 0.632:
            g = "B"
        else:
            g = "C"
        grades[g] = grades.get(g, 0) + 1

    n = len(qualities)
    median = qualities[n // 2] if n % 2 == 1 else (qualities[n // 2 - 1] + qualities[n // 2]) / 2

    return {
        "pool": pool_name,
        "total_pairs": n,
        "avg_quality": sum(qualities) / n,
        "median_quality": median,
        "min_quality": qualities[0],
        "max_quality": qualities[-1],
        "grade_dist": grades,
    }


def find_unmatched(
    users: list[User],
    matched_ids: set[int],
    all_users_map: dict[int, User],
    prefs: dict[int, list[tuple[int, float]]] | None = None,
) -> list[dict]:
    unmatched = []
    for u in users:
        if u.user_id not in matched_ids:
            info = {
                "user_id": u.user_id,
                "nickname": u.nickname,
                "reason": "preference_list_empty" if not prefs or not prefs.get(u.user_id) else "competition_eliminated",
            }
            if prefs and prefs.get(u.user_id):
                top3 = prefs[u.user_id][:3]
                info["top3_candidates"] = [
                    {
                        "user_id": cid,
                        "nickname": all_users_map[cid].nickname if cid in all_users_map else "?",
                        "score": round(s, 4),
                    }
                    for cid, s in top3
                ]
            unmatched.append(info)
    return unmatched


def print_report(
    pool_name: str,
    matches: list[tuple[int, int, float, float]],
    users_map: dict[int, User],
):
    print(f"\n{'='*70}")
    print(f"  Pool: {pool_name}")
    print(f"{'='*70}")

    stats = analyze_matches(matches, pool_name)
    print(f"  Total pairs:    {stats['total_pairs']}")
    if stats["total_pairs"] > 0:
        print(f"  Avg quality:    {stats['avg_quality']:.4f}")
        print(f"  Median quality: {stats['median_quality']:.4f}")
        print(f"  Min quality:    {stats['min_quality']:.4f}")
        print(f"  Max quality:    {stats['max_quality']:.4f}")
        print(f"  Grade dist:     {stats['grade_dist']}")

    print(f"\n  {'ID':>5} {'Nickname A':<14} {'ID':>5} {'Nickname B':<14} {'A→B':>6} {'B→A':>6} {'Quality':>7} {'Grade':>5}")
    print(f"  {'-'*5} {'-'*14} {'-'*5} {'-'*14} {'-'*6} {'-'*6} {'-'*7} {'-'*5}")

    sorted_matches = sorted(matches, key=lambda x: -min(x[2], x[3]))
    for a_id, b_id, sa, sb in sorted_matches:
        dsa = display_score(sa)
        dsb = display_score(sb)
        dq = min(dsa, dsb)
        g = grade_match(min(sa, sb))
        na = users_map[a_id].nickname if a_id in users_map else "?"
        nb = users_map[b_id].nickname if b_id in users_map else "?"
        gap_flag = " ⚠" if abs(dsa - dsb) > 0.15 else ""
        print(f"  {a_id:>5} {na:<14} {b_id:>5} {nb:<14} {dsa:>6.3f} {dsb:>6.3f} {dq:>7.3f} {g:>5}{gap_flag}")


def export_results_csv(
    all_matches: list[tuple[str, list[tuple[int, int, float, float]]]],
    users_map: dict[int, User],
    output_path: str,
):
    with open(output_path, "w", encoding="utf-8", newline="") as f:
        writer = csv.writer(f)
        writer.writerow([
            "pool", "user_a_id", "user_a_nickname", "user_a_email",
            "user_b_id", "user_b_nickname", "user_b_email",
            "score_a_to_b", "score_b_to_a", "match_quality", "grade", "asymmetry_warning",
        ])
        for pool_name, matches in all_matches:
            for a_id, b_id, sa, sb in matches:
                dsa = display_score(sa)
                dsb = display_score(sb)
                dq = min(dsa, dsb)
                ua = users_map.get(a_id)
                ub = users_map.get(b_id)
                writer.writerow([
                    pool_name,
                    a_id, ua.nickname if ua else "", ua.email if ua else "",
                    b_id, ub.nickname if ub else "", ub.email if ub else "",
                    round(dsa, 4), round(dsb, 4), round(dq, 4),
                    grade_match(min(sa, sb)),
                    "yes" if abs(dsa - dsb) > 0.15 else "",
                ])
    print(f"\nResults exported to {output_path}")


def print_detailed_pair(a: User, b: User):
    """Print detailed score breakdown for a specific pair."""
    bd_ab = compute_score_breakdown(a, b)
    bd_ba = compute_score_breakdown(b, a)

    print(f"\n--- Detailed breakdown: {a.nickname} (#{a.user_id}) ↔ {b.nickname} (#{b.user_id}) ---")
    print(f"  {'Dimension':<25} {'A→B':>8} {'B→A':>8}")
    print(f"  {'-'*25} {'-'*8} {'-'*8}")
    dim_names = {
        "D1_purpose_stage": "D1 Purpose & Stage",
        "D2_geography": "D2 Geography",
        "D3_attraction": "D3 Attraction",
        "D4_lifestyle": "D4 Lifestyle",
        "D5_interests": "D5 Interests",
        "D6_attachment": "D6 Attachment",
        "D7_future": "D7 Future Vision",
    }
    for key, label in dim_names.items():
        print(f"  {label:<25} {bd_ab['dimensions'][key]:>8.4f} {bd_ba['dimensions'][key]:>8.4f}")
    print(f"  {'Penalty':<25} {bd_ab['penalty']:>8.2f} {bd_ba['penalty']:>8.2f}")
    print(f"  {'FINAL SCORE':<25} {bd_ab['final_score']:>8.4f} {bd_ba['final_score']:>8.4f}")
    print(f"  Match Quality: {min(bd_ab['final_score'], bd_ba['final_score']):.4f} ({grade_match(min(bd_ab['final_score'], bd_ba['final_score']))})")
