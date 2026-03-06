# Documentation Maintenance Specification

### Purpose
This document defines the mandatory rules for maintaining a6 documentation. Every code change that affects user-facing behavior MUST include corresponding documentation updates. This is enforced in code review.

### Documentation Structure
```
docs/
├── admin-api-spec.md              # APISIX Admin API reference (rarely changes)
├── adr/                           # Architecture Decision Records
│   └── 001-tech-stack.md         # Tech stack decisions
├── golden-example.md              # Reference implementation template
├── coding-standards.md            # Code style guide
├── testing-strategy.md            # Test patterns and requirements
├── documentation-maintenance.md   # THIS file
└── user-guide/                    # End-user documentation
    ├── getting-started.md         # Installation + first commands
    ├── configuration.md           # Context management, config file format
    ├── route.md                   # Route command reference
    ├── upstream.md                # Upstream command reference
    ├── service.md                 # Service command reference
    ├── consumer.md                # Consumer command reference
    ├── ssl.md                     # SSL command reference
    ├── plugin.md                  # Plugin command reference
    └── ...                        # One file per resource command group
```

### Mandatory Documentation Rules

#### Rule 1: New Command → New/Updated User Guide
When adding a new command (e.g., `a6 upstream list`):
- Create or update `docs/user-guide/<resource>.md`
- Include command syntax, all flags with descriptions, examples, and common use cases
- Format: every command section must have Synopsis, Description, Flags table, and Examples

#### Rule 2: New Flag → Update Command Reference
When adding a new flag to an existing command:
- Update the flags table in the corresponding user guide
- Add an example showing the flag in use

#### Rule 3: Behavior Change → Update Affected Docs
When changing command behavior (output format, error messages, defaults):
- Update all affected user guide pages
- If the change affects the golden example pattern, update `docs/golden-example.md`

#### Rule 4: API Client Change → Update Admin API Spec
When the APISIX Admin API changes (new field, endpoint, parameter):
- Update `docs/admin-api-spec.md`
- This should be rare, usually only when tracking a new APISIX version

#### Rule 5: Architecture Change → New ADR
When making a significant architectural decision:
- Create `docs/adr/NNN-<title>.md` following the ADR format
- Link from AGENTS.md

### User Guide Page Template
Resource command reference pages follow this template:

```markdown
# <Resource> Commands

Manage APISIX <resources>.

## a6 <resource> list

List all <resources>.

### Synopsis
\`\`\`
a6 <resource> list [flags]
\`\`\`

### Flags
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| --output | -o | string | (auto) | Output format: table, json, yaml |
| --page | | int | 1 | Page number |
| --page-size | | int | 10 | Items per page (10-500) |
| --name | | string | | Filter by name |
| --label | | string | | Filter by label |

### Examples
\`\`\`bash
# List all routes
a6 route list

# List routes as JSON
a6 route list -o json

# List routes filtered by name
a6 route list --name "users-api"

# List with custom page size
a6 route list --page-size 50
\`\`\`

## a6 <resource> get

Get a specific <resource> by ID.

[... same pattern for get, create, update, delete ...]
```

### Documentation Quality Checklist
Before approving any PR, verify:
- [ ] All new commands have user guide entries
- [ ] All flags are documented with types and defaults
- [ ] At least 2 examples per command
- [ ] Examples are realistic and tested
- [ ] No broken internal links
- [ ] AGENTS.md document map is up to date

### Who Updates Documentation
- AI coding agents: MUST update docs as part of every feature PR
- Human reviewers: MUST check docs checklist before approving
- If docs are missing from a code PR, the PR is NOT ready for merge
