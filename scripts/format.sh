#!/usr/bin/env bash
set -euo pipefail

echo "Running gofmt -w on repository..."
gofmt -w .

if command -v goimports >/dev/null 2>&1; then
  echo "Running goimports -w (found in PATH)..."
  goimports -w .
fi

echo "Files still needing formatting (should be none):"
gofmt -l . || true

echo "Done. If files were changed, review and commit them:
  git add -A
  git commit -m 'ci: format with gofmt/goimports'
"
