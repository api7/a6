# AI Agent Skills

This document describes the skill system for the a6 CLI. Skills are structured knowledge files that enable AI coding agents to work effectively with APISIX through the a6 CLI.

## Overview

Skills are `SKILL.md` files stored in the `skills/` directory. Each skill provides domain-specific instructions, command patterns, and decision guidance for AI agents. The format is compatible with 39+ AI coding agents including Claude Code, OpenCode, Cursor, GitHub Copilot, and Windsurf.

## Directory Structure

```
skills/
├── a6-shared/SKILL.md              # Core a6 conventions (shared skill)
├── a6-plugin-key-auth/SKILL.md     # key-auth plugin skill
├── a6-plugin-jwt-auth/SKILL.md     # jwt-auth plugin skill
├── a6-recipe-blue-green/SKILL.md   # Blue-green deployment recipe
├── a6-persona-operator/SKILL.md    # Platform operator persona
└── ...
```

Each skill lives in its own directory: `skills/<skill-name>/SKILL.md`.

## Skill Taxonomy

Skills follow a naming convention with four types:

| Prefix | Type | Description | Example |
|--------|------|-------------|---------|
| `a6-shared` | Shared | Core project conventions and patterns | `a6-shared` |
| `a6-plugin-*` | Plugin | One APISIX plugin — config, examples, gotchas | `a6-plugin-key-auth` |
| `a6-recipe-*` | Recipe | Multi-step operational task | `a6-recipe-blue-green` |
| `a6-persona-*` | Persona | Role-specific workflow guidance | `a6-persona-operator` |

### Naming Rules

- **Format**: kebab-case
- **Pattern**: `^[a-z0-9]+(-[a-z0-9]+)*$`
- **Directory name must match the `name` field in frontmatter**

## SKILL.md Format

Every skill file has two parts: YAML frontmatter and Markdown body.

### Frontmatter (Required)

```yaml
---
name: a6-plugin-key-auth
description: >-
  Skill for configuring key-auth plugin on APISIX routes and consumers
  using the a6 CLI. Covers API key creation, consumer binding, and
  key lookup configuration.
version: "1.0.0"
author: Apache APISIX Contributors
license: Apache-2.0
metadata:
  category: plugin
  apisix_version: ">=3.0.0"
  plugin_name: key-auth
  a6_commands:
    - a6 route create
    - a6 consumer create
    - a6 plugin get key-auth
---
```

**Required fields:**

| Field | Description |
|-------|-------------|
| `name` | Skill identifier. Must match directory name. Kebab-case. |
| `description` | Multi-line description of what this skill covers. |

**Recommended fields:**

| Field | Description |
|-------|-------------|
| `version` | Semantic version of the skill content. |
| `author` | Who authored the skill. |
| `license` | License identifier (e.g., `Apache-2.0`). |
| `metadata` | Structured metadata for categorization and filtering. |

### Body (Markdown)

The body follows the skill type:

**Plugin skills** typically include:
- What the plugin does (one paragraph)
- When to use it (bullet list of scenarios)
- Configuration reference (key fields, types, defaults)
- Step-by-step: enable on a route
- Step-by-step: configure with consumers
- Common patterns and variations
- Troubleshooting / common mistakes

**Recipe skills** typically include:
- Goal description
- Prerequisites
- Step-by-step instructions with a6 commands
- Verification steps
- Rollback procedure

**Persona skills** typically include:
- Role description and responsibilities
- Common workflows
- Decision trees
- Which other skills to load for each task

## CI Validation

Every PR that modifies `skills/` is validated by `scripts/validate-skills.sh`. The script checks:

1. Every `skills/*/SKILL.md` has valid YAML frontmatter
2. Required fields `name` and `description` are present
3. `name` matches the directory name
4. `name` follows kebab-case pattern
5. `description` is non-empty

Run locally:

```bash
make validate-skills
```

## Adding a New Skill

1. Choose the skill type and name following the [taxonomy](#skill-taxonomy)
2. Create the directory: `mkdir skills/<skill-name>`
3. Create `skills/<skill-name>/SKILL.md` with frontmatter and body
4. Run validation: `make validate-skills`
5. Update this document if adding a new skill type or category

## Skill Roadmap

| PR | Skills | Description |
|----|--------|-------------|
| PR-28 | 1 | Infrastructure + `a6-shared` |
| PR-29 | 5 | Authentication plugins (key-auth, jwt-auth, basic-auth, hmac-auth, openid-connect) |
| PR-30 | 4 | Security + rate limiting (ip-restriction, cors, limit-count, limit-req) |
| PR-31 | 5 | Traffic + transformation (proxy-rewrite, response-rewrite, traffic-split, redirect, grpc-transcode) |
| PR-32 | 5 | Operational recipes (blue-green, canary, circuit-breaker, health-check, mtls) |
| PR-33 | 4 | AI Gateway (ai-proxy, ai-prompt-template, ai-prompt-decorator, ai-content-moderation) |
| PR-34 | 6 | Observability (prometheus, skywalking, zipkin, http-logger, kafka-logger, datadog) |
| PR-35 | 5 | Advanced plugins (serverless, ext-plugin, fault-injection, consumer-restriction, wolf-rbac) |
| PR-36 | 5 | Advanced recipes + personas |

**Total**: 40 skills across 9 PRs.
