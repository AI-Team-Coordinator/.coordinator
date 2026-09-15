#!/bin/bash
# Print the first free TCP listen port at or above START (default 4321).
# Usage: ./utils/free_port.sh [START]

set -e
START="${1:-4321}"
if ! [[ "$START" =~ ^[0-9]+$ ]]; then
    echo "usage: $0 [port]" >&2
    exit 1
fi

end=$((START + 50))
p=$START
while [ "$p" -lt "$end" ]; do
    if ! lsof -nP -tiTCP:"$p" -sTCP:LISTEN >/dev/null 2>&1; then
        echo "$p"
        exit 0
    fi
    p=$((p + 1))
done
echo "no free TCP port in $START–$((end - 1))" >&2
exit 1
