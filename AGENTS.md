# Agentic Engineering Standards (AGENTS.md)

This document mandates the standards for automated agents and human contributors working on the SettlerEngine codebase.

## 1. Binary Management
To maintain a clean, high-performance repository and prevent "git bloat," the following rules are strictly enforced:

*   **Output Directory:** All compiled binaries (executables, test binaries, profiles) MUST reside in the `bin/` directory at the project root.
*   **No Codebase Pollution:** Binaries MUST NOT be placed within module directories (e.g., `core/`, `pkg/`, `apps/`).
*   **Git Exclusion:** The `bin/` directory MUST be included in the `.gitignore` file.
*   **Zero-Commit Policy:** Binaries MUST NEVER be committed to the version control system. Any agent or contributor who accidentally commits a binary is responsible for purging it from the Git history using `git filter-repo` or `git filter-branch`.

## 2. Workspace Integrity
*   Always use `go work` to manage multi-module dependencies.
*   Ensure `go mod tidy` and `go work sync` are executed after any dependency or structural changes.

## 3. Compliance
Agents are programmed to automatically reject or revert any pull requests that violate these binary management standards.
