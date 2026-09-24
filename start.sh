#!/usr/bin/env bash
# Start the Blok2 TTRPG server in the background on the project's fixed port.
#
# The port is intentionally hardcoded: 47821 is the canonical development port
# for this project, so every contributor and script hits the same URL. Set
# PORT in the environment to override it for a one-off run.
#
# Usage:
#   ./start.sh            start the server (logs to .run/server.log)
#   PORT=9000 ./start.sh  start on a different port
set -euo pipefail

PORT="${PORT:-47821}"
RUN_DIR=".run"
PID_FILE="$RUN_DIR/server.pid"
LOG_FILE="$RUN_DIR/server.log"

cd "$(dirname "$0")"
mkdir -p "$RUN_DIR"

# Refuse to start a second instance: an existing live pid means the server is
# already running and starting again would just fail to bind the port.
if [ -f "$PID_FILE" ]; then
    old_pid="$(cat "$PID_FILE")"
    if kill -0 "$old_pid" 2>/dev/null; then
        echo "Server already running (pid $old_pid) on port $PORT."
        echo "Use ./stop.sh first if you want to restart it."
        exit 1
    fi
    # Stale pid file from a crashed or killed run.
    rm -f "$PID_FILE"
fi

echo "Building..."
go build -o "$RUN_DIR/blok2ttrpg" .

echo "Starting on http://localhost:$PORT"
PORT="$PORT" "$RUN_DIR/blok2ttrpg" "$@" >"$LOG_FILE" 2>&1 &
echo $! >"$PID_FILE"

# Give the process a moment to bind so an immediate failure (port in use, bad
# config) is reported here instead of silently in the log.
sleep 1
if ! kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
    echo "Server failed to start. Last log lines:"
    tail -n 20 "$LOG_FILE"
    rm -f "$PID_FILE"
    exit 1
fi

echo "Started (pid $(cat "$PID_FILE")). Logs: $LOG_FILE"
