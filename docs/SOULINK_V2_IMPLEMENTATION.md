# SouLink V2 Implementation

Version: v0.1  
Date: 2026-07-30  
Reference: [SouLink V2 Web Design Requirements](SOULINK_V2_WEB_DESIGN_REQUIREMENTS.md)

## Release Shape

This repository implements the first deployable V2 slice:

> Tsinghua-only intent matching for campus partners, serious dating, and project or CTO candidates.

The primary loop is:

```text
Tsinghua identity
  -> matching profile
  -> natural-language intent
  -> structured review
  -> three recommendations
  -> mutual interest
  -> match room
  -> campus meeting
  -> feedback or safety action
```

The authenticated home page is the working product surface. It starts with “今天，你想认识怎样的人？” and combines intent shortcuts, current campus opportunities, recommendations, active intents, and scheduled meetings.

## MVP Coverage

| Requirement | Implementation | Status |
| --- | --- | --- |
| Tsinghua-only registration | Registration accepts Tsinghua domains only; legacy school URLs redirect to the V2 landing page | Complete |
| Profile and privacy | Skills, interests, availability, scenes, collaboration style, boundaries, AI portrait approval, romantic opt-in, matching pause, account controls | Complete |
| AI intent composer | Natural-language interpretation, inferred fields, confidence, safety notice, up to three clarifying questions, editable structured preview | Complete |
| Intent visibility | Private search, recommended candidates only, or opportunity Feed | Complete |
| Calm opportunity Feed | Finite intent cards with category filters, save, interest, privacy-safe author preview | Complete |
| Three recommendations | Ranked candidates with fit reason, uncertainty, shared context, and profile-aware signals | Complete |
| Mutual interest | Express, withdraw, accept, and decline interest with explicit state transitions | Complete |
| Match room | Fit explanation, icebreakers, campus places, meeting time, status, two-party contact controls | Complete |
| Messaging | Lightweight room messages after mutual consent | Complete |
| Match feedback | Context-linked feedback form accessible from a match room | Complete |
| Safety | Report intent or room, optionally block a user, close unsafe rooms, hide blocked users, delete account | Complete |
| Admin report review | Admin queue, filtering, notes, assignment, resolve and dismiss actions | Complete |
| Project and CTO matching | Structured project memo plus required-skill, availability, and collaboration-style ranking | Complete |
| Romantic matching | Explicit profile opt-in, candidate-only visibility, no public identifiable romantic cards | Complete |
| External LLM integration | Deterministic interpreter and ranking run without a model key; provider-backed V2 interpretation is not connected | Deferred |
| Category configuration and matching analytics | Report moderation exists; category management and quality dashboards are not part of this release | Deferred |
| Group matching and native mobile | Deliberately excluded from the MVP | Deferred |

## Product Decisions

The requirements document left several questions open. The implementation takes these initial positions:

1. Romantic matching is one intent category with stricter privacy, explicit opt-in, and low recommendation volume.
2. Anonymous public intent cards are not supported. Feed cards show only a limited verified preview.
3. Project and CTO intents expose a project memo in the composer. Recommendations work with incomplete memos but clearly surface uncertainty.
4. The first release supports one-to-one matching only.
5. Before mutual consent, a card may show nickname, department, grade, and verification status. Full profile and contact details remain restricted.
6. General campus intents are real-time. Romantic intents are candidate-only and behave like a curated private queue.
7. Match explanations are visible as decision support and never claim certainty.
8. Local safety assistance is conservative and rule-based; reports always support human review.
9. Contact details are revealed inside the match room only after each user opts in.
10. The launch scene is intentionally not hard-coded. The demo emphasizes dinner, study, running, and founder matching so operators can test which loop earns repeat use.

## Main Web Routes

| Route | Purpose |
| --- | --- |
| `/` | Tsinghua V2 landing page |
| `/login`, `/register` | Authentication |
| `/safety` | Public privacy and safety explanation |
| `/home` | Intent-first command center |
| `/discover` | Campus opportunity Feed |
| `/intents/new` | Guided intent composer |
| `/intents` | User intent lifecycle |
| `/intents/:intentId` | Intent, recommendations, and applicants |
| `/matches` | Mutual matches |
| `/matches/:roomId` | Focused match room |
| `/profile` | Matching profile and AI portrait |
| `/account` | Account and contact settings |
| `/feedback` | Match feedback and reporting |
| `/admin/safety` | Moderation queue for configured admins |

## V2 Data Model

Migration `000016_v2_intent_matching.sql` adds:

- `matching_profiles`;
- `connection_intents`;
- `intent_interests`;
- `intent_saves`;
- `match_rooms`;
- `match_room_messages`.

Migration `000017_safety_moderation.sql` adds:

- `user_blocks`;
- `safety_reports`.

Migration `000018_project_memos.sql` adds a validated JSON object to each intent for:

- product and stage;
- weekly commitment;
- compensation expectations;
- required and optional skills;
- working style;
- risk appetite.

## AI Boundary

The current V2 AI-shaped behavior is deterministic and testable:

- intent text is classified and normalized into explicit fields;
- only material missing information becomes a clarification question;
- recommendation scores use category, time, location, constraints, profile skills, availability, and project memo evidence;
- every recommendation returns a fit reason and an uncertainty;
- match rooms receive explainable shared context, conversation starters, and campus meeting suggestions.

This is the fallback contract future model providers must preserve. A model may improve language understanding, but it must not bypass visibility, consent, romantic opt-in, blocking, or safety rules.

## Local Acceptance

A clean local startup must satisfy:

1. `docker compose up -d --build` reaches healthy status.
2. Database migrations run through `000018`.
3. The demo account can sign in from the browser.
4. Demo data contains five users, six intents, one mutual room, and two messages.
5. The demo account sees four public opportunities, one owned CTO intent, and three recommendations.
6. Repeated API startup creates no duplicate demo users, intents, or rooms.
7. Web production build, ESLint, Go tests, and `go vet` pass.

## Deliberate Legacy Compatibility

The repository retains the V1 questionnaire, scheduled matching, friendship, chat, and Python matching modules so historical data and operational tools remain migratable. V2 does not expose those workflows in its primary navigation. New account creation and all visible V2 entry points are Tsinghua-only.
