#!/usr/bin/env bash
set -euo pipefail

app="$HOME/Applications/grafview.app"
src="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/grafview.applescript"
plist="$app/Contents/Info.plist"
lsregister="/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"

mkdir -p "$HOME/Applications"
rm -rf "$app"
osacompile -o "$app" "$src"

/usr/libexec/PlistBuddy -c "Add :CFBundleIdentifier string com.rohanadwankar.grafview.finder" "$plist" 2>/dev/null || \
	/usr/libexec/PlistBuddy -c "Set :CFBundleIdentifier com.rohanadwankar.grafview.finder" "$plist"
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

echo "installed $app"
