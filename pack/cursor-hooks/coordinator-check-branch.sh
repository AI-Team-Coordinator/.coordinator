#!/bin/bash
# PreToolUse wrapper — JSON on stdin/stdout.
DIR=$(cd "$(dirname "$0")" && pwd)
exec python3 "$DIR/coordinator-check-branch.py"
