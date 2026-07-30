# SouLink V2 Web Design Requirements Draft

Draft version: v0.1
Date: 2026-07-30
Scope: Web product redesign for a Tsinghua-only AI-native campus matching platform

## 1. Product Positioning

SouLink V2 is not a swipe-based dating app and not a generic campus forum.

It should become a Tsinghua-only AI relationship matching layer: users express who they want to meet and why, and the system helps turn that intent into a trusted, mutually consented, real campus connection.

Core sentence:

> In Tsinghua, find the person you should probably meet next.

The product should support multiple relationship intents:

- Find a meal, study, sport, event, or conversation partner.
- Find a romantic match.
- Find a project partner, cofounder, CTO, designer, operator, or early teammate.
- Find someone with a specific background, experience, skill, or emotional resonance.

The product should optimize for real introductions, not screen time.

## 2. Product Principles

1. Intent first, people second.
   Users should not begin by browsing people. They should begin by saying what they want to find.

2. Few recommendations, high confidence.
   Avoid infinite candidate feeds. The default recommendation unit should be 3 to 5 strong candidates or opportunities.

3. Mutual consent before exposure.
   The product can help discover relevant people, but sensitive profile details and contact paths should only open after both sides agree.

4. AI as matchmaker, not impersonator.
   AI can ask clarifying questions, summarize fit, suggest openings, and plan next steps. It should not pretend to be a user or automate emotional commitment.

5. Campus-native and offline-oriented.
   Matching should use real Tsinghua scenes: canteens, libraries, labs, teaching buildings, playgrounds, clubs, events, weekends, course seasons, demo days.

6. Trust beats virality.
   Verification, privacy, reporting, and low-pressure consent are core product requirements, not later additions.

## 3. Target Users

Primary users:

- Tsinghua students who want to meet useful, interesting, or compatible people but do not know where to start.
- Students with concrete short-term needs: meal partner, study partner, project partner, sports partner, event companion.
- Students with serious long-term needs: romantic relationship, startup cofounder, CTO, research collaborator.

Secondary users:

- Student builders, club organizers, project initiators, and early-stage founders.
- Users who are socially cautious but willing to meet people when there is a clear reason and a safe structure.

Admin users:

- Product operators who can review reports, manage risky content, configure intent categories, and monitor matching quality.

## 4. Core Product Objects

### User Profile

A profile should include verified identity and self-expression, but avoid becoming a public performance page.

Required fields:

- Tsinghua email verification status.
- Name or nickname.
- School, department, grade, program type.
- Basic self-introduction.
- Contact visibility preference.

Optional fields:

- Skills and domains.
- Current projects.
- Interests.
- Relationship intent status.
- Availability patterns.
- Preferred meeting scenes.
- Boundaries and things the user does not want.

### AI Portrait

An AI-generated structured summary derived from profile, questionnaire, behavior, and user-approved text.

It should include:

- Personality and collaboration style.
- Social energy and communication preference.
- Interests and scenes where the user is likely to connect.
- Project or skill profile.
- Dating-relevant traits only when the user opts into romantic matching.

The portrait must be editable, dismissible, and explainable.

### Intent

An intent is the central product object.

Examples:

- "I want to find someone to have dinner with at Zijing tonight."
- "I want to meet a frontend-heavy CTO candidate for an AI product."
- "I want to find a serious romantic match, preferably someone patient and intellectually curious."
- "I want a running partner for Tuesday and Thursday nights."

Intent fields:

- Intent category.
- Natural-language description.
- Goal.
- Time window.
- Location preference.
- Visibility level.
- Matching constraints.
- Deal breakers.
- Desired next step.
- Expiration time.

### Match Proposal

A match proposal is an AI-generated recommendation connecting one user to another user or intent.

It should include:

- Candidate or opportunity summary.
- Why this may fit.
- What may not fit.
- Confidence level.
- Shared context.
- Suggested first message or action.
- Consent status.

### Match Room

A match room opens only after mutual interest or operator-approved matching.

It should include:

- Short profile snippets from both sides.
- AI explanation of fit.
- Icebreaker prompts.
- Suggested time and campus place.
- Contact unlock or in-app chat.
- Feedback after meeting.

## 5. Information Architecture

### Public Pages

1. Landing page
   Purpose: communicate the product promise and drive Tsinghua email sign-up.

2. Login and registration
   Purpose: authenticate users through Tsinghua email and create a trusted user base.

3. Privacy and safety page
   Purpose: explain consent, visibility, matching, data use, and reporting.

### Authenticated App Pages

1. Home
   The first screen should be the product itself, not a marketing page.

2. Intent composer
   A guided AI flow for turning a vague need into a matchable intent.

3. Intent feed
   A feed of active opportunities, not a feed of people.

4. Recommendations
   AI-selected people or intents relevant to the current user.

5. Match room
   A focused space for one potential connection.

6. Profile
   User identity, AI portrait, skills, interests, privacy controls.

7. My intents
   Published intents, drafts, expired intents, match history.

8. Messages or introductions
   Lightweight communication after mutual consent.

9. Feedback and safety
   Report, block, rate match quality, request deletion.

10. Admin console
    Review reports, manage categories, inspect matching health, moderate abuse.

## 6. Home Page Requirements

The home page should answer one question:

> Who do you want to meet today?

Primary UI:

- A large natural-language input box.
- Intent shortcuts:
  - Find meal partner
  - Find study partner
  - Find sport partner
  - Find romantic match
  - Find project partner
  - Find CTO or cofounder
  - Find someone to talk to
- A button to start AI clarification.

Secondary UI:

- Today on campus:
  A compact list of time-sensitive opportunities.

- Recommended for you:
  3 high-confidence recommendations.

- Your active intents:
  Status of the user's own requests.

- Scheduled matches:
  Upcoming match rooms or offline meetings.

Home should feel like a useful campus command center, not a decorative landing page.

## 7. Intent Composer Requirements

The composer should accept messy natural language and guide the user into a structured request.

Flow:

1. User enters a sentence.
2. AI identifies category, goal, urgency, and missing fields.
3. AI asks at most 3 clarifying questions.
4. User reviews a structured intent preview.
5. User chooses visibility:
   - Private AI search only.
   - Visible to recommended candidates only.
   - Visible in intent feed.
6. User publishes or saves as draft.

The AI should not over-question. It should infer reasonable defaults and let the user edit.

## 8. Intent Feed Requirements

The feed should not be a stream of people. It should be a stream of matchable needs and opportunities.

Card examples:

- "Tonight, 19:00, someone wants to find a dinner partner near Zijing."
- "A student founder is looking for a backend/infra CTO for an AI social product."
- "Someone wants a quiet library study partner for exam week."
- "A user opted into serious relationship matching this week."

Each card should show:

- Intent title.
- Category.
- Time window.
- Suggested fit reason.
- Privacy-safe user preview.
- Action:
  - Interested
  - Not for me
  - Ask AI why
  - Save

The feed should be finite and calm. No endless scroll as the primary pattern.

## 9. Recommendation Requirements

For each active intent, the system should generate a small set of recommendations.

Each recommendation should include:

- Candidate or intent summary.
- Explanation of fit.
- Possible mismatch or uncertainty.
- Shared context.
- Suggested first action.
- Consent status.

The default number should be 3. More candidates can exist, but the interface should not encourage bulk browsing.

Recommendation explanations should be specific, not generic:

Bad:

> You both like music and have similar personalities.

Good:

> You both prefer low-pressure first meetings, mentioned interest in AI tools, and are available around Thursday evening. The main uncertainty is that their project commitment preference is not yet clear.

## 10. Match Room Requirements

A match room opens after both sides show interest or when a scheduled matching cycle creates a proposal.

It should include:

- A short introduction for both sides.
- AI-generated "why you were matched."
- 3 suggested conversation starters.
- Suggested campus meeting places.
- Time coordination.
- Contact reveal controls.
- Report and block actions.
- Post-meeting feedback.

The match room should feel like a bridge to a real interaction, not another chat inbox.

## 11. Romantic Matching Requirements

Romantic matching should be slower and more private than general partner matching.

Requirements:

- Explicit opt-in.
- Separate privacy controls.
- No public romantic intent cards with identifiable details.
- Weekly or periodic curated delivery is preferred over real-time browsing.
- Users should see 1 to 3 thoughtful proposals, not a swipe deck.
- Explanation must avoid deterministic claims.
- Users can pause, skip, or narrow their preference at any time.

The tone should be respectful and grounded. Avoid language like "soulmate guaranteed" or "perfect match."

## 12. Project Partner and CTO Matching Requirements

This is a major differentiator for Tsinghua.

Users should be able to create a project memo:

- What I am building.
- Current stage.
- What kind of person I need.
- Expected weekly time.
- Equity or compensation expectations, if relevant.
- Required skills.
- Nice-to-have skills.
- Working style.
- Risk appetite.

AI should match against:

- Skills.
- Past projects.
- Availability.
- Motivation.
- Collaboration style.
- Domain interest.

The UI should feel more like a founder matching workspace than a dating app.

## 13. Profile Requirements

The profile should be modular and privacy-aware.

Sections:

- Verified Tsinghua identity.
- Basic profile.
- AI portrait.
- Skills and domains.
- Current projects.
- Interests and scenes.
- Matching preferences.
- Privacy controls.

Users should be able to preview how they appear in:

- General partner matching.
- Romantic matching.
- Project partner matching.
- Feed intent cards.

## 14. AI Feature Requirements

The AI layer should support:

1. Intent understanding
   Parse user text into structured matching fields.

2. Clarifying questions
   Ask only for fields that materially improve matching.

3. Profile understanding
   Build editable user portraits from explicit user input.

4. Matching and ranking
   Recommend people or intents based on compatibility and constraints.

5. Explanation generation
   Explain fit and uncertainty in human language.

6. Icebreaker generation
   Suggest opening messages or first-meeting topics.

7. Meeting planning
   Suggest campus place, timing, and low-pressure next step.

8. Safety assistance
   Flag abusive, manipulative, explicit, or privacy-invasive content for review.

AI output must be transparent and reversible. Users should understand when AI is making an inference.

## 15. Privacy and Safety Requirements

Required controls:

- Tsinghua email verification.
- Visibility settings per intent.
- Separate romantic visibility controls.
- Contact reveal only after consent.
- Block user.
- Report user or intent.
- Delete account.
- Delete or expire intent.
- Hide from specific categories.
- Pause matching.

Privacy rules:

- Do not expose full profiles in public feed.
- Do not allow unrestricted people search by default.
- Do not reveal contact information before mutual consent.
- Do not use private romantic preferences for project matching.
- Do not generate sensitive inferences that users did not provide or approve.

## 16. Visual Design Direction

The product should feel:

- Trustworthy.
- Intelligent.
- Campus-native.
- Calm but alive.
- More like a useful relationship operating system than a playful swipe toy.

Visual guidance:

- Use Tsinghua purple as an accent, not as a full-screen purple gradient.
- Use restrained neutrals, warm white, ink tones, and small purple highlights.
- Prefer dense, useful cards over oversized decorative hero sections.
- The first authenticated screen should be the actual matching interface.
- Cards should be compact and scannable.
- Avoid a "dating app" visual language dominated by faces and likes.
- Use icons for intent categories and actions.
- Make trust controls visible but not alarming.

## 17. Content Tone

Tone should be:

- Direct.
- Warm.
- Low-pressure.
- Specific.
- Slightly poetic only in brand moments.

Avoid:

- Overpromising love or success.
- Gamified scarcity.
- Ranking people.
- Manipulative urgency.
- "AI knows you better than yourself" language.

Preferred language:

- "This may be worth a conversation."
- "Here is why this person may fit your request."
- "You can choose what to reveal."
- "No response is also fine."
- "Meet when both sides are ready."

## 18. Metrics

Primary metrics:

- Intent creation rate.
- Intent completion rate.
- Recommendation acceptance rate.
- Mutual opt-in rate.
- Match room open rate.
- Offline meeting confirmation rate.
- Post-meeting positive feedback rate.

Quality metrics:

- Report rate.
- Block rate.
- Match explanation helpfulness.
- Percentage of users who feel recommendations are relevant.
- Number of low-quality or expired intents.

Retention metrics:

- Weekly active users.
- Repeat intent creation.
- Returning users after successful matches.

Do not optimize for:

- Infinite scroll time.
- Number of profiles viewed.
- Swipe count.
- Message volume without meeting intent.

## 19. MVP Scope

Recommended MVP:

1. Tsinghua-only registration and login.
2. Profile and privacy settings.
3. AI intent composer.
4. Intent feed.
5. AI recommendations for each intent.
6. Mutual interest flow.
7. Match room.
8. Basic in-app messaging or contact reveal.
9. Feedback after match.
10. Admin report review.

Defer:

- Multi-school support.
- Public people search.
- Complex social graph.
- Group matching.
- Native mobile app.
- Fully autonomous AI agent communication.
- Payment.

## 20. Non-Goals

SouLink V2 should not be:

- A swipe dating app.
- A public directory of Tsinghua students.
- A campus confession wall.
- A generic forum.
- A LinkedIn clone.
- An AI companion app.
- A product where AI pretends to be users.

## 21. Open Questions for Review

1. Should romantic matching be a separate mode or one intent category inside the same product?
2. Should users be allowed to publish anonymous intent cards?
3. Should CTO/project matching require a project memo before recommendation?
4. Should the system support group intents in the first version?
5. What level of Tsinghua identity should be shown before mutual consent?
6. Should the product start with weekly curated matching or real-time intent matching?
7. How much AI explanation should be visible to the other side?
8. What safety review threshold is acceptable for an early campus MVP?
9. Should contact reveal happen inside SouLink or should the product push users to WeChat?
10. What is the one launch scene most likely to create word-of-mouth at Tsinghua?

## 22. Suggested V2 First Release Shape

The strongest first release should be narrow:

> Tsinghua-only intent matching for meals, study, serious dating, and project partners.

The home page should start with one action:

> Tell SouLink who you want to meet.

The system should then guide the user from intent to 3 recommendations to mutual consent to a real campus meeting.

This is the product loop that can make SouLink feel truly AI-native.
