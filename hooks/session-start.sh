#!/bin/sh
# NeetoKB CLI — session-start hook for Claude Code
# Lightweight auth liveness check. Always exits 0 (informational).

if ! command -v neetokb >/dev/null 2>&1; then
  echo "NeetoKB CLI is not installed or not on PATH."
  exit 0
fi

if neetokb whoami >/dev/null 2>&1; then
  echo "NeetoKB plugin active."
else
  echo "NeetoKB CLI installed but not authenticated. Run 'neetokb login' to authenticate."
fi

exit 0
