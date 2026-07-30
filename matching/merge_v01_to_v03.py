#!/usr/bin/env python3
"""
Merge v0.1 questionnaire answers into v0.3+ export for users
who completed v0.1 but show as incomplete (answered_count=47) in v0.3+.

Mergeable questions: Q6, Q11, Q32
Non-mergeable: Q23 (type changed), Q47 (scale semantics changed)
"""

import csv
import math
from pathlib import Path

DATA_DIR = Path(__file__).parent

Q6_MAP = {
    "校本部": "清华大学 · 校本部",
    "五道口金融学院": "清华大学 · 五道口金融学院",
    "深圳国际研究生院": "深圳国际研究生院",
    "其他": "其他",
}

Q32_MAP = {
    "健身 / 跑步": "健身/跑步",
    "电影 / 剧集": "电影/剧集",
    "音乐 / 演出": "音乐/演出",
    "户外 / 徒步": "户外/徒步",
    "摄影 / 视频": "摄影/视频",
    "展览 / 博物馆": "展览/博物馆",
    "做饭 / 探店": "做饭/探店",
    "编程 / 科技": "编程/科技",
    "创作 / 写作": "创作/写作（比如诗词、小说）",
    "学术讨论": "学术讨论/做科研",
    "球类运动": "球类运动",
    "游戏": "游戏",
    "旅行": "旅行",
    "阅读": "阅读",
}

TOTAL_QUESTIONS = 52


def map_q32(raw: str) -> str:
    items = [s.strip() for s in raw.split(";") if s.strip()]
    mapped = []
    for item in items:
        new_val = Q32_MAP.get(item)
        if new_val is None:
            print(f"  [WARN] Q32 unknown option: '{item}', keeping as-is")
            mapped.append(item)
        else:
            mapped.append(new_val)
    return ";".join(mapped)


def main():
    v01_path = DATA_DIR / "data" / "export_thu-date_v0.1.csv"
    v03_path = DATA_DIR / "data" / "export_thu-date_v0.3.csv"
    out_path = DATA_DIR / "data" / "export_merged.csv"

    # Load v0.1 data keyed by email
    v01_by_email = {}
    with open(v01_path, "r", encoding="utf-8-sig") as f:
        reader = csv.DictReader(f)
        for row in reader:
            email = row.get("email", "").strip()
            if email:
                v01_by_email[email] = row

    print(f"Loaded {len(v01_by_email)} users from v0.1 export")

    # Process v0.3+ export
    merged_count = 0
    not_found_count = 0
    rows_out = []

    with open(v03_path, "r", encoding="utf-8-sig") as f:
        reader = csv.DictReader(f)
        fieldnames = reader.fieldnames
        for row in reader:
            completed = row.get("completed", "")
            answered = row.get("answered_count", "")
            email = row.get("email", "").strip()

            if completed == "false" and answered == "47" and email:
                v01_row = v01_by_email.get(email)
                if v01_row is None:
                    not_found_count += 1
                else:
                    filled = 0

                    # Q6
                    v01_q6 = v01_row.get("6_value", "").strip()
                    if v01_q6 and not row.get("6_value", "").strip():
                        mapped = Q6_MAP.get(v01_q6)
                        if mapped:
                            row["6_value"] = mapped
                            row["6_importance"] = v01_row.get("6_importance", "false")
                            filled += 1
                        else:
                            print(f"  [WARN] Q6 unknown: '{v01_q6}' for {email}")

                    # Q11
                    v01_q11 = v01_row.get("11_value", "").strip()
                    if v01_q11 and not row.get("11_value", "").strip():
                        row["11_value"] = v01_q11
                        row["11_importance"] = v01_row.get("11_importance", "false")
                        filled += 1

                    # Q32
                    v01_q32 = v01_row.get("32_value", "").strip()
                    if v01_q32 and not row.get("32_value", "").strip():
                        row["32_value"] = map_q32(v01_q32)
                        row["32_importance"] = v01_row.get("32_importance", "false")
                        filled += 1

                    if filled > 0:
                        new_count = int(answered) + filled
                        row["answered_count"] = str(new_count)
                        row["progress_percent"] = str(
                            math.floor(new_count / TOTAL_QUESTIONS * 100)
                        )
                        row["completed"] = "true"
                        if v01_row.get("participating") == "true":
                            row["participating"] = "true"
                        merged_count += 1

            rows_out.append(row)

    with open(out_path, "w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(rows_out)

    print(f"\nMerge complete:")
    print(f"  Merged: {merged_count} users (Q6 + Q11 + Q32 filled from v0.1)")
    print(f"  Not found in v0.1: {not_found_count}")
    print(f"  Output: {out_path}")


if __name__ == "__main__":
    main()
