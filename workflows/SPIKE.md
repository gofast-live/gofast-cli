# Spike: Codebase Investigation

Research a topic in the codebase and produce a concise, reusable summary. **No code changes.**

Trigger: `@workflows/SPIKE.md — how does calendar sync work?`

## 1. Scope

*   Clarify the question. If ambiguous, ask before diving in.
*   Pick a short kebab-case name for the spike (e.g. `calendar-sync`).
*   Identify likely starting points (modules, contexts, entry points).

## 2. Investigate

Run these steps in parallel where possible (e.g. using subagents or concurrent tool calls):

### 2a. Broad search
*   Grep for keywords, find relevant modules and entry points.
*   Map the **file structure** — which directories and files are involved? List them as a tree.

### 2b. Deep read
*   Follow the call chain through key files. Understand how data flows.
*   **Map dependencies** — for each key module, identify:
    *   What calls it (callers / entry points)
    *   What it calls (dependencies / downstream)
    *   This is the blast radius — document it.

### 2c. Context
*   **Tests** — tests document edge cases and expected behavior better than comments.
*   **Config** — feature flags, environment variables, application config that affect behavior.

## 3. Output

Write the findings to `workflows/spikes/SPIKE-<spike-name>.md` using this template:

```markdown
# Spike: <Title>

> <1-2 sentence summary of how it works>

## File Structure

<ASCII tree of the relevant directories/files>

## Key Files

| File | Purpose |
|------|---------|
| `path/to/file.ex:fn_name` | what it does |
| `path/to/other.ex:fn_name` | what it does |

## Dependency Graph

<ASCII diagram showing callers → module → dependencies>

Example:
    LiveView.mount ──→ Calendar.Context.sync/2
                            ├──→ Google.API.list_events/1
                            ├──→ Repo.insert_all/2
                            └──→ PubSub.broadcast/2

## Flow

1. Step one
2. Step two
3. ...

## Gotchas

- <anything surprising, non-obvious, or easy to break>
```

Keep it concise. Link to files, not paragraphs of explanation. If the answer is simple, the output should be simple.
