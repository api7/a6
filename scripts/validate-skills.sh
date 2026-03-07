#!/usr/bin/env bash
#
# validate-skills.sh — CI validation for SKILL.md files
#
# Checks:
#   1. Every skills/<name>/SKILL.md has valid YAML frontmatter
#   2. Required fields: name, description
#   3. name matches directory name
#   4. name follows kebab-case: ^[a-z0-9]+(-[a-z0-9]+)*$
#   5. description is non-empty
#
# Usage:
#   ./scripts/validate-skills.sh
#
# Exit codes:
#   0 — all skills valid
#   1 — one or more validation errors

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SKILLS_DIR="$PROJECT_ROOT/skills"

errors=0

# Color output (if terminal)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
fi

log_error() {
    echo -e "${RED}ERROR${NC}: $1" >&2
    errors=$((errors + 1))
}

log_ok() {
    echo -e "${GREEN}OK${NC}: $1"
}

log_info() {
    echo -e "${YELLOW}INFO${NC}: $1"
}

# Check skills directory exists
if [ ! -d "$SKILLS_DIR" ]; then
    log_error "skills/ directory not found at $SKILLS_DIR"
    exit 1
fi

# Find all SKILL.md files
skill_files=$(find "$SKILLS_DIR" -mindepth 2 -maxdepth 2 -name "SKILL.md" -type f | sort)

if [ -z "$skill_files" ]; then
    log_error "no SKILL.md files found in skills/*/"
    exit 1
fi

skill_count=0

for skill_file in $skill_files; do
    skill_count=$((skill_count + 1))
    dir_name=$(basename "$(dirname "$skill_file")")
    rel_path="skills/$dir_name/SKILL.md"

    log_info "Validating $rel_path"

    # Check file is non-empty
    if [ ! -s "$skill_file" ]; then
        log_error "$rel_path: file is empty"
        continue
    fi

    # Extract frontmatter (content between first two --- lines)
    frontmatter=$(awk '/^---$/{if(++n==2)exit}n==1{print}' "$skill_file")

    if [ -z "$frontmatter" ]; then
        log_error "$rel_path: no YAML frontmatter found (must start with --- and end with ---)"
        continue
    fi

    # Extract 'name' field from frontmatter
    # Handles: name: value, name: "value", name: 'value'
    name=$(echo "$frontmatter" | grep -E '^name:' | head -1 | sed 's/^name:[[:space:]]*//' | sed 's/^["'\'']//' | sed 's/["'\'']$//' | tr -d '\r')

    if [ -z "$name" ]; then
        log_error "$rel_path: missing required field 'name'"
        continue
    fi

    # Validate name matches directory name
    if [ "$name" != "$dir_name" ]; then
        log_error "$rel_path: name '$name' does not match directory name '$dir_name'"
    fi

    # Validate name follows kebab-case
    if ! echo "$name" | grep -qE '^[a-z0-9]+(-[a-z0-9]+)*$'; then
        log_error "$rel_path: name '$name' does not follow kebab-case pattern (^[a-z0-9]+(-[a-z0-9]+)*$)"
    fi

    # Extract 'description' field from frontmatter
    # Handle both single-line and multi-line (YAML block scalar) descriptions
    description=$(echo "$frontmatter" | awk '
        /^description:/ {
            # Remove "description:" prefix
            sub(/^description:[[:space:]]*/, "")
            # If line has content after "description:", check for block scalar indicators
            if ($0 ~ /^[>|]/) {
                # Multi-line block scalar — read next lines
                while (getline > 0) {
                    if ($0 ~ /^[a-zA-Z]/) break  # Next top-level key
                    gsub(/^[[:space:]]+/, "")
                    desc = desc $0 " "
                }
                print desc
            } else if ($0 != "") {
                # Single-line value
                gsub(/^["'\''"]/, "", $0)
                gsub(/["'\''"]$/, "", $0)
                print $0
            } else {
                print ""
            }
            exit
        }
    ')

    if [ -z "$description" ]; then
        log_error "$rel_path: missing or empty required field 'description'"
    fi

    # If no errors for this file, log success
    if [ $errors -eq 0 ] || true; then
        log_ok "$rel_path: name=$name"
    fi
done

echo ""
echo "Validated $skill_count skill(s)."

if [ $errors -gt 0 ]; then
    echo -e "${RED}Found $errors error(s).${NC}"
    exit 1
fi

echo -e "${GREEN}All skills valid.${NC}"
exit 0
