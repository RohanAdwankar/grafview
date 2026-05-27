package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverDashboardsUsesFullRelativeFolder(t *testing.T) {
	root := t.TempDir()
	writeDash(t, filepath.Join(root, "group-a", "summary", "01.json"), "Group A")
	writeDash(t, filepath.Join(root, "group-b", "details", "host.json"), "Group B Host")
	if err := os.WriteFile(filepath.Join(root, "note.json"), []byte(`{"not":"dashboard"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := discoverDashboards([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 dashboards, got %d", len(got))
	}
	if got[0].Folder != "group-a / summary" || got[1].Folder != "group-b / details" {
		t.Fatalf("unexpected folders: %#v %#v", got[0].Folder, got[1].Folder)
	}
}

func TestDiscoverDashboardAtRootUsesRootFolder(t *testing.T) {
	root := t.TempDir()
	writeDash(t, filepath.Join(root, "overview.json"), "Overview")
	got, err := discoverDashboards([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Folder != filepath.Base(root) {
		t.Fatalf("unexpected folder %q", got[0].Folder)
	}
}

func TestSanitizeDashboardRewritesDatasources(t *testing.T) {
	d := dashboardFile{
		Source: "/private/path/04.json",
		Rel:    "group-a/summary/04.json",
		Folder: "group-a / summary",
		Data: map[string]any{
			"title": "Slurm",
			"uid":   "private",
			"panels": []any{
				map[string]any{
					"datasource": map[string]any{"type": "prometheus", "uid": "${ds_prometheus}"},
					"targets": []any{
						map[string]any{"expr": "runningjobs"},
					},
				},
				map[string]any{
					"datasource": map[string]any{"type": "loki", "uid": "${loki}"},
				},
			},
		},
	}

	clean := sanitizeDashboard(d)
	b, err := json.Marshal(clean)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	for _, bad := range []string{"${ds_prometheus}", "${loki}", "/private/path"} {
		if contains(text, bad) {
			t.Fatalf("sanitized dashboard still contains %q: %s", bad, text)
		}
	}
	if !contains(text, mockPrometheusUID) || !contains(text, mockLokiUID) {
		t.Fatalf("mock datasource UIDs missing: %s", text)
	}
}

func writeDash(t *testing.T, path, title string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"title":` + quote(title) + `,"schemaVersion":39,"panels":[]}`)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
}

func quote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
