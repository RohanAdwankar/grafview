package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type config struct {
	port     int
	mockPort int
	image    string
	name     string
	open     bool
	keep     bool
	inputs   []string
}

func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "grafana_mock_viewer:", err)
		os.Exit(1)
	}
}

func parseFlags() config {
	cfg := config{}
	flag.IntVar(&cfg.port, "port", 0, "Grafana host port; 0 chooses a free port")
	flag.IntVar(&cfg.mockPort, "mock-port", 0, "mock datasource host port; 0 chooses a free port")
	flag.StringVar(&cfg.image, "image", "grafana/grafana:latest", "Grafana Docker image")
	flag.StringVar(&cfg.name, "name", "", "Docker container name")
	flag.BoolVar(&cfg.open, "open", true, "open the local Grafana URL")
	flag.BoolVar(&cfg.keep, "keep", false, "keep Docker container and temp runtime files on exit")
	flag.Parse()
	cfg.inputs = flag.Args()
	return cfg
}

func run(cfg config) error {
	if len(cfg.inputs) == 0 {
		return fmt.Errorf("usage: grafana_mock_viewer [flags] <dashboard.json|folder> [more...]")
	}
	files, err := discoverDashboards(cfg.inputs)
	if err != nil {
		return err
	}
	mock, err := startMockDataServer(cfg.mockPort)
	if err != nil {
		return err
	}
	defer mock.Close(context.Background())

	runtimeDir, err := os.MkdirTemp("", "grafana-mock-viewer-*")
	if err != nil {
		return err
	}
	if !cfg.keep {
		defer os.RemoveAll(runtimeDir)
	}
	if _, err := writeRuntimeDashboards(files, runtimeDir); err != nil {
		return err
	}
	if err := writeProvisioning(runtimeDir, uniqueFolders(files), mock.Port); err != nil {
		return err
	}

	if cfg.port == 0 {
		cfg.port, err = freeTCPPort()
		if err != nil {
			return err
		}
	}
	if cfg.name == "" {
		cfg.name = fmt.Sprintf("grafana-mock-viewer-%d", os.Getpid())
	}
	g := grafanaInstance{
		Name:       cfg.name,
		Image:      cfg.image,
		Port:       cfg.port,
		RuntimeDir: runtimeDir,
	}
	if err := g.Start(); err != nil {
		return err
	}
	if !cfg.keep {
		defer g.Stop()
	}
	if err := waitGrafana(cfg.port, 60*time.Second); err != nil {
		return err
	}

	url := grafanaURL(cfg.port)
	if len(files) == 1 {
		if dashURL, err := firstDashboardURL(cfg.port); err == nil && dashURL != "" {
			url = dashURL
		}
	}
	fmt.Printf("Grafana: %s\n", url)
	fmt.Printf("Dashboards: %d\n", len(files))
	fmt.Printf("Mock data: http://127.0.0.1:%d\n", mock.Port)
	fmt.Printf("Container: %s\n", cfg.name)
	fmt.Printf("Runtime files: %s\n", runtimeDir)
	if cfg.open {
		_ = openBrowser(url)
	}
	fmt.Println("Press Ctrl-C to stop.")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	fmt.Println()
	return nil
}

func uniqueFolders(files []dashboardFile) []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range files {
		if !seen[f.Folder] {
			seen[f.Folder] = true
			out = append(out, f.Folder)
		}
	}
	return out
}

func grafanaURL(port int) string {
	return fmt.Sprintf("http://localhost:%d/dashboards", port)
}

func firstDashboardURL(port int) (string, error) {
	items, err := grafanaSearch(port)
	if err != nil || len(items) == 0 {
		return "", err
	}
	return fmt.Sprintf("http://localhost:%d%s?orgId=1&from=now-6h&to=now&timezone=browser", port, items[0].URL), nil
}

func absRuntime(path string) string {
	out, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return out
}
