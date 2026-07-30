from __future__ import annotations

import argparse
import json
from pathlib import Path
import sys

from .pipeline import build_report, run_batch, run_single
from .simulator import save_simulated_responses


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="SoulAgent MVP CLI")
    sub = parser.add_subparsers(dest="command", required=True)

    simulate = sub.add_parser("simulate", help="Generate simulated questionnaire responses")
    simulate.add_argument("--count", type=int, default=10)

    batch = sub.add_parser("batch", help="Generate reports for all simulated responses")
    batch.add_argument("--limit", type=int, default=None)
    batch.add_argument("--skill", default="tsinghua_soulink")

    single = sub.add_parser("single", help="Generate one report by respondent id")
    single.add_argument("respondent_id")
    single.add_argument("--skill", default="tsinghua_soulink")

    portrait = sub.add_parser("portrait", help="Generate one portrait from questionnaire JSON")
    portrait.add_argument("--input", default="-", help="JSON file path, or '-' to read from stdin")
    portrait.add_argument("--skill", default="tsinghua_soulink")

    return parser


def main() -> None:
    parser = build_parser()
    args = parser.parse_args()
    project_root = Path(__file__).resolve().parent.parent

    if args.command == "simulate":
        path = project_root / "data" / "simulated_responses.json"
        save_simulated_responses(path, count=args.count)
        print(f"saved simulated responses -> {path}")
        return

    if args.command == "batch":
        def progress(index: int, total: int, respondent_id: str) -> None:
            print(f"[{index}/{total}] processing {respondent_id}...")

        outputs = run_batch(
            project_root,
            limit=args.limit,
            progress_callback=progress,
            skill_id=args.skill,
        )
        print(f"generated {len(outputs)} reports")
        for path in outputs:
            print(path)
        return

    if args.command == "single":
        output = run_single(project_root, args.respondent_id, skill_id=args.skill)
        print(output)
        return

    if args.command == "portrait":
        if args.input == "-":
            payload = json.load(sys.stdin)
        else:
            payload = json.loads(Path(args.input).read_text(encoding="utf-8"))

        traits, report = build_report(payload, skill_id=args.skill)
        print(
            json.dumps(
                {
                    "traits": traits,
                    "report": report,
                },
                ensure_ascii=False,
            )
        )
        return


if __name__ == "__main__":
    main()
