package grafview

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type grafanaInstance struct {
	Name       string
	Image      string
	Port       int
	RuntimeDir string
}

func (g grafanaInstance) Start() error {
	_ = exec.Command("docker", "rm", "-f", g.Name).Run()
	args := []string{
		"run", "-d",
		"--name", g.Name,
		"-p", fmt.Sprintf("127.0.0.1:%d:3000", g.Port),
		"-e", "GF_AUTH_ANONYMOUS_ENABLED=true",
		"-e", "GF_AUTH_ANONYMOUS_ORG_ROLE=Admin",
		"-e", "GF_AUTH_DISABLE_LOGIN_FORM=true",
		"-e", "GF_USERS_DEFAULT_THEME=light",
		"-e", "GF_DASHBOARDS_MIN_REFRESH_INTERVAL=1s",
		"-v", absRuntime(g.RuntimeDir) + "/provisioning:/etc/grafana/provisioning:ro",
		"-v", absRuntime(g.RuntimeDir) + "/dashboards:/var/lib/grafana/dashboards:ro",
	}
	if runtime.GOOS == "linux" {
		args = append(args, "--add-host=host.docker.internal:host-gateway")
	}
	args = append(args, g.Image)
	out, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker run failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (g grafanaInstance) Stop() {
	_ = exec.Command("docker", "rm", "-f", g.Name).Run()
}

func waitGrafana(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	url := fmt.Sprintf("http://127.0.0.1:%d/api/health", port)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := http.DefaultClient.Do(req)
		cancel()
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("Grafana did not become healthy on port %d", port)
}

type grafanaSearchItem struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func grafanaSearch(port int) ([]grafanaSearchItem, error) {
	url := fmt.Sprintf("http://127.0.0.1:%d/api/search?type=dash-db", port)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var items []grafanaSearchItem
	return items, json.NewDecoder(resp.Body).Decode(&items)
}

func freeTCPPort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Run()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Run()
	default:
		return exec.Command("xdg-open", url).Run()
	}
}
