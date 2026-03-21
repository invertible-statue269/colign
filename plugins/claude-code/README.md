# Colign Plugin for Claude Code

Connect Claude Code to the Colign spec management platform. Read, write, and manage specs directly from your terminal.

No binary to install — connects via Streamable HTTP.

## Setup

### 1. Generate an API Token

Go to Colign Settings > AI & API Keys and generate a new API token.

### 2. Set Environment Variable

```bash
export COLIGN_API_TOKEN=col_your_token_here

# Optional: for self-hosted instances
export COLIGN_MCP_URL=https://your-instance.com/mcp  # defaults to https://app.colign.dev/mcp
```

### 3. Install the Plugin

```bash
# Local testing
claude --plugin-dir ./plugins/claude-code

# Or install from marketplace
claude plugin install colign
```

## Workflow

The plugin follows Colign's change lifecycle:

```
onboard → explore → propose → plan → implement → complete
```

| Skill | Stage | Description |
|-------|-------|-------------|
| `/colign:onboard` | Setup | Verify MCP connection and API token |
| `/colign:explore` | Any | Browse projects, read specs, check change status |
| `/colign:propose` | Draft → Problem | Define the problem, scope, and write a structured proposal |
| `/colign:plan` | Problem → Solution | Break proposal into architecture, implementation steps, and tasks |
| `/colign:implement` | Solution → Review | Code against the spec, update task progress |
| `/colign:complete` | Review → Done | Verify all tasks are done, advance the workflow |

Skills auto-trigger based on context (e.g., "implement the next task") or can be invoked explicitly.

## MCP Tools

| Tool | Description |
|------|-------------|
| `list_projects` | List all accessible projects |
| `get_change` | Get change details including stage and artifacts |
| `read_spec` | Read a spec document (proposal, design, spec, tasks) |
| `write_spec` | Write or update a spec document |
| `list_tasks` | List implementation tasks for a change |
| `update_task` | Update a task's status (todo, in_progress, done) |
| `suggest_spec` | Get AI suggestions for improving a spec |

## Example Usage

```
# Explicit skill invocation
> /colign:explore
> Show me the current spec for change 42

# Auto-triggered by context
> I want to build a new notification system
  (triggers: propose)

> Break this down into implementation tasks
  (triggers: plan)

> Start coding the first task
  (triggers: implement)

> All tasks are done, let's wrap up
  (triggers: complete)
```
