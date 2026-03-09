# Ship: Fire-and-Forget PR

When the user triggers this flow (e.g. "@workflows/SHIP.md - change drawer to modal"), follow these steps strictly to ensure an isolated, high-speed delivery.

## 1. Isolation
*   **Generate Branch Name:** Create a descriptive kebab-case branch name based on the request (e.g., `feat/drawer-to-modal`, `fix/login-bug`).
*   **Create Worktree:** Create a new worktree in the `worktree/` directory to keep the main working directory clean.
    ```bash
    git fetch origin main
    git worktree add worktree/ship-<branch-name> -b <branch-name> origin/main
    ```

## 2. Implementation
*   **Context:** All file operations (read, write, search) **MUST** be performed within the `worktree/ship-<branch-name>` directory.
*   **Reading Files:** Some tools may fail on gitignored directories (like `worktree/`). If a read tool fails, fall back to shell commands (e.g., `cat <file_path>`).
*   **Execute:** Apply the requested code changes, refactors, or bug fixes.
*   **Tests:** Add or update any relevant tests for your changes. Do NOT run tests — the worktree lacks full app setup. Let CI/CD handle test execution.

> **⚠️ WORKTREE DISCIPLINE — READ THIS CAREFULLY**
>
> You MUST use the worktree path (`worktree/ship-<branch-name>/...`) for ALL file operations during implementation. You have a tendency to drift back to the repo root and read/write files there — this means you are working on the **wrong branch** and your changes will be lost or applied to the wrong place.
>
> Rules:
> 1. **Every** `Read`/`Edit`/`Write` call must use an absolute path starting with the worktree directory.
> 2. **Every** `Grep`/`Glob` call must have `path` set to the worktree directory.
> 3. **Before each file operation**, mentally verify the path contains `worktree/ship-`. If it doesn't, you are working in the wrong place.
> 4. **Never** read from the repo root to "check something quickly" — the root is a different branch.

## 3. Delivery
*   **Review (MANDATORY):** Before committing, verify your changes are correct and complete. Do NOT skip this step.
    ```bash
    cd worktree/ship-<branch-name>
    git status
    git diff
    ```
*   **Commit:**
    ```bash
    git add .
    git commit -m "<type>: <description>"
    ```
*   **Push:**
    ```bash
    git push -u origin <branch-name>
    ```
*   **Pull Request:**
    ```bash
    gh pr create --title "<Title>" --body "<Description>" --head <branch-name>
    ```

## 4. Cleanup
*   **Remove worktree:**
    ```bash
    git worktree remove worktree/ship-<branch-name>
    ```

## 5. Final Output
*   **Report:** Output *only* the link to the created Pull Request.
