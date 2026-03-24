---
name: propose
description: Create or update a structured proposal for a Colign change. Use when the user wants to define a new feature, write a problem statement, scope work, says "let's write a proposal", "create a new change", "I want to build X", or any request that involves capturing requirements and saving them to Colign.
---

# Propose a Change

Create a structured proposal and save it to the Colign platform. This captures the "why" and "what" — not the "how".

**Input**: The user describes what they want to build, or names an existing change to update.

---

## Workflow

1. Call `mcp__colign__list_projects` to find the target project
2. If no change exists yet, call `mcp__colign__create_change` with the project ID and a name
3. If a change already exists, call `mcp__colign__read_spec` with `doc_type: proposal` to check for existing content
4. Optionally call `mcp__colign__get_memory` for project conventions and context
5. Gather context from the user about what they want to build
6. Draft a structured proposal:

```markdown
## Problem
Why is this change needed? What user pain does it solve?

## Scope
What specifically will change? Be concrete about deliverables.

## Out of Scope
What is explicitly NOT included in this change?
```

7. Present the draft to the user for review
8. After approval, call `mcp__colign__write_spec` with `doc_type: proposal` to save

---

## Guidelines

- **Keep proposals concise** — focus on "why" and "what", not implementation details
- **Always check existing content** before overwriting — call `read_spec` first
- **Ask when unclear** — don't guess requirements, ask the user
- **Use project memory** — call `get_memory` for project conventions if available
- **Don't design yet** — architecture, data models, and task breakdowns belong in the plan phase

---

## Next Steps

After the proposal is saved, guide the user based on what they need:

```
Proposal saved!

→ To continue with design and tasks: /colign:plan
→ Or stop here — another team member can pick it up from the platform.
```
