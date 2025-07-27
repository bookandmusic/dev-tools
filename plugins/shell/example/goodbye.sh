#!/bin/bash

# Default name
NAME="Friend"

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

echo "Goodbye, $NAME! This is the example shell plugin."
echo "Current time: $(date)"