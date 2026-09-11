## Summary

<!-- What changed and why. -->

## Agent skills checklist

The agent skills for this CLI live in [api7/agent-skills](https://github.com/api7/agent-skills). CI runs `test/skills` against that repository's `main`, so keep the two in step:

- [ ] This PR **adds** a command, flag, or plugin → merge this PR first, then open the reference update in api7/agent-skills.
- [ ] This PR **removes or renames** a command or flag → merge the api7/agent-skills PR that stops using it first, then this PR.
- [ ] No CLI surface change → nothing to do in api7/agent-skills.
