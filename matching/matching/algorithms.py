"""
Matching algorithms:
  - Gale-Shapley with incomplete preference lists (for male-female pool)
  - Irving's Stable Roommates with incomplete lists (for same-sex pools)
  - Maximum Weight Matching fallback (via networkx)
"""
from __future__ import annotations
import networkx as nx
from .models import User
from .filters import pass_all_filters
from .scoring import compute_score
from .config import MIN_ACCEPTABLE_SCORE


# ---------------------------------------------------------------------------
# Preference list builder
# ---------------------------------------------------------------------------

def build_preference_lists(
    users_a: list[User],
    users_b: list[User] | None = None,
    min_score: float = MIN_ACCEPTABLE_SCORE,
) -> dict[int, list[tuple[int, float]]]:
    """
    Build quality-gated preference lists.
    If users_b is None, builds within a single pool (same-sex).
    Returns {user_id: [(candidate_id, score), ...]} sorted descending.
    """
    is_bipartite = users_b is not None
    prefs: dict[int, list[tuple[int, float]]] = {}

    if is_bipartite:
        for a in users_a:
            ranked = []
            for b in users_b:
                if not pass_all_filters(a, b):
                    continue
                s = compute_score(a, b)
                if s >= min_score:
                    ranked.append((b.user_id, s))
            ranked.sort(key=lambda x: -x[1])
            prefs[a.user_id] = ranked

        for b in users_b:
            ranked = []
            for a in users_a:
                if not pass_all_filters(b, a):
                    continue
                s = compute_score(b, a)
                if s >= min_score:
                    ranked.append((a.user_id, s))
            ranked.sort(key=lambda x: -x[1])
            prefs[b.user_id] = ranked
    else:
        all_users = users_a
        for a in all_users:
            ranked = []
            for b in all_users:
                if a.user_id == b.user_id:
                    continue
                if not pass_all_filters(a, b):
                    continue
                s = compute_score(a, b)
                if s >= min_score:
                    ranked.append((b.user_id, s))
            ranked.sort(key=lambda x: -x[1])
            prefs[a.user_id] = ranked

    return prefs


# ---------------------------------------------------------------------------
# Gale-Shapley with incomplete lists
# ---------------------------------------------------------------------------

def gale_shapley(
    proposers: list[User],
    acceptors: list[User],
    prefs: dict[int, list[tuple[int, float]]],
) -> list[tuple[int, int, float, float]]:
    """
    Gale-Shapley with incomplete preference lists.
    Returns list of (proposer_id, acceptor_id, proposer_score, acceptor_score).
    """
    proposer_ids = [u.user_id for u in proposers]
    acceptor_ids = set(u.user_id for u in acceptors)

    # proposer state: index into their pref list
    prop_next = {pid: 0 for pid in proposer_ids}
    prop_prefs = {pid: prefs.get(pid, []) for pid in proposer_ids}

    # acceptor ranking lookup: acceptor_id → {proposer_id: rank}
    acc_ranking: dict[int, dict[int, int]] = {}
    for aid in acceptor_ids:
        plist = prefs.get(aid, [])
        acc_ranking[aid] = {cid: rank for rank, (cid, _) in enumerate(plist)}

    # current engagements: acceptor_id → proposer_id
    engaged: dict[int, int] = {}
    free_proposers = list(proposer_ids)

    while free_proposers:
        pid = free_proposers.pop(0)
        plist = prop_prefs[pid]
        idx = prop_next[pid]

        matched = False
        while idx < len(plist):
            target_id, _ = plist[idx]
            idx += 1
            prop_next[pid] = idx

            if target_id not in acceptor_ids:
                continue

            target_ranks = acc_ranking.get(target_id, {})
            if pid not in target_ranks:
                continue

            current_holder = engaged.get(target_id)
            if current_holder is None:
                engaged[target_id] = pid
                matched = True
                break
            else:
                if target_ranks[pid] < target_ranks.get(current_holder, float("inf")):
                    engaged[target_id] = pid
                    free_proposers.append(current_holder)
                    matched = True
                    break

        # if idx exhausted → pid stays unmatched (intended)

    # Build result with scores
    score_lookup = {}
    for uid, plist in prefs.items():
        for cid, s in plist:
            score_lookup[(uid, cid)] = s

    results = []
    for aid, pid in engaged.items():
        ps = score_lookup.get((pid, aid), 0.0)
        as_ = score_lookup.get((aid, pid), 0.0)
        results.append((pid, aid, ps, as_))

    return results


def gale_shapley_bidirectional(
    males: list[User],
    females: list[User],
    prefs: dict[int, list[tuple[int, float]]],
) -> list[tuple[int, int, float, float]]:
    """
    Run GS twice (male-propose and female-propose), return the higher quality result.
    """
    result_mp = gale_shapley(males, females, prefs)
    result_fp = gale_shapley(females, males, prefs)

    def avg_quality(results):
        if not results:
            return 0.0
        return sum(min(ps, as_) for _, _, ps, as_ in results) / len(results)

    q_mp = avg_quality(result_mp)
    q_fp = avg_quality(result_fp)

    return result_mp if q_mp >= q_fp else result_fp


# ---------------------------------------------------------------------------
# Irving's Stable Roommates (SRI) with incomplete lists
# ---------------------------------------------------------------------------

def _irving_phase1(prefs_work: dict[int, list[int]]) -> bool:
    """
    Phase 1: proposal round. Returns False if someone's list becomes empty
    while still unmatched (no stable matching via this phase).
    """
    persons = list(prefs_work.keys())
    proposals: dict[int, int | None] = {p: None for p in persons}
    free = list(persons)

    prop_idx = {p: 0 for p in persons}
    rank_of: dict[int, dict[int, int]] = {}
    for p in persons:
        rank_of[p] = {q: i for i, q in enumerate(prefs_work[p])}

    while free:
        p = free.pop(0)
        plist = prefs_work[p]

        matched = False
        while prop_idx[p] < len(plist):
            q = plist[prop_idx[p]]
            prop_idx[p] += 1

            if q not in prefs_work:
                continue

            current = proposals[q]
            if current is None:
                proposals[q] = p
                matched = True
                # Remove everyone after p in q's list
                r = rank_of[q]
                if p in r:
                    cutoff = r[p]
                    to_remove = prefs_work[q][cutoff + 1:]
                    prefs_work[q] = prefs_work[q][: cutoff + 1]
                    for removed in to_remove:
                        if removed in prefs_work and q in prefs_work.get(removed, []):
                            prefs_work[removed] = [
                                x for x in prefs_work[removed] if x != q
                            ]
                break
            else:
                r = rank_of[q]
                if p in r and (current not in r or r[p] < r[current]):
                    proposals[q] = p
                    # Remove current and everyone after p in q's list
                    cutoff = r[p]
                    to_remove = prefs_work[q][cutoff + 1:]
                    prefs_work[q] = prefs_work[q][: cutoff + 1]
                    for removed in to_remove:
                        if removed in prefs_work and q in prefs_work.get(removed, []):
                            prefs_work[removed] = [
                                x for x in prefs_work[removed] if x != q
                            ]
                    free.append(current)
                    matched = True
                    break

        if not matched:
            prefs_work.pop(p, None)
            for other in list(prefs_work.keys()):
                prefs_work[other] = [x for x in prefs_work[other] if x != p]

    return True


def _irving_phase2(prefs_work: dict[int, list[int]]) -> dict[int, int] | None:
    """
    Phase 2: rotation elimination. Returns matching dict or None if no stable matching.
    """
    # Check if already solved
    all_single = all(len(v) <= 1 for v in prefs_work.values())
    if all_single:
        matching = {}
        for p, lst in prefs_work.items():
            if lst:
                matching[p] = lst[0]
        return matching

    max_iterations = len(prefs_work) * len(prefs_work) + 100
    iteration = 0

    while True:
        iteration += 1
        if iteration > max_iterations:
            return None

        # Find person with list length > 1
        pivot = None
        for p, lst in prefs_work.items():
            if len(lst) > 1:
                pivot = p
                break

        if pivot is None:
            break

        # Find rotation
        rotation = []
        p = pivot
        visited = set()
        found_rotation = False

        for _ in range(len(prefs_work) + 1):
            if p in visited:
                # Extract the cycle
                cycle_start = None
                for i, (a, _) in enumerate(rotation):
                    if a == p:
                        cycle_start = i
                        break
                if cycle_start is not None:
                    rotation = rotation[cycle_start:]
                    found_rotation = True
                break
            visited.add(p)

            if len(prefs_work.get(p, [])) < 2:
                break

            q = prefs_work[p][1]  # second on p's list
            rotation.append((p, q))

            # Find last on q's list
            q_list = prefs_work.get(q, [])
            if not q_list:
                break
            p = q_list[-1]

        if not found_rotation or len(rotation) < 2:
            # Cannot find rotation → no stable matching
            return None

        # Eliminate rotation: for each (a_i, b_i) in rotation,
        # reject the pair (b_i, a_{i+1})
        for i in range(len(rotation)):
            a_i = rotation[i][0]
            b_i = rotation[i][1]
            a_next = rotation[(i + 1) % len(rotation)][0]

            # Remove a_next from b_i's list and b_i from a_next's list
            if b_i in prefs_work:
                prefs_work[b_i] = [x for x in prefs_work[b_i] if x != a_next]
            if a_next in prefs_work:
                prefs_work[a_next] = [x for x in prefs_work[a_next] if x != b_i]

        # Remove anyone with empty list
        empty = [p for p, lst in prefs_work.items() if not lst]
        for p in empty:
            prefs_work.pop(p, None)
            for other in prefs_work:
                prefs_work[other] = [x for x in prefs_work[other] if x != p]

        # Check if all lists have length <= 1
        if all(len(v) <= 1 for v in prefs_work.values()):
            break

    matching = {}
    for p, lst in prefs_work.items():
        if lst:
            matching[p] = lst[0]
    return matching


def irving_stable_roommates(
    users: list[User],
    prefs: dict[int, list[tuple[int, float]]],
) -> list[tuple[int, int, float, float]] | None:
    """
    Irving's SRI. Returns matched pairs or None if no stable matching exists.
    """
    if len(users) < 2:
        return []

    score_lookup = {}
    for uid, plist in prefs.items():
        for cid, s in plist:
            score_lookup[(uid, cid)] = s

    # Build working pref lists (id-only, ordered)
    prefs_work: dict[int, list[int]] = {}
    for u in users:
        plist = prefs.get(u.user_id, [])
        if plist:
            prefs_work[u.user_id] = [cid for cid, _ in plist]

    if not prefs_work:
        return []

    _irving_phase1(prefs_work)
    matching = _irving_phase2(prefs_work)

    if matching is None:
        return None

    # Deduplicate pairs
    seen = set()
    results = []
    for a, b in matching.items():
        pair = (min(a, b), max(a, b))
        if pair not in seen:
            seen.add(pair)
            sa = score_lookup.get((a, b), 0.0)
            sb = score_lookup.get((b, a), 0.0)
            results.append((a, b, sa, sb))

    return results


# ---------------------------------------------------------------------------
# Maximum Weight Matching fallback (networkx)
# ---------------------------------------------------------------------------

def max_weight_matching(
    users: list[User],
    prefs: dict[int, list[tuple[int, float]]],
) -> list[tuple[int, int, float, float]]:
    """
    Fallback: maximum weight matching on a general graph.
    Edge weight = min(score(A→B), score(B→A)) to ensure mutual quality.
    """
    score_lookup = {}
    for uid, plist in prefs.items():
        for cid, s in plist:
            score_lookup[(uid, cid)] = s

    G = nx.Graph()
    user_ids = [u.user_id for u in users]
    G.add_nodes_from(user_ids)

    for i, a in enumerate(users):
        for b in users[i + 1:]:
            sa = score_lookup.get((a.user_id, b.user_id))
            sb = score_lookup.get((b.user_id, a.user_id))
            if sa is not None and sb is not None:
                w = min(sa, sb)
                if w >= MIN_ACCEPTABLE_SCORE:
                    G.add_edge(a.user_id, b.user_id, weight=w)

    raw_matching = nx.max_weight_matching(G, maxcardinality=False)

    results = []
    for a_id, b_id in raw_matching:
        sa = score_lookup.get((a_id, b_id), 0.0)
        sb = score_lookup.get((b_id, a_id), 0.0)
        results.append((a_id, b_id, sa, sb))

    return results


# ---------------------------------------------------------------------------
# High-level pool matchers
# ---------------------------------------------------------------------------

def match_mf_pool(
    males: list[User],
    females: list[User],
    min_score: float = MIN_ACCEPTABLE_SCORE,
) -> list[tuple[int, int, float, float]]:
    """Match male-female pool using bidirectional Gale-Shapley."""
    print(f"  [MF Pool] Males: {len(males)}, Females: {len(females)}")
    prefs = build_preference_lists(males, females, min_score)
    results = gale_shapley_bidirectional(males, females, prefs)
    print(f"  [MF Pool] Matched pairs: {len(results)}")
    return results


def match_same_sex_pool(
    users: list[User],
    pool_name: str,
    min_score: float = MIN_ACCEPTABLE_SCORE,
) -> list[tuple[int, int, float, float]]:
    """Match same-sex pool using Irving SRI with MWM fallback."""
    print(f"  [{pool_name}] Users: {len(users)}")

    if len(users) < 2:
        print(f"  [{pool_name}] Too few users, skipping.")
        return []

    prefs = build_preference_lists(users, None, min_score)

    # Try Irving first
    result = irving_stable_roommates(users, prefs)
    if result is not None:
        print(f"  [{pool_name}] Irving SRI succeeded. Matched pairs: {len(result)}")
        return result

    # Fallback to MWM
    print(f"  [{pool_name}] Irving SRI failed, falling back to Max Weight Matching.")
    result = max_weight_matching(users, prefs)
    print(f"  [{pool_name}] MWM matched pairs: {len(result)}")
    return result
