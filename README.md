# SouLink V2

> At Tsinghua, find the person you should probably meet next.

SouLink V2 is an AI-native campus matching platform built exclusively for Tsinghua University. A user first describes who they want to meet and why. SouLink turns that request into a structured intent, recommends a small number of relevant people or opportunities, explains the fit and uncertainty, and helps both sides move toward a consent-based campus meeting.

SouLink is not a public people directory or an infinite profile feed. The current MVP supports:

- meal, study, sport, event, and conversation partners;
- private, explicit opt-in romantic matching;
- project partner, cofounder, and CTO matching;
- intent interpretation and explainable recommendations;
- mutual interest, focused match rooms, messaging, and contact reveal;
- reporting, blocking, account deletion, and admin moderation.

See the [V2 web requirements](docs/SOULINK_V2_WEB_DESIGN_REQUIREMENTS.md) and the [implementation coverage](docs/SOULINK_V2_IMPLEMENTATION.md). A Chinese version of this guide is available in [README_CN.md](README_CN.md).

## Run Locally

Requirements:

- Docker Desktop or a compatible Docker Engine;
- Docker Compose v2.

From the repository root:

```bash
docker compose up -d --build
```

See the [local deployment and acceptance guide](docs/LOCAL_DEPLOYMENT.md) for reset, health, observability, and production notes.

Open the services:

| Service | URL |
| --- | --- |
| SouLink Web | http://localhost:18080 |
| API health check | http://localhost:18081/healthz |
| Mailpit inbox | http://localhost:8025 |
| PostgreSQL | `127.0.0.1:15432` |
| Redis | `127.0.0.1:16379` |

The local stack applies all migrations and creates an idempotent Tsinghua demo workspace:

```text
Email: demo@mails.tsinghua.edu.cn
Password: Soulink2026!
```

The local-only Hack administrator is `hack@mails.tsinghua.edu.cn` with password `SoulinkHack2026!`.

The login page can fill these credentials automatically. The workspace includes profiles, intent opportunities, a structured project memo, CTO candidates, a pending application, a mutual match room, a meeting plan, and messages.

Useful commands:

```bash
docker compose ps
docker compose logs -f api
docker compose down
```

To remove local data as well:

```bash
docker compose down -v
```

## Product Loop

1. Register or sign in with a Tsinghua email.
2. Add skills, availability, preferred scenes, and personal boundaries.
3. Describe who you want to meet in natural language.
4. Answer up to three material clarification questions.
5. Review and publish a private, candidate-only, or feed-visible intent.
6. Review up to three recommendations with specific fit and uncertainty.
7. Enter a match room after mutual interest.
8. Coordinate a campus meeting, chat, and reveal contact details separately.
9. Leave feedback, report, or block when needed.

Project and CTO intents also include a project memo with stage, expected weekly time, compensation or equity expectations, required and optional skills, working style, and risk appetite.

## Architecture

| Layer | Stack | Responsibility |
| --- | --- | --- |
| Web | React 19, TypeScript, Vite 8, Tailwind CSS 4 | Intent creation, opportunity feed, recommendations, match rooms, profiles, moderation |
| API | Go 1.26, Chi, JWT, WebSocket | Authentication, intent lifecycle, ranking, consent, messages, safety |
| Data | PostgreSQL 16 | Users, profiles, intents, interests, rooms, messages, reports, blocks |
| Cache | Redis 7 | Sessions and runtime caching |
| Local mail | Mailpit | Development verification emails |
| Delivery | Docker Compose, Nginx | Builds, migrations, reverse proxy, health checks |

The V2 intent interpreter and recommender use deterministic, testable rules and explicit profile signals, so the complete product works without an external model key. AI output identifies inferred fields, reasons, and uncertainty. It never commits or communicates on behalf of a user.

## Repository Layout

```text
backend/      Go API, database migrations, and tests
frontend/     React web application
matching/     Retained V1 periodic matching engine
docs/         V2 requirements, implementation notes, migration, and privacy
deploy/       Production deployment helpers
```

The principal V2 migrations are:

- `000016_v2_intent_matching.sql`
- `000017_safety_moderation.sql`
- `000018_project_memos.sql`

## Verification

Backend:

```bash
cd backend
go test ./...
go vet ./...
```

Frontend:

```bash
cd frontend
npm ci
npm run lint
npm run build
```

Full stack:

```bash
docker compose config
docker compose up -d --build
docker compose ps
```

## Production Notes

Local demo data is controlled by `DEMO_SEED_ENABLED=true`. The API refuses to seed demo data when `APP_ENV=production`. A production deployment must still:

- use a strong random `JWT_SECRET`;
- disable demo seeding and remove the demo admin account;
- configure a real SMTP provider and HTTPS;
- restrict `ALLOWED_ORIGINS`;
- use independent PostgreSQL and Redis credentials with backups;
- establish report review, deletion, and privacy-response procedures.

See [backend/.env.example](backend/.env.example) and [backend/README.md](backend/README.md) for configuration details.

## Product Principles

- Intent before people.
- A few strong recommendations before infinite browsing.
- Mutual consent before exposure.
- AI as a matchmaker, never an impersonator.
- Real meetings before screen time.
- Trust, safety, and reversible choices as core product behavior.
