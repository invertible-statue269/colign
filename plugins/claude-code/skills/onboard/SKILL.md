---
name: onboard
description: Set up and verify the Colign MCP connection. Use when the user first installs the plugin, says "set up colign", "connect to colign", asks how to get started, or when any colign MCP tool call fails with a connection or authentication error.
---

# Onboard to Colign

Guide the user through connecting to the Colign MCP server.

## Workflow

### Step 1: Check API token

```bash
echo $COLIGN_API_TOKEN
```

If `COLIGN_API_TOKEN` is not set:
1. Tell the user to go to **Colign Settings > AI & API Keys**
2. Click **Generate Token** and name it (e.g., "Claude Code")
3. Copy the token (it's only shown once)
4. Set it: `export COLIGN_API_TOKEN=col_...`

### Step 2: Check MCP URL (optional)

```bash
echo $COLIGN_MCP_URL
```

- **SaaS (default)**: No action needed — defaults to `https://app.colign.dev/mcp`
- **Self-hosted**: Set `export COLIGN_MCP_URL=https://your-instance.com/mcp`
- **Local dev**: Set `export COLIGN_MCP_URL=http://localhost:8080/mcp`

### Step 3: Verify the connection

Try calling `mcp__colign__list_projects` to verify the MCP server is working.

- **Success**: Show the project list and confirm everything is connected
- **Auth error**: Token is invalid or expired — regenerate in Colign Settings
- **Connection error**: Check if the URL is correct and the server is reachable

### Step 4: Confirm setup

```
Colign MCP: Connected
URL: [url]
Projects: [count] accessible

You're all set! Here's how to get started:
- /colign:explore — Browse your projects and specs
- /colign:propose — Start a new change with a proposal
```

## Error Recovery

This skill should also trigger when other colign skills fail due to:
- Authentication failures (401)
- Connection refused errors
- Missing API token errors

In these cases, guide the user through the relevant fix step.
