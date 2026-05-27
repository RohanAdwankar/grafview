package grafview

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func writeProvisioning(runtimeDir string, folders []string, mockPort int) error {
	if err := os.MkdirAll(filepath.Join(runtimeDir, "provisioning", "datasources"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(runtimeDir, "provisioning", "dashboards"), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(runtimeDir, "provisioning", "datasources", "datasources.yaml"), []byte(datasourcesYAML(mockPort)), 0o644); err != nil {
		return err
	}
	sort.Strings(folders)
	return os.WriteFile(filepath.Join(runtimeDir, "provisioning", "dashboards", "dashboards.yaml"), []byte(dashboardProvidersYAML(folders)), 0o644)
}

func datasourcesYAML(mockPort int) string {
	url := fmt.Sprintf("http://host.docker.internal:%d", mockPort)
	return fmt.Sprintf(`apiVersion: 1
datasources:
  - name: Mock Prometheus
    uid: %s
    type: prometheus
    access: proxy
    url: %s
    isDefault: true
  - name: Mock Loki
    uid: %s
    type: loki
    access: proxy
    url: %s
`, mockPrometheusUID, url, mockLokiUID, url)
}

func dashboardProvidersYAML(folders []string) string {
	var b strings.Builder
	b.WriteString("apiVersion: 1\nproviders:\n")
	for i, folder := range folders {
		fmt.Fprintf(&b, "  - name: %s\n", yamlQuote(fmt.Sprintf("gmv-%d-%s", i, folder)))
		b.WriteString("    orgId: 1\n")
		fmt.Fprintf(&b, "    folder: %s\n", yamlQuote(folder))
		b.WriteString("    type: file\n")
		b.WriteString("    disableDeletion: false\n")
		b.WriteString("    editable: true\n")
		b.WriteString("    updateIntervalSeconds: 5\n")
		b.WriteString("    options:\n")
		fmt.Fprintf(&b, "      path: %s\n", yamlQuote("/var/lib/grafana/dashboards/"+folderDirName(folder)))
	}
	return b.String()
}

func yamlQuote(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}
