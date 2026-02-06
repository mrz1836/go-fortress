#!/bin/bash
# Simple env loader for fixture testing

set -a  # Export all variables

for env_file in "$(dirname "$0")"/*.env; do
  if [[ -f "$env_file" ]]; then
    source "$env_file"
  fi
done

set +a
