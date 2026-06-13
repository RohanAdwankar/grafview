package grafview

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestMockPrometheusQueryRange(t *testing.T) {
	m, err := startMockDataServer(0, mockModeJagged)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close(context.Background())

	var body map[string]any
	getJSON(t, "http://127.0.0.1:"+itoa(m.Port)+"/api/v1/query_range?query=up&start=1&end=3&step=1", &body)
	if body["status"] != "success" {
		t.Fatalf("unexpected status: %#v", body)
	}
	data := body["data"].(map[string]any)
	if data["resultType"] != "matrix" {
		t.Fatalf("unexpected result type: %#v", data)
	}
	if len(data["result"].([]any)) == 0 {
		t.Fatalf("missing result: %#v", data)
	}
}

func TestMockLokiQueryRange(t *testing.T) {
	m, err := startMockDataServer(0, mockModeJagged)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close(context.Background())

	var body map[string]any
	getJSON(t, "http://127.0.0.1:"+itoa(m.Port)+"/loki/api/v1/query_range?query={job=\"syslog\"}", &body)
	if body["status"] != "success" {
		t.Fatalf("unexpected status: %#v", body)
	}
	data := body["data"].(map[string]any)
	if data["resultType"] != "streams" {
		t.Fatalf("unexpected result type: %#v", data)
	}
}

func TestMockModes(t *testing.T) {
	jagged := (&mockDataServer{Mode: mockModeJagged}).sample("up", 0, 120)
	sine := (&mockDataServer{Mode: mockModeSine}).sample("up", 0, 120)
	if jagged == sine {
		t.Fatalf("jagged mode matched sine mode: %v", jagged)
	}
	if _, err := startMockDataServer(0, "flat"); err == nil {
		t.Fatal("invalid mock mode succeeded")
	}
}

func getJSON(t *testing.T, url string, out any) {
	t.Helper()
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatal(err)
	}
}

func itoa(i int) string {
	return fmtInt(i)
}

func fmtInt(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	n := i
	for n > 0 {
		pos--
		b[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(b[pos:])
}
