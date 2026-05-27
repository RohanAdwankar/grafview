package grafview

import (
	"strings"
	"testing"
)

func TestDashboardProvidersYAML(t *testing.T) {
	y := dashboardProvidersYAML([]string{"group-a", "group-b / summary"})
	for _, want := range []string{`folder: "group-a"`, `folder: "group-b / summary"`, `/var/lib/grafana/dashboards/`} {
		if !strings.Contains(y, want) {
			t.Fatalf("missing %q in:\n%s", want, y)
		}
	}
}

func TestDatasourcesYAML(t *testing.T) {
	y := datasourcesYAML(19090)
	for _, want := range []string{mockPrometheusUID, mockLokiUID, "http://host.docker.internal:19090"} {
		if !strings.Contains(y, want) {
			t.Fatalf("missing %q in:\n%s", want, y)
		}
	}
}
