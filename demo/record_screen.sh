#!/usr/bin/env bash
set -euo pipefail

rm -f /out/grafview-demo.mp4
export DISPLAY=:99
Xvfb "$DISPLAY" -screen 0 1440x900x24 -nolisten tcp 2>/tmp/xvfb.log &
XVFB_PID=$!
sleep 1

ffmpeg -y -hide_banner -loglevel error \
  -video_size 1440x900 \
  -framerate 30 \
  -f x11grab \
  -i "${DISPLAY}.0" \
  -codec:v libx264 \
  -pix_fmt yuv420p \
  -movflags +faststart \
  /out/grafview-demo.mp4 &
FFMPEG_PID=$!

stop_ffmpeg() {
  kill -INT "$FFMPEG_PID" >/dev/null 2>&1 || true
  wait "$FFMPEG_PID" >/dev/null 2>&1 || true
  kill "$XVFB_PID" >/dev/null 2>&1 || true
  wait "$XVFB_PID" >/dev/null 2>&1 || true
}
trap stop_ffmpeg EXIT

node /runner/record_demo.mjs
sleep 1
