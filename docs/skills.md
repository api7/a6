# AI Agent Skills

The `a6` agent skill teaches AI coding agents (Claude Code, Cursor, Codex,
GitHub Copilot, Windsurf, OpenCode and others) how to configure and operate
Apache APISIX through the a6 CLI: routes, services, upstreams, consumers,
SSL, 29 plugins, 8 operational recipes, and developer/operator personas.

## Where it lives

The skill content is maintained in the dedicated
[api7/agent-skills](https://github.com/api7/agent-skills) repository and
published at [skills.sh/api7/agent-skills/a6](https://skills.sh/api7/agent-skills/a6).
It is no longer stored in this repository.

`skills/a6/SKILL.md` is a short router; detailed guidance lives under
`skills/a6/references/` (`shared.md`, `plugins/`, `recipes/`, `personas/`) and
is loaded by the agent only when a task needs it.

## Install

```bash
# install the a6 skill into the current project
npx skills add api7/agent-skills --skill a6

# target a specific agent, e.g. claude-code, cursor, codex, github-copilot
npx skills add api7/agent-skills --skill a6 -a claude-code

# install globally (for every project) instead of into the current one
npx skills add api7/agent-skills --skill a6 -g
```

Update later with `npx skills update`. Without Node, `install.sh` in this
repository copies the skill into a directory of your choice
(default `~/.claude/skills/a6`).

Installing copies instructions only. It does not install `a6`, connect to
APISIX, or run any command; you still need `a6` on your `PATH` and a
reachable Admin API.

## Operating discipline

Use a non-production APISIX instance for a first run. Ask the agent to
inspect the current resources, propose an exact change, wait for approval,
apply only the approved change, verify the result, and keep a rollback path.
Never put an Admin API key in a prompt or a committed file; configure it
through `a6 context` or the `A6_API_KEY` environment variable instead.

## Contributing

Changes to skill content (new plugins, recipes, wording fixes) go to
[api7/agent-skills](https://github.com/api7/agent-skills). This repository
only validates that the shell examples in the skill use commands and flags
that exist in the current `a6` CLI: `make test-skills` runs `test/skills`
against a checkout of api7/agent-skills next to this repository, or against
the directory given by `SKILLS_DIR` (CI checks out the repository and sets
`SKILLS_DIR` automatically).
