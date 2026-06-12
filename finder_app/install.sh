#!/usr/bin/env bash
set -euo pipefail

app="$HOME/Applications/grafview.app"
finder_app_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
src="$finder_app_dir/grafview.applescript"
logo="$finder_app_dir/logo.jpeg"
workflow_src="$finder_app_dir/Open in grafview.workflow"
workflow_dst="$HOME/Library/Services/Open in grafview.workflow"
plist="$app/Contents/Info.plist"
lsregister="/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"

mkdir -p "$HOME/Applications"
rm -rf "$app"
osacompile -s -o "$app" "$src"
iconset="$app/Contents/Resources/grafview.iconset"
mkdir -p "$iconset"
sips -s format png -z 16 16 "$logo" --out "$iconset/icon_16x16.png" >/dev/null
sips -s format png -z 32 32 "$logo" --out "$iconset/icon_16x16@2x.png" >/dev/null
sips -s format png -z 32 32 "$logo" --out "$iconset/icon_32x32.png" >/dev/null
sips -s format png -z 64 64 "$logo" --out "$iconset/icon_32x32@2x.png" >/dev/null
sips -s format png -z 128 128 "$logo" --out "$iconset/icon_128x128.png" >/dev/null
sips -s format png -z 256 256 "$logo" --out "$iconset/icon_128x128@2x.png" >/dev/null
sips -s format png -z 256 256 "$logo" --out "$iconset/icon_256x256.png" >/dev/null
sips -s format png -z 512 512 "$logo" --out "$iconset/icon_256x256@2x.png" >/dev/null
sips -s format png -z 512 512 "$logo" --out "$iconset/icon_512x512.png" >/dev/null
sips -s format png -z 1024 1024 "$logo" --out "$iconset/icon_512x512@2x.png" >/dev/null
iconutil -c icns "$iconset" -o "$app/Contents/Resources/grafview.icns"
rm -rf "$iconset"

/usr/libexec/PlistBuddy -c "Add :CFBundleIdentifier string com.rohanadwankar.grafview.finder" "$plist" 2>/dev/null || \
	/usr/libexec/PlistBuddy -c "Set :CFBundleIdentifier com.rohanadwankar.grafview.finder" "$plist"
/usr/libexec/PlistBuddy -c "Add :CFBundleIconFile string grafview.icns" "$plist" 2>/dev/null || \
	/usr/libexec/PlistBuddy -c "Set :CFBundleIconFile grafview.icns" "$plist"
/usr/libexec/PlistBuddy -c "Delete :CFBundleDocumentTypes" "$plist" 2>/dev/null || true
/usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes array" "$plist"
/usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0 dict" "$plist"
/usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:CFBundleTypeName string Grafana JSON or folder" "$plist"
/usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:CFBundleTypeRole string Viewer" "$plist"
/usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:LSHandlerRank string Alternate" "$plist"
/usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:LSItemContentTypes array" "$plist"
/usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:LSItemContentTypes:0 string public.json" "$plist"
/usr/libexec/PlistBuddy -c "Add :CFBundleDocumentTypes:0:LSItemContentTypes:1 string public.folder" "$plist"
"$lsregister" -f "$app"

mkdir -p "$HOME/Library/Services"
rm -rf "$workflow_dst"
cp -R "$workflow_src" "$workflow_dst"
/System/Library/CoreServices/pbs -flush >/dev/null 2>&1 || true

echo "installed $app"
echo "installed $workflow_dst"
