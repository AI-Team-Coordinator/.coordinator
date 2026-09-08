#!/bin/bash
exec "$(cd "$(dirname "$0")" && pwd)/utils/log_event.sh" "$@"
