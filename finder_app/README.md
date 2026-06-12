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

The app logs launch output to `/tmp/grafview-finder.log`.
