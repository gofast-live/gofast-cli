# PR Review

When the user triggers this flow (e.g. "@workflows/PR.md - fix/quiet-calendar-sync-log"), review the pull request thoroughly and report findings.

## 1. Gather Context
*   **Checkout the branch** into a worktree so you can read full files, not just diffs:
    ```bash
    git fetch origin <branch-name>
    git worktree add worktree/review-<branch-name> origin/<branch-name>
    ```
*   **Fetch PR details:**
    ```bash
    gh pr view <branch-name> --json title,body,files
    gh pr diff <branch-name>
    ```
*   **Verify (MANDATORY):** Confirm you are reading from the correct worktree before proceeding.
    ```bash
    cd worktree/review-<branch-name>
    git branch --show-current
    pwd
    ```
*   **Read changed files in full** from the worktree — not just the diff. Understand the surrounding code, callers, and related modules.

> **⚠️ WORKTREE DISCIPLINE — READ THIS CAREFULLY**
>
> You MUST use the worktree path (`worktree/review-<branch-name>/...`) for ALL file reads during the review. You have a tendency to drift back to the repo root and read files from there — this means you are reading the **wrong branch** and your review will be incorrect.
>
> Rules:
> 1. **Every** `Read` call must use an absolute path starting with the worktree directory.
> 2. **Every** `Grep`/`Glob` call must have `path` set to the worktree directory.
> 3. **Before each file read**, mentally verify the path contains `worktree/review-`. If it doesn't, you are reading from the wrong place.
> 4. **Never** read from the repo root to "check something quickly" — the root is a different branch.

> **⚠️ DO NOT run tests, builds, or any other commands in the worktree.** This is a static review only — read the code, don't execute it.

## 2. Review Priorities

Focus on what matters. **3 critical findings beat 20 nitpicks.**

### Critical (always check)
1.  **Security** — SQL injection, auth bypasses, missing permission checks, exposed secrets, XSS. Double-check every raw SQL query and every auth/authorization path.
2.  **Correctness** — Logic bugs, race conditions, missing error handling at system boundaries, data loss risks.
3.  **Data integrity** — Missing multi-tenant scoping (account_id), unsafe migrations, broken constraints.

### Important (check if relevant)
4.  **Performance** — N+1 queries, missing indexes for new queries, unbounded lists.
5.  **Breaking changes** — API contract changes, removed fields, changed return types.

### Skip
*   Style nitpicks already covered by formatters/linters
*   Defensive programming suggestions (don't add guards for impossible states)
*   Missing docs/comments on clear code
*   Hypothetical future concerns

## 3. Cleanup (BEFORE reporting)

> **⚠️ MANDATORY CLEANUP — DO NOT SKIP**
>
> You MUST remove the worktree **before** writing your review output.
> LLMs routinely skip this step. You are not special — you will skip it too unless you do it RIGHT NOW.
>
> Run this command immediately after finishing your analysis, **before** composing your response:
> ```bash
> git worktree remove worktree/review-<branch-name>
> ```
> **Do NOT proceed to step 4 until the worktree is removed.** If removal fails, run `git worktree remove --force worktree/review-<branch-name>`.

## 4. Output

Report findings grouped by severity:

```
## 🔴 Critical
<issue + file:line + why it matters + suggested fix>

## 🟡 Important
<issue + file:line + suggestion>

## ✅ Looks Good
<brief note on what was checked and passed>
```

If there are no critical or important findings, say so clearly — don't invent issues to fill the report.
