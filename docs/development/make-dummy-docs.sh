#!/usr/bin/env bash
set -euo pipefail

# ----------------------------
# Config
# ----------------------------

ROOT_DIR="${ROOT_DIR:-dummy-docs}"

# 1 = delete ROOT_DIR before generating.
# 0 = add into existing tree.
RECREATE="${RECREATE:-1}"

# Extensions are chosen randomly per file.
# Include dots or omit them; both work.
EXTENSIONS=("md")

# Folder shape.
MIN_AREAS=4
MAX_AREAS=7

MIN_SUBAREAS_PER_AREA=1
MAX_SUBAREAS_PER_AREA=4

MIN_TOPICS_PER_SUBAREA=0
MAX_TOPICS_PER_SUBAREA=3

# File counts per level.
MIN_ROOT_FILES=2
MAX_ROOT_FILES=6

MIN_AREA_FILES=1
MAX_AREA_FILES=5

MIN_SUBAREA_FILES=1
MAX_SUBAREA_FILES=6

MIN_TOPIC_FILES=1
MAX_TOPIC_FILES=4

# ----------------------------
# Helpers
# ----------------------------

rand_between() {
  local min="$1"
  local max="$2"
  echo $((RANDOM % (max - min + 1) + min))
}

random_extension() {
  local ext="${EXTENSIONS[$((RANDOM % ${#EXTENSIONS[@]}))]}"
  ext="${ext#.}"
  echo "$ext"
}

write_file() {
  local path="$1"
  local title="$2"
  local ext="${path##*.}"

  case "$ext" in
    md)
      cat > "$path" <<EOT
# $title

Dummy markdown content.

## Purpose

Generated test file for recursive docs-index testing.

## Overview

This file is safe to delete.

## Notes

Generated at: $(date -Iseconds)
EOT
      ;;
    json)
      cat > "$path" <<EOT
{
  "title": "$title",
  "type": "dummy",
  "generated_at": "$(date -Iseconds)",
  "safe_to_delete": true
}
EOT
      ;;
    toml)
      cat > "$path" <<EOT
title = "$title"
type = "dummy"
generated_at = "$(date -Iseconds)"
safe_to_delete = true
EOT
      ;;
    yaml|yml)
      cat > "$path" <<EOT
title: "$title"
type: dummy
generated_at: "$(date -Iseconds)"
safe_to_delete: true
EOT
      ;;
    txt)
      cat > "$path" <<EOT
$title

Dummy text content.
Generated at: $(date -Iseconds)
Safe to delete.
EOT
      ;;
    *)
      cat > "$path" <<EOT
$title

Dummy content for .$ext file.
Generated at: $(date -Iseconds)
Safe to delete.
EOT
      ;;
  esac
}

make_random_files() {
  local dir="$1"
  local min="$2"
  local max="$3"
  local count
  count="$(rand_between "$min" "$max")"

  mkdir -p "$dir"

  for i in $(seq 1 "$count"); do
    local ext
    ext="$(random_extension)"

    local path="$dir/doc-$i-$RANDOM.$ext"
    write_file "$path" "Dummy Doc $i"
  done
}

# ----------------------------
# Generate
# ----------------------------

if [[ "$RECREATE" == "1" ]]; then
  rm -rf "$ROOT_DIR"
fi

mkdir -p "$ROOT_DIR"

make_random_files "$ROOT_DIR" "$MIN_ROOT_FILES" "$MAX_ROOT_FILES"

area_count="$(rand_between "$MIN_AREAS" "$MAX_AREAS")"

for a in $(seq 1 "$area_count"); do
  area="$ROOT_DIR/area-$a"
  mkdir -p "$area"
  make_random_files "$area" "$MIN_AREA_FILES" "$MAX_AREA_FILES"

  subarea_count="$(rand_between "$MIN_SUBAREAS_PER_AREA" "$MAX_SUBAREAS_PER_AREA")"

  for s in $(seq 1 "$subarea_count"); do
    subarea="$area/subarea-$s"
    mkdir -p "$subarea"
    make_random_files "$subarea" "$MIN_SUBAREA_FILES" "$MAX_SUBAREA_FILES"

    topic_count="$(rand_between "$MIN_TOPICS_PER_SUBAREA" "$MAX_TOPICS_PER_SUBAREA")"

    if [[ "$topic_count" -gt 0 ]]; then
      for t in $(seq 1 "$topic_count"); do
        topic="$subarea/topic-$t"
        mkdir -p "$topic"
        make_random_files "$topic" "$MIN_TOPIC_FILES" "$MAX_TOPIC_FILES"
      done
    fi
  done
done

echo "Generated dummy tree at: $ROOT_DIR"
echo
find "$ROOT_DIR" | sort
