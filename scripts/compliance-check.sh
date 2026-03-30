#!/bin/bash
# Compliance vocabulary scanner (ANALYSIS.md 6.10)
# Scans staged files for forbidden insurance-related terms.
# Runs as Git pre-commit hook and in CI pipeline.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Forbidden patterns (Chinese + English)
# Note: "险" alone is too broad (appears in "风险" which is acceptable).
# We check compound forms and the specific standalone terms from 6.10.
FORBIDDEN_ZH="保险|保费|理赔|承保|投保"
FORBIDDEN_EN="insurance|premium|underwrite|insure[^_]"

# Allowlist: files that legitimately discuss compliance rules
ALLOWLIST=(
  "shared/compliance-terms.json"
  "scripts/compliance-check.sh"
  "docs/ANALYSIS.md"
  "docs/PRD.md"
  ".claude/plans/"
)

build_exclude_args() {
  local args=""
  for pattern in "${ALLOWLIST[@]}"; do
    args="$args --exclude=$pattern"
  done
  echo "$args"
}

EXCLUDE_ARGS=$(build_exclude_args)

echo "Running compliance scan..."

VIOLATIONS=0

# Check staged files if in git context, otherwise check all tracked files
if git rev-parse --git-dir > /dev/null 2>&1; then
  # Get list of staged files (for pre-commit) or all tracked files (for CI)
  if [ "${CI:-}" = "true" ]; then
    FILES=$(git ls-files -- '*.go' '*.ts' '*.tsx' '*.js' '*.jsx' '*.json' '*.yaml' '*.yml' '*.md' | grep -v -E "$(IFS='|'; echo "${ALLOWLIST[*]}")" || true)
  else
    FILES=$(git diff --cached --name-only --diff-filter=ACM -- '*.go' '*.ts' '*.tsx' '*.js' '*.jsx' '*.json' '*.yaml' '*.yml' | grep -v -E "$(IFS='|'; echo "${ALLOWLIST[*]}")" || true)
  fi
else
  echo "Not a git repository. Scanning all files..."
  FILES=$(find "$PROJECT_ROOT" -type f \( -name "*.go" -o -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.json" \) | grep -v node_modules | grep -v .git | grep -v -E "$(IFS='|'; echo "${ALLOWLIST[*]}")" || true)
fi

if [ -z "$FILES" ]; then
  echo "No files to scan."
  exit 0
fi

# Scan for Chinese forbidden terms
for file in $FILES; do
  if [ -f "$PROJECT_ROOT/$file" ] 2>/dev/null || [ -f "$file" ]; then
    TARGET="${PROJECT_ROOT}/${file}"
    [ -f "$file" ] && TARGET="$file"

    if grep -nP "$FORBIDDEN_ZH" "$TARGET" 2>/dev/null; then
      echo "  VIOLATION in: $file (Chinese forbidden term)"
      VIOLATIONS=$((VIOLATIONS + 1))
    fi

    if grep -niP "$FORBIDDEN_EN" "$TARGET" 2>/dev/null; then
      echo "  VIOLATION in: $file (English forbidden term)"
      VIOLATIONS=$((VIOLATIONS + 1))
    fi
  fi
done

if [ "$VIOLATIONS" -gt 0 ]; then
  echo ""
  echo "BLOCKED: $VIOLATIONS file(s) contain forbidden insurance terminology."
  echo "See shared/compliance-terms.json for allowed alternatives."
  echo "Allowlisted files: ${ALLOWLIST[*]}"
  exit 1
else
  echo "Compliance scan passed. No forbidden terms found."
  exit 0
fi
