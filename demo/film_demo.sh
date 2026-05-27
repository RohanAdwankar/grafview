#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="$ROOT/demo/out"
DASHBOARD_DIR="${1:-}"
PORT="${GRAFVIEW_DEMO_PORT:-3147}"
MOCK_PORT="${GRAFVIEW_DEMO_MOCK_PORT:-3148}"
PLAYWRIGHT_BASE_IMAGE="${PLAYWRIGHT_BASE_IMAGE:-mcr.microsoft.com/playwright:v1.56.1-noble}"
PLAYWRIGHT_VERSION="${PLAYWRIGHT_VERSION:-1.56.1}"
RECORDER_IMAGE="${GRAFVIEW_DEMO_RECORDER_IMAGE:-grafview-demo-recorder:latest}"
CONTAINER_NAME="${GRAFVIEW_DEMO_CONTAINER:-grafview-demo}"
RECORDER_CONTAINER="${GRAFVIEW_DEMO_RECORDER_CONTAINER:-grafview-demo-recorder-run}"
TMP=""

mkdir -p "$OUT"

write_demo_dashboard() {
  local path="$1"
  local title="$2"
  local trend_title="$3"
  local stat_title="$4"
  local query="$5"
  mkdir -p "$(dirname "$path")"
  cat >"$path" <<JSON
{
  "title": "$title",
  "schemaVersion": 39,
  "refresh": "5s",
  "time": { "from": "now-6h", "to": "now" },
  "panels": [
    {
      "type": "timeseries",
      "title": "$trend_title",
      "gridPos": { "x": 0, "y": 0, "w": 12, "h": 8 },
      "datasource": { "type": "prometheus", "uid": "prometheus" },
      "targets": [{ "expr": "$query", "refId": "A" }]
    },
    {
      "type": "stat",
      "title": "$stat_title",
      "gridPos": { "x": 12, "y": 0, "w": 6, "h": 8 },
      "datasource": { "type": "prometheus", "uid": "prometheus" },
      "targets": [{ "expr": "${query}_current", "refId": "A" }]
    },
    {
      "type": "logs",
      "title": "Recent logs",
      "gridPos": { "x": 0, "y": 8, "w": 18, "h": 8 },
      "datasource": { "type": "loki", "uid": "loki" },
      "targets": [{ "expr": "{job=\"demo\"}", "refId": "A" }]
    }
  ]
}
JSON
}

if [[ -z "$DASHBOARD_DIR" ]]; then
  TMP="$(mktemp -d)"
  DASHBOARD_DIR="$TMP/dashboard"
  write_demo_dashboard "$DASHBOARD_DIR/sre/cluster-health.json" "SRE Cluster Health" "Node health trend" "Healthy nodes" "sre_cluster_health_score"
  write_demo_dashboard "$DASHBOARD_DIR/sre/incident-response.json" "SRE Incident Response" "Alert volume" "Open incidents" "sre_alerts_total"
  write_demo_dashboard "$DASHBOARD_DIR/customer/financial-kpis.json" "Customer Financial KPIs" "Monthly spend" "Forecast spend" "customer_monthly_spend"
  write_demo_dashboard "$DASHBOARD_DIR/customer/usage-kpis.json" "Customer Usage KPIs" "GPU hours" "Active projects" "customer_gpu_hours"
fi

DASHBOARD_DIR="$(cd "$DASHBOARD_DIR" && pwd)"
LOG="$OUT/grafview-demo.log"
BIN="$OUT/grafview-demo-bin"
MP4="$OUT/grafview-demo.mp4"
GIF="$OUT/grafview-demo.gif"

cleanup() {
  local code=$?
  if [[ -n "${GRAFVIEW_PID:-}" ]]; then
    kill "$GRAFVIEW_PID" >/dev/null 2>&1 || true
    wait "$GRAFVIEW_PID" >/dev/null 2>&1 || true
  fi
  docker rm -f "$CONTAINER_NAME" "$RECORDER_CONTAINER" >/dev/null 2>&1 || true
  [[ -n "$TMP" ]] && rm -rf "$TMP"
  exit "$code"
}
trap cleanup EXIT

cd "$ROOT"
go build -o "$BIN" ./cmd/grafview
rm -f "$MP4" "$GIF"
rm -rf "$OUT/raw-video"
docker build \
  --quiet \
  --build-arg PLAYWRIGHT_BASE_IMAGE="$PLAYWRIGHT_BASE_IMAGE" \
  --build-arg PLAYWRIGHT_VERSION="$PLAYWRIGHT_VERSION" \
  -t "$RECORDER_IMAGE" \
  -f "$ROOT/demo/Dockerfile" \
  "$ROOT/demo" >/dev/null
docker rm -f "$CONTAINER_NAME" "$RECORDER_CONTAINER" >/dev/null 2>&1 || true
"$BIN" -port "$PORT" -mock-port "$MOCK_PORT" -name "$CONTAINER_NAME" -open=false "$DASHBOARD_DIR" >"$LOG" 2>&1 &
GRAFVIEW_PID=$!

for _ in $(seq 1 90); do
  if curl -fsS "http://127.0.0.1:$PORT/api/health" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
curl -fsS "http://127.0.0.1:$PORT/api/health" >/dev/null

docker run --rm \
  --name "$RECORDER_CONTAINER" \
  --ipc=host \
  --add-host=host.docker.internal:host-gateway \
  -e GRAFANA_URL="http://host.docker.internal:$PORT" \
  -e DASHBOARD_FILE_URL="file:///Users/admin/dashboard/" \
  -v "$ROOT:/work:ro" \
  -v "$DASHBOARD_DIR:/Users/admin/dashboard:ro" \
  -v "$OUT:/out" \
  "$RECORDER_IMAGE" \
  bash -lc 'cp /work/demo/record_demo.mjs /runner/record_demo.mjs && bash /runner/record_screen.sh'

test -s "$MP4"
test -s "$GIF"
printf "%s\n%s\n" "$MP4" "$GIF"
