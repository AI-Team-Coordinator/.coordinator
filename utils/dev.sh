#!/bin/bash
# Alias of ./utils/run.sh — API + Vite (Alina Assist) or static board (installs).

exec "$(cd "$(dirname "$0")" && pwd)/run.sh" "$@"
