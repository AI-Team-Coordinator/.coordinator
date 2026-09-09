#!/usr/bin/env python3
"""Start a process in a new session so it survives the Cursor chat shell."""

from __future__ import annotations

import os
import subprocess
import sys


def main() -> None:
    if len(sys.argv) < 5:
        raise SystemExit("usage: daemonize.py PIDFILE LOG CWD -- CMD...")
    pidfile, log_path, cwd = sys.argv[1], sys.argv[2], sys.argv[3]
    try:
        sep = sys.argv.index("--")
    except ValueError:
        raise SystemExit("usage: daemonize.py PIDFILE LOG CWD -- CMD...")
    cmd = sys.argv[sep + 1 :]
    if not cmd:
        raise SystemExit("missing command")
    os.makedirs(os.path.dirname(os.path.abspath(pidfile)), exist_ok=True)
    os.makedirs(os.path.dirname(os.path.abspath(log_path)), exist_ok=True)
    if cwd:
        os.chdir(cwd)
    with open(log_path, "ab") as log:
        proc = subprocess.Popen(
            cmd,
            stdout=log,
            stderr=subprocess.STDOUT,
            stdin=subprocess.DEVNULL,
            start_new_session=True,
            close_fds=True,
        )
    with open(pidfile, "w", encoding="utf-8") as f:
        f.write(str(proc.pid) + "\n")
    print(proc.pid)


if __name__ == "__main__":
    main()
