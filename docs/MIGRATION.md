# Repository migration record

This monorepo was assembled from four private SouLink repositories while
preserving their reachable Git histories.

## Source mapping

| Original repository | Monorepo directory | Preserved commits across all refs | Historical branches | Historical tags |
| --- | --- | ---: | ---: | ---: |
| `THU-Soulink-Team/Soulink-Web` | `frontend/` | 148 | 10 | 3 |
| `THU-Soulink-Team/Soulink-Web-Backend` | `backend/` | 115 | 11 | 3 |
| `THU-Soulink-Team/Soulink-Web-Deploy` | `deploy/` | 3 | 2 | 0 |
| `THU-Soulink-Team/Soulink-Web-Matching` | `matching/` | 13 | 1 | 0 |

The component histories contain 279 source commits in total. Three merge commits
join the four default branches, and two integration commits add the monorepo
documentation, safe templates, compatibility fixes, and public product
presentation. That makes 284 commits reachable across all refs in this
repository, of which 274 are reachable from the combined `main` branch.

## Ref layout

Original branches are retained as:

```text
history/frontend/<original-branch>
history/backend/<original-branch>
history/deploy/<original-branch>
history/matching/<original-branch>
```

Original tags are retained as:

```text
history/frontend/<original-tag>
history/backend/<original-tag>
```

The namespace prevents identically named branches and tags from colliding.

## What was preserved

- commit messages (one Git-generated revert hash was updated to the
  corresponding rewritten commit ID);
- author names and email addresses;
- author timestamps;
- committer timestamps;
- parent/merge relationships;
- every commit reachable from the original branches and tags.

An automated comparison verified the author/committer identity and timestamps
for all 279 source commits, along with every commit's parent count. Messages
matched byte-for-byte for 278 commits; the single exception is the mapped revert
reference described above.

Commit IDs necessarily changed because Git commit hashes include the tree and
parent hashes. Moving each repository into a subdirectory and sanitizing
sensitive content therefore produced new IDs.

## Sanitation record

The migration removed these categories from every reachable revision:

- committed `.env` files;
- local questionnaire and matching import configuration;
- SMTP account files that contained or resembled credentials;
- an experimental mail script containing a historical API credential;
- routable server IP addresses in scripts, documentation, and configuration;
- community QR-code images.

Safe `.env.example` and SMTP example files were then recreated with empty or
clearly non-secret values. The final repository was scanned over its complete
history with Gitleaks and a separate IPv4 rule.

No secret value is recorded in this document. Credentials that existed in any
source repository should still be rotated, because history rewriting is not a
substitute for revocation.
