from __future__ import annotations
from .models import User
from .config import SMOKING_INTENSITY, EMAIL_DOMAIN_TO_UNIVERSITY, UNIVERSITY_ALIASES

_blocked_pairs: set[frozenset[str]] = set()


def set_blocked_pairs(pairs: set[frozenset[str]]):
    """Set the previously matched pairs that cannot be matched again."""
    global _blocked_pairs
    _blocked_pairs = pairs


def pass_previous_pair(a: User, b: User) -> bool:
    """Block pairs that were already matched in a previous round."""
    if not _blocked_pairs:
        return True
    return frozenset({a.email, b.email}) not in _blocked_pairs


def pass_age(a: User, b: User) -> bool:
    """Check bidirectional age range compatibility (Q7 ↔ Q8)."""
    a_age = a.q(7)
    b_age = b.q(7)
    a_range = a.q(8)
    b_range = b.q(8)

    if a_age is None or b_age is None:
        return True  # missing data → don't filter
    if a_range is None or b_range is None:
        return True

    a_in_b = b_range["min"] <= a_age <= b_range["max"]
    b_in_a = a_range["min"] <= b_age <= a_range["max"]
    return a_in_b and b_in_a


def pass_height(a: User, b: User) -> bool:
    """Check bidirectional height range compatibility (Q9 ↔ Q10)."""
    a_h = a.q(9)
    b_h = b.q(9)
    a_range = a.q(10)
    b_range = b.q(10)

    if a_h is None or b_h is None:
        return True
    if a_range is None or b_range is None:
        return True

    a_in_b = b_range["min"] <= a_h <= b_range["max"]
    b_in_a = a_range["min"] <= b_h <= a_range["max"]
    return a_in_b and b_in_a


def pass_smoking(a: User, b: User) -> bool:
    """Hard veto when one smokes frequently and the other can't tolerate it (Q29 ↔ Q30)."""
    a_smoke = SMOKING_INTENSITY.get(a.q(29), 0)
    b_smoke = SMOKING_INTENSITY.get(b.q(29), 0)
    a_tol = a.q(30)
    b_tol = b.q(30)

    if a_tol is None or b_tol is None:
        return True

    if a_smoke >= 3 and b_tol <= 2:
        return False
    if b_smoke >= 3 and a_tol <= 2:
        return False
    return True


def _get_university(user: User) -> str | None:
    """Determine a user's actual university from their email domain."""
    email = user.email or ""
    domain = email.split("@")[-1] if "@" in email else ""
    return EMAIL_DOMAIN_TO_UNIVERSITY.get(domain)


def _normalize_university_set(unis: set | None) -> set | None:
    """Normalize university names using aliases (e.g. '北大大学' -> '北京大学')."""
    if not unis:
        return unis
    return {UNIVERSITY_ALIASES.get(u, u) for u in unis}


def pass_university(a: User, b: User) -> bool:
    """Hard filter: each user's Q6 preferred universities must include the other's actual university.
    '任意' in Q6 means accept any university. Skip if actual university can't be determined."""
    a_uni = _get_university(a)
    b_uni = _get_university(b)

    a_prefs = _normalize_university_set(a.q(6))
    b_prefs = _normalize_university_set(b.q(6))

    if a_prefs and "任意" not in a_prefs and b_uni:
        if b_uni not in a_prefs:
            return False

    if b_prefs and "任意" not in b_prefs and a_uni:
        if a_uni not in b_prefs:
            return False

    return True


def pass_all_filters(a: User, b: User) -> bool:
    return (pass_previous_pair(a, b) and pass_university(a, b)
            and pass_age(a, b) and pass_smoking(a, b))
