#!/bin/sh
# Install the Apache APISIX AI agent skills into your AI coding agent.
#
# Each skill is a SKILL.md knowledge pack that teaches an agent (Claude Code,
# Cursor, Copilot, Windsurf, OpenCode, ...) how to configure Apache APISIX
# through the a6 CLI. This script copies them into your agent's skills directory.
#
# Quick start (installs into ~/.claude/skills for Claude Code):
#   curl -fsSL https://raw.githubusercontent.com/api7/a6/main/install.sh | sh
#
# Install somewhere else (e.g. a project-local Cursor rules dir):
#   curl -fsSL https://raw.githubusercontent.com/api7/a6/main/install.sh | sh -s -- --dir ./.cursor/rules
#   SKILLS_DIR=~/.config/opencode/skills sh -c "$(curl -fsSL https://raw.githubusercontent.com/api7/a6/main/install.sh)"
set -eu

REPO="api7/a6"
BRANCH="main"
LABEL="Apache APISIX"

# Target directory. Default: Claude Code personal skills. Override with
# SKILLS_DIR=... or --dir <path>.
SKILLS_DIR="${SKILLS_DIR:-$HOME/.claude/skills}"

while [ $# -gt 0 ]; do
  case "$1" in
    --dir) SKILLS_DIR="${2:?--dir needs a path}"; shift 2 ;;
    --dir=*)
      SKILLS_DIR="${1#*=}"
      [ -n "$SKILLS_DIR" ] || { echo "install.sh: --dir needs a path" >&2; exit 1; }
      shift
      ;;
    -h | --help)
      echo "Usage: install.sh [--dir <path>]   (default: \$HOME/.claude/skills)"
      exit 0
      ;;
    *) echo "install.sh: unknown option '$1'" >&2; exit 1 ;;
  esac
done

if command -v curl >/dev/null 2>&1; then
  fetch() { curl -fsSL "$1"; }
elif command -v wget >/dev/null 2>&1; then
  fetch() { wget -qO- "$1"; }
else
  echo "install.sh: need curl or wget on your PATH." >&2
  exit 1
fi

echo "Installing ${LABEL} agent skills into ${SKILLS_DIR} ..."

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT INT TERM

# Download the repo tarball (no git required) and extract just the skills.
fetch "https://codeload.github.com/${REPO}/tar.gz/refs/heads/${BRANCH}" >"$TMP/repo.tgz" ||
  { echo "install.sh: download failed." >&2; exit 1; }
tar -xzf "$TMP/repo.tgz" -C "$TMP"

SRC="$(find "$TMP" -maxdepth 2 -type d -name skills | head -n 1)"
if [ -z "$SRC" ] || [ ! -d "$SRC" ]; then
  echo "install.sh: could not find a skills/ directory in the download." >&2
  exit 1
fi

mkdir -p "$SKILLS_DIR"
count=0
for dir in "$SRC"/*/; do
  [ -f "${dir}SKILL.md" ] || continue
  name="$(basename "$dir")"
  rm -rf "${SKILLS_DIR}/${name}"
  cp -R "$dir" "${SKILLS_DIR}/${name}"
  count=$((count + 1))
done

if [ "$count" -eq 0 ]; then
  echo "install.sh: no SKILL.md packs found to install." >&2
  exit 1
fi

echo "Installed ${count} skills to ${SKILLS_DIR}"
echo
echo "Next: ask your AI coding agent to configure ${LABEL} in plain language, e.g."
echo "  \"add key-auth to my /orders route and rate-limit it to 100 requests per minute\""
echo
echo "Browse the catalog: https://docs.api7.ai/apisix/ai-agent-skills"
