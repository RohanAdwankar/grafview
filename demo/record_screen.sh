#!/usr/bin/env bash
set -euo pipefail

RAW=/out/grafview-demo-raw.mp4
MP4=/out/grafview-demo.mp4
GIF=/out/grafview-demo.gif
PALETTE=/tmp/grafview-demo-palette.png

rm -f "$RAW" "$MP4" "$GIF" "$PALETTE"
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
  "$RAW" &
FFMPEG_PID=$!

stop_capture() {
  kill -INT "$FFMPEG_PID" >/dev/null 2>&1 || true
  wait "$FFMPEG_PID" >/dev/null 2>&1 || true
}

cleanup() {
  stop_capture
  kill "$XVFB_PID" >/dev/null 2>&1 || true
  wait "$XVFB_PID" >/dev/null 2>&1 || true
}
trap cleanup EXIT

node /runner/record_demo.mjs
sleep 1
stop_capture

duration="$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$RAW")"
trimmed="$(awk -v d="$duration" 'BEGIN { t=d-2; if (t < 1) t=1; printf "%.3f", t }')"
ffmpeg -y -hide_banner -loglevel error \
  -ss 1 \
  -t "$trimmed" \
  -i "$RAW" \
  -codec:v libx264 \
  -pix_fmt yuv420p \
  -movflags +faststart \
  "$MP4"

ffmpeg -y -hide_banner -loglevel error \
  -i "$MP4" \
  -vf "fps=10,scale=960:-1:flags=lanczos,palettegen" \
  "$PALETTE"
ffmpeg -y -hide_banner -loglevel error \
  -i "$MP4" \
  -i "$PALETTE" \
  -lavfi "fps=10,scale=960:-1:flags=lanczos[x];[x][1:v]paletteuse" \
  "$GIF"
