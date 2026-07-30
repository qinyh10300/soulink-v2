import csv
import json
from .models import User
from .config import LIKERT_QUESTIONS, MULTI_SELECT_QUESTIONS, NUMERIC_QUESTIONS, RANGE_QUESTIONS


def _parse_value(q_num: int, raw: str):
    if raw is None or raw.strip() == "":
        return None

    if q_num in RANGE_QUESTIONS:
        try:
            data = json.loads(raw)
            if data.get("unlimited"):
                return {"min": 0, "max": 999}
            return {"min": data.get("min"), "max": data.get("max")}
        except (json.JSONDecodeError, TypeError):
            return None

    if q_num in NUMERIC_QUESTIONS:
        try:
            return int(raw)
        except ValueError:
            return None

    if q_num in LIKERT_QUESTIONS:
        try:
            return int(raw)
        except ValueError:
            return None

    if q_num in MULTI_SELECT_QUESTIONS:
        items = [s.strip() for s in raw.split(";") if s.strip()]
        return set(items) if items else None

    return raw


def load_users(csv_path: str) -> list[User]:
    users = []
    with open(csv_path, "r", encoding="utf-8-sig") as f:
        reader = csv.DictReader(f)
        for row in reader:
            user = User(
                user_id=int(row["user_id"]),
                email=row.get("email", ""),
                nickname=row.get("nickname", ""),
                department=row.get("department", ""),
                grade=row.get("grade", ""),
                completed=row.get("completed", "false") == "true",
                participating=row.get("participating", "false") == "true",
            )
            for q in range(1, 53):
                val_key = f"{q}_value"
                imp_key = f"{q}_importance"
                user.answers[q] = _parse_value(q, row.get(val_key, ""))
                user.importance[q] = row.get(imp_key, "false") == "true"
            users.append(user)
    return users


def filter_valid_users(users: list[User]) -> list[User]:
    return [u for u in users if u.completed and u.participating]


def load_previously_matched_pairs(*results_csvs: str) -> set[frozenset[str]]:
    """Load matched email pairs from one or more previous match_results CSVs.
    Returns a set of frozenset({email_a, email_b}) for each pair."""
    pairs = set()
    for results_csv in results_csvs:
        try:
            with open(results_csv, "r", encoding="utf-8-sig") as f:
                reader = csv.DictReader(f)
                for row in reader:
                    email_a = row.get("user_a_email", "").strip()
                    email_b = row.get("user_b_email", "").strip()
                    if email_a and email_b:
                        pairs.add(frozenset({email_a, email_b}))
        except FileNotFoundError:
            pass
    return pairs


def load_previously_matched_emails(*results_csvs: str) -> set[str]:
    """Load all emails that were successfully matched in previous rounds."""
    emails = set()
    for results_csv in results_csvs:
        try:
            with open(results_csv, "r", encoding="utf-8-sig") as f:
                reader = csv.DictReader(f)
                for row in reader:
                    email_a = row.get("user_a_email", "").strip()
                    email_b = row.get("user_b_email", "").strip()
                    if email_a:
                        emails.add(email_a)
                    if email_b:
                        emails.add(email_b)
        except FileNotFoundError:
            pass
    return emails
