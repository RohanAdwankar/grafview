# Finder app

This builds a small macOS Finder app that opens Grafana dashboard JSON files or folders with `grafview`.

## Install

Install `grafview` first:

```bash
go install github.com/RohanAdwankar/grafview/cmd/grafview@latest
```

Then build and register the Finder app and Quick Action:

```bash
./finder_app/install.sh
```

The app is installed at `~/Applications/grafview.app`. The Quick Action is installed at `~/Library/Services/Open in grafview.workflow`.

## Use

In Finder, right click a dashboard JSON or dashboard folder, then choose `Quick Actions` -> `Open in grafview`.

For dashboard JSON files, `Open With` -> `grafview` also works.

Opening a folder is the same as running `grafview /path/to/dashboards`: `grafview` recursively finds Grafana JSON files in that folder.

The app stays open while `grafview` is running. Quit `grafview` from the Dock or app switcher to stop the Docker container and remove the temporary runtime files.

The app logs launch output to `/tmp/grafview-finder.log`.
