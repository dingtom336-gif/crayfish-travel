#!/bin/bash
# Download Stitch assets with redirect handling
URL="$1"
DEST="$2"
mkdir -p "$(dirname "$DEST")"
curl -sL -o "$DEST" "$URL" && echo "OK: $DEST ($(wc -c < "$DEST") bytes)" || echo "FAIL: $DEST"
