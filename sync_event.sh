#!/bin/bash
exec "$(cd "$(dirname "$0")" && pwd)/utils/sync_event.sh" "$@"
