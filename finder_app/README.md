# Finder app

This builds a small macOS Finder app that opens Grafana dashboard JSON files or folders with `grafview`.

## Install

Install `grafview` first:

```bash
go install github.com/RohanAdwankar/grafview/cmd/grafview@latest
```

Then build and register the Finder app:

```bash
./finder_app/install.sh
```

The app is installed at `~/Applications/grafview.app`.

## Use

In Finder, right click a dashboard JSON or dashboard folder, then choose `Open With` -> `grafview`.

Opening a folder is the same as running `grafview /path/to/dashboards`: `grafview` recursively finds Grafana JSON files in that folder. If Finder does not show `Open With` for a folder, drag the folder onto `~/Applications/grafview.app`.

The app logs launch output to `/tmp/grafview-finder.log`.
