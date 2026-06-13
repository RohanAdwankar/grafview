package grafview

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	mockPrometheusUID = "mock_prometheus"
	mockLokiUID       = "mock_loki"
)

type dashboardFile struct {
	Source string
	Rel    string
	Folder string
	Data   map[string]any
}

func discoverDashboards(inputs []string) ([]dashboardFile, error) {
	var out []dashboardFile
	for _, input := range inputs {
		root, err := filepath.Abs(input)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(root)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			d, ok, err := readDashboard(root)
			if err != nil || !ok {
				return nil, err
			}
			out = append(out, dashboardFile{
				Source: root,
				Rel:    filepath.Base(root),
				Folder: fallbackFolder(filepath.Base(filepath.Dir(root))),
				Data:   d,
			})
			continue
		}
		err = filepath.WalkDir(root, func(p string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || strings.ToLower(filepath.Ext(p)) != ".json" {
				return nil
			}
			d, ok, err := readDashboard(p)
			if err != nil || !ok {
				return err
			}
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			out = append(out, dashboardFile{
				Source: p,
				Rel:    rel,
				Folder: folderFor(root, rel),
				Data:   d,
			})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Source < out[j].Source })
	if len(out) == 0 {
		return nil, fmt.Errorf("no Grafana dashboard JSON files found")
	}
	return out, nil
}

func readDashboard(path string) (map[string]any, bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}
	var d map[string]any
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, false, nil
	}
	if _, ok := d["title"].(string); !ok {
		return nil, false, nil
	}
	if _, hasPanels := d["panels"]; hasPanels {
		return d, true, nil
	}
	if _, hasRows := d["rows"]; hasRows {
		return d, true, nil
	}
	if _, hasSchema := d["schemaVersion"]; hasSchema {
		return d, true, nil
	}
	return nil, false, nil
}

func folderFor(root, rel string) string {
	dir := filepath.Dir(rel)
	if dir == "." {
		return fallbackFolder(filepath.Base(root))
	}
	parts := splitRelDir(dir)
	if len(parts) == 0 {
		return fallbackFolder(filepath.Base(root))
	}
	return strings.Join(parts, " / ")
}

func splitRelDir(dir string) []string {
	raw := strings.Split(filepath.ToSlash(dir), "/")
	parts := raw[:0]
	for _, p := range raw {
		if p != "" && p != "." {
			parts = append(parts, p)
		}
	}
	return parts
}

func fallbackFolder(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "Imported"
	}
	return name
}

func writeRuntimeDashboards(files []dashboardFile, runtimeDir string) (map[string]string, error) {
	paths := map[string]string{}
	for _, f := range files {
		clean := sanitizeDashboard(f)
		folderDir := filepath.Join(runtimeDir, "dashboards", folderDirName(f.Folder))
		if err := os.MkdirAll(folderDir, 0o755); err != nil {
			return nil, err
		}
		out := filepath.Join(folderDir, outputName(f.Rel))
		b, err := json.MarshalIndent(clean, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(out, append(b, '\n'), 0o644); err != nil {
			return nil, err
		}
		paths[out] = f.Folder
	}
	return paths, nil
}

func sanitizeDashboard(f dashboardFile) map[string]any {
	clean := cloneMap(f.Data)
	clean["id"] = nil
	clean["uid"] = "gmv-" + hashText(f.Source)[:18]
	if tags, ok := clean["tags"].([]any); ok {
		clean["tags"] = append(tags, "grafview")
	} else {
		clean["tags"] = []any{"grafview"}
	}
	sanitizeValue(clean)
	return clean
}

func sanitizeValue(v any) {
	switch x := v.(type) {
	case map[string]any:
		if x["type"] == "timeseries" {
			shapeTimeseriesPanel(x)
		}
		for k, val := range x {
			if k == "datasource" {
				x[k] = mockDatasource(val)
				continue
			}
			sanitizeValue(val)
		}
		if targets, ok := x["targets"].([]any); ok {
			for _, target := range targets {
				if m, ok := target.(map[string]any); ok {
					if _, hasExpr := m["expr"]; !hasExpr && datasourceKind(x["datasource"]) == "prometheus" {
						m["expr"] = "mock_metric"
					}
				}
			}
		}
	case []any:
		for _, item := range x {
			sanitizeValue(item)
		}
	}
}

func shapeTimeseriesPanel(panel map[string]any) {
	options := ensureMap(panel, "options")
	options["legend"] = mergeMap(options["legend"], map[string]any{
		"displayMode": "list",
		"placement":   "bottom",
		"showLegend":  true,
	})
	options["tooltip"] = mergeMap(options["tooltip"], map[string]any{"mode": "single", "sort": "none"})
	fieldConfig := ensureMap(panel, "fieldConfig")
	defaults := ensureMap(fieldConfig, "defaults")
	defaults["custom"] = timeseriesCustom(defaults["custom"])
}

func timeseriesCustom(v any) map[string]any {
	return mergeMap(v, map[string]any{
		"axisBorderShow":    false,
		"axisCenteredZero":  false,
		"axisColorMode":     "text",
		"axisLabel":         "",
		"axisPlacement":     "auto",
		"barAlignment":      0,
		"drawStyle":         "line",
		"fillOpacity":       8,
		"gradientMode":      "none",
		"hideFrom":          map[string]any{"legend": false, "tooltip": false, "viz": false},
		"insertNulls":       false,
		"lineInterpolation": "linear",
		"lineWidth":         2,
		"pointSize":         4,
		"scaleDistribution": map[string]any{"type": "linear"},
		"showPoints":        "never",
		"spanNulls":         true,
		"stacking":          map[string]any{"group": "A", "mode": "none"},
		"thresholdsStyle":   map[string]any{"mode": "off"},
	})
}

func ensureMap(parent map[string]any, key string) map[string]any {
	if m, ok := parent[key].(map[string]any); ok {
		return m
	}
	m := map[string]any{}
	parent[key] = m
	return m
}

func mergeMap(v any, add map[string]any) map[string]any {
	out, _ := v.(map[string]any)
	if out == nil {
		out = map[string]any{}
	}
	for k, val := range add {
		out[k] = val
	}
	return out
}

func mockDatasource(ds any) any {
	switch datasourceKind(ds) {
	case "loki":
		return map[string]any{"type": "loki", "uid": mockLokiUID}
	case "grafana":
		return ds
	default:
		return map[string]any{"type": "prometheus", "uid": mockPrometheusUID}
	}
}

func datasourceKind(ds any) string {
	switch x := ds.(type) {
	case string:
		return datasourceKindString(x)
	case map[string]any:
		return datasourceKindString(fmt.Sprint(x["type"]) + " " + fmt.Sprint(x["uid"]))
	default:
		return "prometheus"
	}
}

func datasourceKindString(s string) string {
	s = strings.ToLower(s)
	switch {
	case strings.Contains(s, "loki"):
		return "loki"
	case strings.Contains(s, "grafana"), strings.Contains(s, "__expr__"), strings.Contains(s, "-- mixed --"), strings.Contains(s, "-- dashboard --"):
		return "grafana"
	default:
		return "prometheus"
	}
}

func cloneMap(m map[string]any) map[string]any {
	b, _ := json.Marshal(m)
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	return out
}

func outputName(rel string) string {
	base := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
	return slug(base) + "-" + hashText(rel)[:8] + ".json"
}

func folderDirName(folder string) string {
	return slug(folder) + "-" + hashText(folder)[:6]
}

func slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		ok := r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
		if ok {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "imported"
	}
	return out
}

func hashText(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
