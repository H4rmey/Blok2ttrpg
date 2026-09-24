#!/usr/bin/env bash
# Stop the Blok2 TTRPG server started by ./start.sh.
#
# The pid recorded by start.sh is used when available. As a fallback (for a
# server started by hand) any process listening on the project port is
# terminated, so the port is always released.
#
# Usage:
#   ./stop.sh            stop the server on the default port
#   PORT=9000 ./stop.sh  stop a server started on another port
set -uo pipefail

PORT="${PORT:-47821}"
RUN_DIR=".run"
PID_FILE="$RUN_DIR/server.pid"

cd "$(dirname "$0")"

stopped=0

if [ -f "$PID_FILE" ]; then
    pid="$(cat "$PID_FILE")"
    if kill -0 "$pid" 2>/dev/null; then
        echo "Stopping server (pid $pid)..."
        kill "$pid"
        # Allow a graceful exit before escalating to SIGKILL.
        for _ in 1 2 3 4 5; do
            kill -0 "$pid" 2>/dev/null || break
            sleep 1
        done
        if kill -0 "$pid" 2>/dev/null; then
            echo "Still running; sending SIGKILL."
            kill -9 "$pid" 2>/dev/null
        fi
        stopped=1
    else
        echo "Stale pid file (pid $pid is not running)."
    fi
    rm -f "$PID_FILE"
fi

# Fallback: catch a server that was started without start.sh and is still
# holding the port. lsof/fuser are used if present; absence of both is not an
# error, it just means we cannot look the process up by port.
if [ "$stopped" -eq 0 ]; then
    if command -v lsof >/dev/null 2>&1; then
        pids="$(lsof -ti ":$PORT" 2>/dev/null)"
        if [ -n "$pids" ]; then
            echo "Stopping process(es) on port $PORT: $pids"
            kill $pids 2>/dev/null
            stopped=1
        fi
    elif command -v fuser >/dev/null 2>&1; then
        if fuser -k "$PORT/tcp" >/dev/null 2>&1; then
            echo "Stopping process on port $PORT."
            stopped=1
        fi
    fi
fi

if [ "$stopped" -eq 0 ]; then
    echo "No running server found on port $PORT."
else
    echo "Stopped."
fi
