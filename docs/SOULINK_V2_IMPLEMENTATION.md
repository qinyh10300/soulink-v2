# SouLink V2 Implementation

Version: v1.0
Date: 2026-08-07
Reference: [SouLink V2 Web Design Requirements](SOULINK_V2_WEB_DESIGN_REQUIREMENTS.md)

## Release Shape

This repository implements the first deployable V2 slice:

> Tsinghua-only intent matching for campus partners, serious dating, and project or CTO candidates.

The primary matching loop is:

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

The deployable release adds an anonymous Tsinghua plaza, direct and group chat, privacy-aware scheduling, AI group formation, meetup feedback, private chat assistance, trust and safety workflows, campus themes, notifications, and operational observability around that loop.

## MVP Coverage

| Requirement | Implementation | Status |
| --- | --- | --- |
| Tsinghua-only registration | Registration accepts Tsinghua domains only; legacy school URLs redirect to the V2 landing page | Complete |
| Profile and privacy | Skills, interests, availability, scenes, collaboration style, boundaries, AI portrait approval, romantic opt-in, category controls, matching pause, account controls | Complete |
| AI intent composer | Natural-language interpretation, inferred fields, confidence, safety notice, up to three clarifying questions, editable structured preview | Complete |
| Intent visibility | Private search, recommended candidates only, or opportunity Feed | Complete |
| Calm opportunity Feed | Finite intent cards with category filters, saved view, persistent dismiss, interest, privacy-safe author preview | Complete |
| Three recommendations | Ranked candidates with fit reason, uncertainty, shared context, and profile-aware signals | Complete |
| Mutual interest | Express, withdraw, accept, and decline interest with explicit state transitions | Complete |
| Match room | Fit explanation, icebreakers, campus places, meeting time, met/closed lifecycle, two-party contact controls | Complete |
| Messaging | Match-room messages plus persistent one-to-one and group conversations, unread state, history, and member management | Complete |
| Anonymous plaza | Anonymous posts and nested replies, stable per-hole aliases, topics, search, likes, and hot/latest views without author identity in the API | Complete |
| Shared calendar | Weekly calendar, event CRUD, public busy/free views, profile invitations, attendees, and RSVP | Complete |
| Match feedback | Context-linked feedback form accessible from a match room | Complete |
| Safety | Report intent or room, optionally block a user, close unsafe rooms, hide blocked users, delete account | Complete |
| Admin report review | Admin queue, filtering, notes, assignment, resolve and dismiss actions | Complete |
| Project and CTO matching | Structured project memo plus required-skill, availability, and collaboration-style ranking | Complete |
| Romantic matching | Explicit profile opt-in, candidate-only visibility, no public identifiable romantic cards | Complete |
| External LLM integration | OpenAI-compatible structured generation with timeout, retry, validation, safety filtering, and deterministic fallback | Complete |
| Campus themes and analytics | Configurable campus themes plus supply, demand, formation, meetup, funnel, and safety dashboards | Complete |
| AI group matching | Role-complementary candidate ranking, anonymous consent, waitlist, group creation, scheduling, and deterministic fallback | Complete |
| Meetup workflow | Availability, common-time suggestions, proposals, RSVP, calendar projection, completion, and private feedback | Complete |
| Private chat assistant | Openers, summaries, consensus, and editable action drafts isolated per user and context | Complete |
| Trust and safety | Independent trust signals, photo review, risky-message confirmation, trusted-contact sharing, reports, appeals, and audit history | Complete |
| Retention and operations | PWA shell, opt-in notifications, onboarding, campus themes, cursor pagination, JSON logs, request IDs, and admin health metrics | Complete |
| Native mobile clients | Responsive Web and installable PWA are included; native iOS and Android clients are outside this release | Out of scope |

## Product Decisions

The requirements document left several questions open. The implementation takes these initial positions:

1. Romantic matching is one intent category with stricter privacy, explicit opt-in, and low recommendation volume.
2. Anonymous public intent cards are not supported. Feed cards show only a limited verified preview.
3. Project and CTO intents expose a project memo in the composer. Recommendations work with incomplete memos but clearly surface uncertainty.
4. The release supports one-to-one matching, role-complementary group matching, and plaza-originated temporary groups.
5. Before mutual consent, a card may show nickname, department, grade, and verification status. Full profile and contact details remain restricted.
6. General campus intents are real-time. Romantic intents are candidate-only and behave like a curated private queue.
7. Match explanations are visible as decision support and never claim certainty.
8. Local safety assistance is conservative and rule-based; reports always support human review.
9. Contact details are revealed inside the match room only after each user opts in.
10. The launch scene is intentionally not hard-coded. The demo emphasizes dinner, study, running, and founder matching so operators can test which loop earns repeat use.
11. Plaza anonymity is enforced at the API boundary. The server retains author identity for abuse review, while clients receive only “洞主” and stable per-hole participant aliases.
12. Public profile calendars expose occupied time, not private event details. Details are visible only for public events or events in which the viewer participates.
13. Direct messages require an accepted friendship. Groups can be created manually from friends or by an authorized match, meetup, or plaza workflow.

## Main Web Routes

| Route | Purpose |
| --- | --- |
| `/` | Tsinghua V2 landing page |
| `/login`, `/register` | Authentication |
| `/safety` | Public privacy and safety explanation |
| `/connect` | Current action, daily picks, pending meetups, and limited campus invitations |
| `/connect/discover` | Finite campus opportunity Feed |
| `/connect/new` | Guided intent composer |
| `/connect/groups` | Group intent, candidate, invitation, and formation workspace |
| `/connect/intents`, `/connect/intents/:intentId` | Intent lifecycle, recommendations, and applicants |
| `/connect/rooms`, `/connect/rooms/:roomId` | Mutual matches and focused match rooms |
| `/plaza` | Anonymous Tsinghua plaza |
| `/chat`, `/chat/direct/:conversationId`, `/chat/group/:groupId` | Direct and group messaging |
| `/calendar` | Personal week calendar, invitations, and RSVP |
| `/profile` | Matching profile and AI portrait |
| `/profile/:userId` | Privacy-safe public profile and shared availability |
| `/account` | Account and contact settings |
| `/feedback` | Match feedback and reporting |
| `/admin/safety` | Safety review, product metrics, campus theme operations, and runtime health |

The earlier `/home`, `/discover`, `/intents`, and `/matches` addresses redirect to their `/connect` equivalents.

## V2 Data Model

Migration `000016_v2_intent_matching.sql` adds:

- `matching_profiles`;
- `connection_intents`;
- `intent_interests`;
- `intent_saves`;
- `match_rooms`;
- `match_room_messages`.

Migration `000017_safety_reports.sql` adds:

- `user_blocks`;
- `safety_reports`.

Migration `000018_project_memos.sql` adds a validated JSON object to each intent for:

- product and stage;
- weekly commitment;
- compensation expectations;
- required and optional skills;
- working style;
- risk appetite.

Migration `000019_intent_preferences.sql` adds:

- persistent category visibility controls;
- per-user intent dismissals used by Feed and recommendations.

Migration `000020_safety_review_workflow.sql` adds:

- report assignment and reviewer identity;
- durable review notes and update timestamps.

Migration `000021_anonymous_plaza.sql` adds:

- plaza topics and topic feed indexes;
- content search support while retaining server-side author accountability.

Migration `000022_group_chat.sql` adds:

- persistent groups and owner/member roles;
- per-member read cursors;
- paginated group messages.

Migration `000023_calendar.sql` adds:

- privacy-aware calendar events;
- event attendees and pending, accepted, or declined RSVP state;
- organizer and attendee time indexes.

Migrations `000024` through `000031` add:

- trusted server-side product events and aggregate indexes;
- availability, meetup proposals, RSVP, calendar projection, and private feedback;
- AI preferences, recommendation learning, daily picks, and assistant provenance;
- group intents, complementary candidates, invitations, waitlists, and formed groups;
- plaza-originated temporary groups and identity reveal policy;
- private chat assistant artifacts and user controls;
- photo review, message risk decisions, trusted contacts, report evidence, appeals, and audit history;
- notification preferences, subscriptions, delivery state, and campus theme follows.

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
2. Database migrations run through `000031_notifications.sql`.
3. The demo and local Hack administrator accounts can sign in from the browser.
4. A clean demo workspace contains six users, eight intents, one mutual room, three anonymous plaza posts, one group, and the related chat and scheduling fixtures.
5. Save, dismiss, category visibility, and global matching pause affect subsequent Feed and recommendation requests.
6. A match room can move from open to met to closed; closed rooms reject messages and retract contact sharing.
7. The Hack administrator can take a report, persist review notes, resolve or dismiss it, and filter the queue by status.
8. Plaza API responses contain no author user ID, nickname, email, department, or grade; replies retain stable aliases within a hole.
9. Direct and group messages persist, update unread state, and reject access from non-members.
10. A public profile shows busy time without private details and can create an invitation that the recipient accepts or declines.
11. Repeated API startup creates no duplicate demo users, intents, rooms, posts, conversations, groups, messages, events, or administrator accounts.
12. Web production build, ESLint, Go tests, and `go vet` pass.

## Final Completion Matrix

| Area | Delivered result | Status |
| --- | --- | --- |
| Product foundation | Tsinghua-only identity, event dictionary, API conventions, privacy boundaries, and reproducible development plan | Complete |
| Connection loop | Intent creation, explainable recommendations, mutual consent, rooms, messaging, meetup scheduling, and private feedback | Complete |
| AI-native workflows | Optional model provider, deterministic fallback, editable provenance, preference learning, group matching, and chat assistance | Complete |
| Social surfaces | Finite opportunity Feed, anonymous plaza, direct chat, group chat, shared availability, and campus themes | Complete |
| Safety and trust | Independent trust signals, risky-message confirmation, trusted contacts, evidence snapshots, appeals, and audit records | Complete |
| Retention and operations | PWA, opt-in notifications, onboarding, cursor pagination, reconnect recovery, metrics, and structured logs | Complete |
| Delivery quality | Full test/vet/build/lint checks, phone/tablet/desktop browser acceptance, and clean-volume Compose bootstrap | Complete |

## Deliberate Legacy Compatibility

The repository retains the V1 questionnaire, scheduled matching, and Python matching modules so historical data and operational tools remain migratable. Friendship and chat have now been promoted into the primary V2 navigation. New account creation and all visible V2 entry points are Tsinghua-only.
