# dailyup

CLI tool that queries Azure DevOps and produces a Markdown summary of your work within a sprint — useful for standups, weekly updates, or status reports.

## Installation

```bash
go install github.com/pathcl/dailyup@latest
```

Or build from source:

```bash
git clone https://github.com/pathcl/dailyup
cd dailyup
go build -o dailyup .
```

## Authentication

dailyup uses the Azure CLI credential — no PAT tokens or secrets to manage.

**1. Install the Azure CLI**

```bash
# macOS
brew install azure-cli

# Linux (Debian/Ubuntu)
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
```

**2. Log in**

```bash
az login
```

This opens a browser window. After completing the login, the CLI caches a token locally. dailyup picks it up automatically on every run.

**3. (If you have multiple tenants) select the right one**

```bash
az account list --output table
az account set --subscription "<subscription-name-or-id>"
```

You can verify the active account at any time:

```bash
az account show
```

Tokens are refreshed automatically by the Azure CLI — you only need to re-run `az login` if your session expires (typically after several days of inactivity).

## Configuration

Create the config file at `~/.config/dailyup/config.toml`:

```bash
mkdir -p ~/.config/dailyup
```

```toml
# ~/.config/dailyup/config.toml

# Required
organization = "myorg"       # Azure DevOps organization name
                              # e.g. https://dev.azure.com/myorg → "myorg"
project      = "myproject"   # Team project name

# Optional — work item filters
tags         = ["backend"]   # Narrow sprint results to items with these tags (OR logic per tag, AND with sprint)
assigned_to  = "@Me"         # Filter by assignee: "@Me" for yourself, or a display name like "Jane Doe"

# Optional — PR and commit window
weeks        = 2             # How many weeks back to fetch PRs and commits (default: 2)
email        = "you@example.com"  # Your email address, used to filter commits by author
```

### Configuration reference

| Field          | Required | Default | Description |
|----------------|----------|---------|-------------|
| `organization` | yes      | —       | Azure DevOps org name (the subdomain in `dev.azure.com/<org>`) |
| `project`      | yes      | —       | Team project name |
| `tags`         | no       | —       | Work item tags to filter on; omit to get all items in the sprint |
| `assigned_to`  | no       | —       | `@Me` for the current user, or a display name/email |
| `weeks`        | no       | `2`     | Look-back window for PRs and commits |
| `email`        | no       | —       | Your email, used as the commit author filter |

## Usage

```bash
# Summarise the current sprint (uses @CurrentIteration)
dailyup summary

# Summarise a specific named sprint
dailyup summary --sprint "Sprint 68"

# Override the assignee filter at runtime
dailyup summary --sprint "Sprint 68" --assigned-to "@Me"
dailyup summary --sprint "Sprint 68" --assigned-to "Jane Doe"

# Override the PR/commit look-back window
dailyup summary --weeks 1

# Use a different config file
dailyup summary --config /path/to/config.toml
```

### Output

The command prints Markdown to stdout, ready to paste into a doc, PR description, or chat message:

```markdown
# Work Summary — May 12, 2026 – May 26, 2026

## Work Items

### Task
- [#1234] Implement feature X (Active) — tags: backend
- [#1235] Fix edge case (Closed) — tags: backend

### User Story
- [#1100] User can export data (In Progress)

## Pull Requests
- [#42] feat: Add export endpoint — **Merged** — my-repo — 2026-05-24

## Commits
- `abc1234` Implement export logic — my-repo — 2026-05-24
```

To save to a file:

```bash
dailyup summary > update.md
```
