#!/bin/bash

# Default name
NAME="World"

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -n|--name)
      NAME="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

echo "Hello, $NAME! This is the example shell plugin."
echo "Current directory: $(pwd)"
echo "Script path: $0"