#!/usr/bin/env python3
"""
Soulink Matching System — Main Entry Point (Round 3)

Usage:
    python run_matching.py [--csv PATH] [--min-score FLOAT] [--output PATH]
                           [--prev-results PATH,...] [--priority-emails EMAIL,...]
"""
import argparse
import os
import sys

from matching.parser import (
    load_users,
    filter_valid_users,
    load_previously_matched_pairs,
    load_previously_matched_emails,
)
from matching.pools import assign_pools
from matching.algorithms import (
    match_mf_pool,
    match_same_sex_pool,
    build_preference_lists,
)
from matching.report import (
    print_report,
    export_results_csv,
    find_unmatched,
)
from matching.scoring import compute_score, set_score_boosts
from matching.filters import pass_all_filters, set_blocked_pairs
from matching.config import MIN_ACCEPTABLE_SCORE


def _pre_match_priority(
    priority_users: list,
    candidate_pool: list,
    users_map: dict,
) -> tuple[list[tuple[int, int, float, float]], set[int]]:
    """
    Pre-matching: match priority users FIRST before the main algorithm.
    Each priority user gets paired with their best compatible candidate.
    Returns (matches, matched_ids) so those users can be removed from pools.
    """
    matches = []
    matched_ids = set()

    for pu in priority_users:
        if pu.user_id in matched_ids:
            continue
        best_cid, best_sa, best_sb = None, 0.0, 0.0
        for candidate in candidate_pool:
            if candidate.user_id == pu.user_id:
                continue
            if candidate.user_id in matched_ids:
                continue
            if not pass_all_filters(pu, candidate):
                continue
            sa = compute_score(pu, candidate)
            sb = compute_score(candidate, pu)
            quality = min(sa, sb)
            if quality > min(best_sa, best_sb):
                best_cid, best_sa, best_sb = candidate.user_id, sa, sb
        if best_cid is not None:
            matches.append((pu.user_id, best_cid, best_sa, best_sb))
            matched_ids.add(pu.user_id)
            matched_ids.add(best_cid)
            c_name = users_map.get(best_cid)
            print(f"  [Priority] Pre-matched #{pu.user_id} {pu.nickname} "
                  f"↔ #{best_cid} {c_name.nickname if c_name else '?'} "
                  f"(quality={min(best_sa, best_sb):.3f})")
        else:
            print(f"  [Priority] WARNING: Could not find any match for #{pu.user_id} {pu.nickname}")

    return matches, matched_ids


def _compute_score_boosts(
    valid_users: list,
    prev_results_files: list[str],
    prev_data_files: list[str],
) -> dict[int, float]:
    """Compute score boost multipliers for users who weren't matched in previous rounds.
    1 round unmatched = 1.05x, 2 rounds = 1.10x, 3+ rounds = 1.15x."""
    if not prev_results_files:
        return {}

    matched_per_round = []
    for path in prev_results_files:
        matched_per_round.append(load_previously_matched_emails(path))

    participated_per_round = []
    for path in prev_data_files:
        participated = set()
        try:
            for u in filter_valid_users(load_users(path)):
                participated.add(u.email)
        except FileNotFoundError:
            pass
        participated_per_round.append(participated)

    boosts = {}
    for u in valid_users:
        unmatched_rounds = 0
        for i in range(min(len(matched_per_round), len(participated_per_round))):
            if u.email in participated_per_round[i] and u.email not in matched_per_round[i]:
                unmatched_rounds += 1

        if unmatched_rounds >= 3:
            boosts[u.user_id] = 1.15
        elif unmatched_rounds == 2:
            boosts[u.user_id] = 1.10
        elif unmatched_rounds == 1:
            boosts[u.user_id] = 1.05

    return boosts


def main():
    parser = argparse.ArgumentParser(description="Soulink Matching System (Round 3)")
    data_dir = os.path.join(os.path.dirname(__file__), "data")
    parser.add_argument(
        "--csv",
        default=os.path.join(data_dir, "export_thu-date_v0.6.csv"),
        help="Path to questionnaire export CSV",
    )
    parser.add_argument(
        "--min-score",
        type=float,
        default=MIN_ACCEPTABLE_SCORE,
        help=f"Minimum acceptable score threshold (default: {MIN_ACCEPTABLE_SCORE})",
    )
    parser.add_argument(
        "--output",
        default=os.path.join(data_dir, "match_results_r3.csv"),
        help="Path to output CSV file",
    )
    parser.add_argument(
        "--prev-results",
        default=",".join([
            os.path.join(data_dir, "match_results.csv"),
            os.path.join(data_dir, "match_results_r2.csv"),
        ]),
        help="Comma-separated paths to previous rounds' match_results CSVs",
    )
    parser.add_argument(
        "--prev-data",
        default=",".join([
            os.path.join(data_dir, "export_thu-date_v0.3.csv"),
            os.path.join(data_dir, "export_thu-date_v0.5.csv"),
        ]),
        help="Comma-separated paths to previous rounds' questionnaire CSVs (for boost calculation)",
    )
    parser.add_argument(
        "--priority-emails",
        default="qinyh22@mails.tsinghua.edu.cn",
        help="Comma-separated list of emails that must be matched if possible",
    )
    args = parser.parse_args()

    priority_emails = set(e.strip() for e in args.priority_emails.split(",") if e.strip())
    prev_results_files = [f.strip() for f in args.prev_results.split(",") if f.strip()]
    prev_data_files = [f.strip() for f in args.prev_data.split(",") if f.strip()]

    # Step 1: Load and filter
    print("Step 1: Loading data...")
    all_users = load_users(args.csv)
    valid_users = filter_valid_users(all_users)
    print(f"  Total users: {len(all_users)}, Valid (completed + participating): {len(valid_users)}")

    print("\n  Loading previous match results (blocked pairs)...")
    prev_pairs = load_previously_matched_pairs(*prev_results_files)
    print(f"  Previously matched pairs: {len(prev_pairs)} (from {len(prev_results_files)} file(s))")
    set_blocked_pairs(prev_pairs)

    print("\n  Computing score boosts for previously unmatched users...")
    boosts = _compute_score_boosts(valid_users, prev_results_files, prev_data_files)
    boost_1x = sum(1 for v in boosts.values() if v == 1.05)
    boost_2x = sum(1 for v in boosts.values() if v == 1.10)
    print(f"  Boosted users: {len(boosts)} total ({boost_1x} at 1.05x, {boost_2x} at 1.10x)")
    set_score_boosts(boosts)

    if priority_emails:
        found = [u for u in valid_users if u.email in priority_emails]
        not_found = priority_emails - {u.email for u in valid_users}
        print(f"\n  Priority users found: {len(found)}")
        for u in found:
            print(f"    #{u.user_id} {u.nickname} ({u.email})")
        if not_found:
            print(f"  Priority emails NOT found in data: {not_found}")

    users_map = {u.user_id: u for u in valid_users}

    # Step 2: Pool assignment (all valid users enter pools)
    print("\nStep 2: Assigning pools...")
    pools = assign_pools(valid_users)
    print(f"  MF pool — Males: {len(pools['mf_male'])}, Females: {len(pools['mf_female'])}")
    print(f"  MM pool: {len(pools['mm'])}")
    print(f"  FF pool: {len(pools['ff'])}")

    all_matches = []
    priority_user_objs = [u for u in valid_users if u.email in priority_emails]

    # Step 3: Match MF pool — priority users FIRST, then Gale-Shapley on the rest
    print("\nStep 3: Matching MF pool...")
    mf_matches = []
    mf_pre_matched_ids = set()

    mf_priority_male = [u for u in priority_user_objs if u in pools["mf_male"]]
    mf_priority_female = [u for u in priority_user_objs if u in pools["mf_female"]]

    if mf_priority_male or mf_priority_female:
        print("  Step 3a: Pre-matching priority users...")
        if mf_priority_male:
            pm, pm_ids = _pre_match_priority(mf_priority_male, pools["mf_female"], users_map)
            mf_matches.extend(pm)
            mf_pre_matched_ids |= pm_ids
        if mf_priority_female:
            pm, pm_ids = _pre_match_priority(mf_priority_female, pools["mf_male"], users_map)
            mf_matches.extend(pm)
            mf_pre_matched_ids |= pm_ids

    remaining_male = [u for u in pools["mf_male"] if u.user_id not in mf_pre_matched_ids]
    remaining_female = [u for u in pools["mf_female"] if u.user_id not in mf_pre_matched_ids]
    print(f"  Step 3b: Gale-Shapley on remaining {len(remaining_male)} males, {len(remaining_female)} females...")
    gs_matches = match_mf_pool(remaining_male, remaining_female, args.min_score)
    mf_matches.extend(gs_matches)

    mf_matched_ids = set(mf_pre_matched_ids)
    for a, b, _, _ in gs_matches:
        mf_matched_ids.add(a)
        mf_matched_ids.add(b)

    all_matches.append(("MF", mf_matches))
    print_report("Male-Female", mf_matches, users_map)

    # Unmatched in MF
    mf_all = pools["mf_male"] + pools["mf_female"]
    mf_prefs = build_preference_lists(pools["mf_male"], pools["mf_female"], args.min_score)
    mf_unmatched = find_unmatched(mf_all, mf_matched_ids, users_map, mf_prefs)
    if mf_unmatched:
        print(f"\n  Unmatched in MF pool: {len(mf_unmatched)}")
        for info in mf_unmatched:
            reason = info["reason"]
            top = info.get("top3_candidates", [])
            top_str = ", ".join(f"{c['nickname']}({c['score']:.3f})" for c in top) if top else "none"
            print(f"    #{info['user_id']} {info['nickname']:<14} reason={reason}  top3=[{top_str}]")

    # Step 4: Match MM pool
    if len(pools["mm"]) >= 2:
        print("\nStep 4: Matching MM pool...")
        mm_matches = []
        mm_pre_matched_ids = set()

        mm_priority = [u for u in priority_user_objs if u in pools["mm"]]
        if mm_priority:
            print("  Step 4a: Pre-matching priority users...")
            pm, pm_ids = _pre_match_priority(mm_priority, pools["mm"], users_map)
            mm_matches.extend(pm)
            mm_pre_matched_ids |= pm_ids

        remaining_mm = [u for u in pools["mm"] if u.user_id not in mm_pre_matched_ids]
        if len(remaining_mm) >= 2:
            print(f"  Step 4b: Irving SRI on remaining {len(remaining_mm)} users...")
            sri_matches = match_same_sex_pool(remaining_mm, "MM", args.min_score)
            mm_matches.extend(sri_matches)

        mm_matched_ids = set(mm_pre_matched_ids)
        for a, b, _, _ in mm_matches:
            mm_matched_ids.add(a)
            mm_matched_ids.add(b)

        all_matches.append(("MM", mm_matches))
        print_report("Male-Male", mm_matches, users_map)

        mm_prefs = build_preference_lists(pools["mm"], None, args.min_score)
        mm_unmatched = find_unmatched(pools["mm"], mm_matched_ids, users_map, mm_prefs)
        if mm_unmatched:
            print(f"\n  Unmatched in MM pool: {len(mm_unmatched)}")
            for info in mm_unmatched:
                reason = info["reason"]
                top = info.get("top3_candidates", [])
                top_str = ", ".join(f"{c['nickname']}({c['score']:.3f})" for c in top) if top else "none"
                print(f"    #{info['user_id']} {info['nickname']:<14} reason={reason}  top3=[{top_str}]")
    else:
        print("\nStep 4: MM pool too small, skipping.")

    # Step 5: Match FF pool
    if len(pools["ff"]) >= 2:
        print("\nStep 5: Matching FF pool...")
        ff_matches = []
        ff_pre_matched_ids = set()

        ff_priority = [u for u in priority_user_objs if u in pools["ff"]]
        if ff_priority:
            print("  Step 5a: Pre-matching priority users...")
            pm, pm_ids = _pre_match_priority(ff_priority, pools["ff"], users_map)
            ff_matches.extend(pm)
            ff_pre_matched_ids |= pm_ids

        remaining_ff = [u for u in pools["ff"] if u.user_id not in ff_pre_matched_ids]
        if len(remaining_ff) >= 2:
            print(f"  Step 5b: Irving SRI on remaining {len(remaining_ff)} users...")
            sri_matches = match_same_sex_pool(remaining_ff, "FF", args.min_score)
            ff_matches.extend(sri_matches)

        ff_matched_ids = set(ff_pre_matched_ids)
        for a, b, _, _ in ff_matches:
            ff_matched_ids.add(a)
            ff_matched_ids.add(b)

        all_matches.append(("FF", ff_matches))
        print_report("Female-Female", ff_matches, users_map)

        ff_prefs = build_preference_lists(pools["ff"], None, args.min_score)
        ff_unmatched = find_unmatched(pools["ff"], ff_matched_ids, users_map, ff_prefs)
        if ff_unmatched:
            print(f"\n  Unmatched in FF pool: {len(ff_unmatched)}")
            for info in ff_unmatched:
                reason = info["reason"]
                top = info.get("top3_candidates", [])
                top_str = ", ".join(f"{c['nickname']}({c['score']:.3f})" for c in top) if top else "none"
                print(f"    #{info['user_id']} {info['nickname']:<14} reason={reason}  top3=[{top_str}]")
    else:
        print("\nStep 5: FF pool too small, skipping.")

    # Step 6: Export
    print(f"\nStep 6: Exporting results...")
    export_results_csv(all_matches, users_map, args.output)

    # Summary
    total_matched = sum(len(m) * 2 for _, m in all_matches)
    print(f"\n{'='*70}")
    print(f"  SUMMARY")
    print(f"{'='*70}")
    print(f"  Valid users:    {len(valid_users)}")
    print(f"  Matched users:  {total_matched}")
    print(f"  Match rate:     {total_matched / len(valid_users) * 100:.1f}%" if valid_users else "  N/A")
    print(f"  Results saved:  {args.output}")

    # Priority check
    if priority_emails:
        all_matched_ids = set()
        for _, matches in all_matches:
            for a, b, _, _ in matches:
                all_matched_ids.add(a)
                all_matched_ids.add(b)
        print(f"\n  Priority user status:")
        for u in priority_user_objs:
            status = "MATCHED" if u.user_id in all_matched_ids else "UNMATCHED"
            print(f"    #{u.user_id} {u.nickname} ({u.email}): {status}")


if __name__ == "__main__":
    main()
