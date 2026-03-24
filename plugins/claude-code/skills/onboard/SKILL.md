---
name: onboard
description: Set up and verify the Colign MCP connection. Use when the user first installs the plugin, says "set up colign", "connect to colign", asks how to get started, or when any colign MCP tool call fails with a connection or authentication error.
---

# Get Started with Colign

Verify the connection and guide the user on what to do next based on their current project state.

---

## Step 1: Verify Connection

Call `mcp__colign__list_projects` to verify the MCP server is working.

- **Success**: Continue to Step 2
- **Auth error**: Token is invalid or expired — guide the user to Colign Settings > AI & API Keys to regenerate
- **Connection error**: Check if the MCP server URL is correct and reachable

## Step 2: Show Current State

Call `mcp__colign__list_projects` and for each project, optionally call `mcp__colign__get_project_dashboard` to understand what's in progress.

Present a quick overview:

```
Connected to Colign!

Projects:
● My Project — 2 changes (1 in Design, 1 in Draft)
● Other Project — no changes yet
```

## Step 3: Recommend Next Action

Based on what you found, suggest the most relevant skill:

| Situation | Recommendation |
|-----------|---------------|
| No projects exist | "Create a project first on the Colign web app, then come back" |
| Project exists, no changes | `/colign:propose` — "Start by proposing a new change" |
| Change exists with proposal only | `/colign:plan` — "There's a proposal ready. Want to plan the implementation?" |
| Change exists with design + tasks | `/colign:implement` — "Tasks are ready. Want to start implementing?" |
| All tasks done, not advanced | `/colign:complete` — "Looks like implementation is done. Ready to wrap up?" |
| Just want to look around | `/colign:explore` — "Let's explore what's in the project" |

## Available Skills

```
/colign:explore   — Browse projects and think through ideas
/colign:propose   — Write a proposal for a new change
/colign:plan      — Design the implementation and create tasks
/colign:implement — Code against the task list
/colign:complete  — Verify, advance stage, and archive
```

---

## Error Recovery

This skill should also trigger when other colign skills fail due to:
- Authentication failures (401)
- Connection refused errors
- Missing API token errors

In these cases, guide the user through the relevant fix step.
