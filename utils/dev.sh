#!/bin/bash
# Alias of ./utils/run.sh — Vite HMR + detached API.

exec "$(cd "$(dirname "$0")" && pwd)/run.sh" "$@"
