package grafview

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type mockDataServer struct {
	Port   int
	server *http.Server
}

func startMockDataServer(port int) (*mockDataServer, error) {
	addr := "127.0.0.1:0"
	if port > 0 {
		addr = fmt.Sprintf("127.0.0.1:%d", port)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	actual := ln.Addr().(*net.TCPAddr).Port
	m := &mockDataServer{Port: actual}
	m.server = &http.Server{Handler: m}
	go func() { _ = m.server.Serve(ln) }()
	return m, nil
}

func (m *mockDataServer) Close(ctx context.Context) error {
	return m.server.Shutdown(ctx)
}

func (m *mockDataServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	switch {
	case r.URL.Path == "/ready":
		_, _ = w.Write([]byte("ready\n"))
	case strings.HasPrefix(r.URL.Path, "/loki/api/v1/"):
		m.handleLoki(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/v1/"):
		m.handlePrometheus(w, r)
	default:
		writeJSON(w, http.StatusOK, map[string]any{"status": "success"})
	}
}

func (m *mockDataServer) handlePrometheus(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/api/v1/query_range":
		writeJSON(w, http.StatusOK, promSuccess("matrix", promMatrix(r)))
	case r.URL.Path == "/api/v1/query":
		writeJSON(w, http.StatusOK, promSuccess("vector", promVector(r.URL.Query().Get("query"))))
	case r.URL.Path == "/api/v1/labels":
		writeJSON(w, http.StatusOK, map[string]any{"status": "success", "data": labelNames()})
	case strings.HasPrefix(r.URL.Path, "/api/v1/label/") && strings.HasSuffix(r.URL.Path, "/values"):
		label := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/label/"), "/values")
		writeJSON(w, http.StatusOK, map[string]any{"status": "success", "data": labelValues(label)})
	case r.URL.Path == "/api/v1/series":
		writeJSON(w, http.StatusOK, map[string]any{"status": "success", "data": []map[string]string{
			{"__name__": "mock_metric", "instance": "node-01", "rack": "rack-a", "job": "mock"},
			{"__name__": "mock_metric", "instance": "node-02", "rack": "rack-b", "job": "mock"},
		}})
	case r.URL.Path == "/api/v1/metadata":
		writeJSON(w, http.StatusOK, map[string]any{"status": "success", "data": map[string]any{}})
	case r.URL.Path == "/api/v1/status/buildinfo":
		writeJSON(w, http.StatusOK, map[string]any{"status": "success", "data": map[string]any{"version": "mock"}})
	default:
		writeJSON(w, http.StatusOK, promSuccess("vector", promVector(r.URL.Query().Get("query"))))
	}
}

func promSuccess(resultType string, result any) map[string]any {
	return map[string]any{
		"status": "success",
		"data": map[string]any{
			"resultType": resultType,
			"result":     result,
		},
	}
}

func promMatrix(r *http.Request) []map[string]any {
	q := r.URL.Query()
	end := parseUnix(q.Get("end"), time.Now().Unix())
	start := parseUnix(q.Get("start"), end-6*3600)
	step := parseDurationSeconds(q.Get("step"), 60)
	if step <= 0 {
		step = 60
	}
	if points := (end - start) / step; points > 240 {
		step = (end - start) / 240
		if step < 1 {
			step = 1
		}
	}
	query := q.Get("query")
	out := make([]map[string]any, 0, 2)
	for series := 0; series < 2; series++ {
		values := make([][]any, 0, 128)
		for ts := start; ts <= end; ts += step {
			values = append(values, []any{float64(ts), fmt.Sprintf("%.3f", sample(query, series, ts))})
		}
		out = append(out, map[string]any{
			"metric": metricLabels(query, series),
			"values": values,
		})
	}
	return out
}

func promVector(query string) []map[string]any {
	now := time.Now().Unix()
	return []map[string]any{
		{"metric": metricLabels(query, 0), "value": []any{float64(now), fmt.Sprintf("%.3f", sample(query, 0, now))}},
		{"metric": metricLabels(query, 1), "value": []any{float64(now), fmt.Sprintf("%.3f", sample(query, 1, now))}},
	}
}

func metricLabels(query string, series int) map[string]string {
	return map[string]string{
		"__name__":  "mock_metric",
		"instance":  fmt.Sprintf("node-%02d", series+1),
		"rack":      fmt.Sprintf("rack-%c", 'a'+series),
		"job":       "mock",
		"partition": "slurm/default",
		"user":      fmt.Sprintf("user%d", series+1),
		"job_name":  fmt.Sprintf("job-%d", series+1),
		"query":     shortQuery(query),
	}
}

func (m *mockDataServer) handleLoki(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/loki/api/v1/query_range", r.URL.Path == "/loki/api/v1/query":
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "success",
			"data": map[string]any{
				"resultType": "streams",
				"result":     lokiStreams(r.URL.Query().Get("query")),
			},
		})
	case r.URL.Path == "/loki/api/v1/labels":
		writeJSON(w, http.StatusOK, map[string]any{"status": "success", "data": []string{"job", "filename", "level", "node"}})
	case strings.HasPrefix(r.URL.Path, "/loki/api/v1/label/") && strings.HasSuffix(r.URL.Path, "/values"):
		label := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/loki/api/v1/label/"), "/values")
		writeJSON(w, http.StatusOK, map[string]any{"status": "success", "data": labelValues(label)})
	default:
		writeJSON(w, http.StatusOK, map[string]any{"status": "success", "data": []any{}})
	}
}

func lokiStreams(query string) []map[string]any {
	now := time.Now()
	values := make([][]string, 0, 8)
	for i := 0; i < 8; i++ {
		ts := now.Add(time.Duration(-i) * time.Minute).UnixNano()
		values = append(values, []string{
			strconv.FormatInt(ts, 10),
			fmt.Sprintf("mock log line %d for %s", i+1, shortQuery(query)),
		})
	}
	return []map[string]any{{
		"stream": map[string]string{"job": "mock", "filename": "/var/log/mock.log", "level": "info", "node": "node-01"},
		"values": values,
	}}
}

func labelNames() []string {
	return []string{"rack", "name", "partition", "user", "job_name", "state", "instance", "node", "job", "device", "gpu", "namespace", "pod"}
}

func labelValues(label string) []string {
	switch strings.Trim(label, "/") {
	case "rack":
		return []string{"rack-a", "rack-b"}
	case "name", "partition":
		return []string{"slurm/default", "slurm/gpu"}
	case "user":
		return []string{"user1", "user2"}
	case "job_name":
		return []string{"job-1", "job-2"}
	case "state":
		return []string{"idle", "allocated", "down", "draining"}
	case "filename":
		return []string{"/var/log/syslog", "/var/log/slurm/slurmctld.log", "/var/log/mock.log"}
	case "level":
		return []string{"info", "warn", "error"}
	default:
		return []string{"mock-a", "mock-b"}
	}
}

func parseUnix(s string, def int64) int64 {
	if s == "" {
		return def
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return int64(f)
}

func parseDurationSeconds(s string, def int64) int64 {
	if s == "" {
		return def
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int64(f)
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return def
	}
	return int64(d.Seconds())
}

func sample(query string, series int, ts int64) float64 {
	sum := sha1.Sum([]byte(fmt.Sprintf("%s:%d", query, series)))
	base := float64(binary.BigEndian.Uint32(sum[:4])%1000) / 10
	return 50 + math.Sin(float64(ts)/300+base)*35 + float64(series*12)
}

func shortQuery(query string) string {
	query = strings.Join(strings.Fields(query), " ")
	if len(query) > 48 {
		return query[:45] + "..."
	}
	if query == "" {
		return "mock query"
	}
	return query
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
