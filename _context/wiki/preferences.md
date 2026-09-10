# Working Preferences & Standards

## How I use AI

- **Bug fixes** — the most common use case. Understand the problem before suggesting a fix.
- **Version bumps** — help identify all places that need updating when tracking a new `icp-cert-manager` release.
- **Understanding unfamiliar code** — orientation to areas not recently touched.

## Communication preferences

- Be direct and technical. No filler phrases ("Great!", "Certainly!", etc.).
- Investigate before answering — never speculate about code you haven't read.
- When explaining unfamiliar code: start with what it *does*, then how it works.
- Flag assumptions explicitly.

## Code style & engineering standards

- **Minimal changes.** Produce the smallest diff that solves the problem. No opportunistic refactors.
- **Trace every changed line** back to the stated requirement.
- **Go conventions.** Follow standard Go idioms and the existing style in the file being edited.
- **Scope discipline.** This operator does one thing — deploy and manage `icp-cert-manager`. Do not add complexity here.

## What to be careful about

- **This operator is a prerequisite.** Changes that cause it to fail on install will block all of CPfs.
- **Version bump completeness.** When bumping `icp-cert-manager` versions, check all image references and any version-pinned dependencies.
- **Do not re-add cert rotation logic.** That was intentionally moved out. If cert-rotation work comes up, it belongs in `ibm-common-service-operator`.
