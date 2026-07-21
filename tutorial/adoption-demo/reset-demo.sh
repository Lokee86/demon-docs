#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
source_dir="$script_dir/fixture"
repo_root="$(cd -- "$script_dir/../.." && pwd)"
workspace_root="$(cd -- "$repo_root/.." && pwd)"
requested_target="${1:-$workspace_root/demon-docs-adoption-demo}"
target_parent="$(cd -- "$(dirname -- "$requested_target")" && pwd)"
target_dir="$target_parent/$(basename -- "$requested_target")"

case "$target_dir" in
  "$repo_root"|"$repo_root"/*)
    printf 'ERROR: refusing to create the disposable demo inside the Demon Docs checkout.\n' >&2
    printf 'Tracked source fixture: %s\n' "$source_dir" >&2
    printf 'Requested target:       %s\n' "$target_dir" >&2
    printf 'Use the default sibling target or another directory outside %s.\n' "$repo_root" >&2
    exit 2
    ;;
esac

printf 'Tracked source fixture: %s\n' "$source_dir"
printf 'Disposable workspace:   %s\n' "$target_dir"
printf 'WARNING: the disposable workspace will be deleted and recreated.\n'

if [ -d "$target_dir/.ddocs" ] && command -v ddocs >/dev/null 2>&1; then
  ddocs demon run --false "$target_dir" >/dev/null 2>&1 || true
  sleep 1
fi
rm -rf -- "$target_dir"
mkdir -p -- "$target_dir"
cp -a -- "$source_dir/." "$target_dir/"

printf '\nDemo workspace ready.\n'
printf 'Open this directory as the Obsidian vault: %s\n' "$target_dir"
printf 'Do NOT open the tracked fixture under the Demon Docs checkout.\n'
printf 'Next: cd "%s" && ddocs init --root docs\n' "$target_dir"
