#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
source_dir="$script_dir/fixture"
workspace_root="$(cd -- "$script_dir/../../.." && pwd)"
target_dir="${1:-$workspace_root/demon-docs-adoption-demo}"

if [ -d "$target_dir/.ddocs" ] && command -v ddocs >/dev/null 2>&1; then
  ddocs demon run --false "$target_dir" >/dev/null 2>&1 || true
  sleep 1
fi
rm -rf -- "$target_dir"
mkdir -p -- "$target_dir"
cp -a -- "$source_dir/." "$target_dir/"

printf 'Reset Demon Docs adoption demo at %s\n' "$target_dir"
printf 'Next: cd "%s" && ddocs init --root docs\n' "$target_dir"
