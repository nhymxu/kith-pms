<!-- intent-skills:start -->
## Skill Loading

Before substantial work:
- Skill check: run `npx @tanstack/intent@latest list`, or use skills already listed in context.
- Skill guidance: if one local skill clearly matches the task, run `npx @tanstack/intent@latest load <package>#<skill>` and follow the returned `SKILL.md`.
- Monorepos: when working across packages, run the skill check from the workspace root and prefer the local skill for the package being changed.
- Multiple matches: prefer the most specific local skill for the package or concern you are changing; load additional skills only when the task spans multiple packages or concerns.
<!-- intent-skills:end -->

This file provides guidance to AI when working with code in this repository.

See `README.md` for full stack details, auth contract, frontend conventions, known gotchas, and development workflow.

---

## Code Style

- **No pointless comments**: Only add comments when explaining non-obvious "why" decisions — never describe what the code does.
