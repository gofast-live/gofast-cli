# DESC: Generate Domain `CONTEXT.md`

Generate a structured `CONTEXT.md` for a specific folder from one inline prompt.

Trigger example:
- `@workflows/DESC.md - lib/meetings, capture context for meetings domain`

Expected output for example:
- `lib/meetings/CONTEXT.md`

---

## 1. Parse Input

Input format:
- `<target_folder>, <short freeform context>`

Rules:
- First comma-separated segment is `target_folder`.
- Remaining text is domain context/hints.
- If no comma is provided, treat whole input as `target_folder` and proceed.

---

## 2. Scope of Investigation

Primary scope:
- All files under `target_folder/**`

Secondary scope (only when referenced by code/tests):
- Direct dependencies in neighboring modules
- Related jobs/workers, GraphQL/LiveView entry points, telemetry, and auth/scoping paths

Use fast discovery first (`rg`, file tree, key entry points), then deep-read only important files.

---

## 3. Output Location

Always write to:
- `<target_folder>/CONTEXT.md`

If file exists:
- Refresh/restructure in place.
- Remove stale duplicates.
- Keep still-accurate content.

---

## 4. Required `CONTEXT.md` Structure

Start from:
- `workflows/TEMPLATE.md`

Use this section order:
1. Metadata
2. Context Maintenance Protocol (LLM-First)
3. Summary
4. Test Strategy
5. Architecture (with Graphs)
6. File Tree (Curated)
7. Core Contracts
8. Telemetry and Observability
9. Current Work
10. Future Work
11. Critical Invariants and Tricky Flows
12. Quick Reference APIs
13. Runbook

Formatting rules:
- Mark sections `[STABLE]` or `[VOLATILE]`.
- Use concise, factual language.
- Prefer Mermaid for architecture/flow diagrams.
- Include file paths for non-obvious claims.
- Keep auth, tenant scoping, and failure/retry behavior explicit.

---

## 5. Quality Gate

Before finishing, verify:
- `CONTEXT.md` reflects current code, not assumptions.
- Test strategy is explicit, including unit vs integration guidance.
- LLM default testing policy is present (update/add unit tests for changes; add integration tests for boundary/contract changes).
- Security/scoping invariants are explicit.
- Telemetry update points are listed.
- Current vs Future work are separated.
- `Last updated` is today.

Final response should report:
- written file path
- short summary of what was captured
- unknowns (if any)
