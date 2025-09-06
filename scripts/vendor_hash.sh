#!/usr/bin/env bash

# inspired by https://github.com/ghostty-org/ghostty/blob/main/nix/build-support/check-zig-cache-hash.sh

set -ex
VENDOR_HASH_FILE=$(realpath "$(dirname "$0")/../nix/goVendorHash.nix")

CURRENT_HASH=$(nix eval --raw --file "$VENDOR_HASH_FILE")
echo "Current hash: $CURRENT_HASH"

GO_VENDOR_DIR="$(mktemp --directory --suffix=gophertype-vendor)"

go mod vendor -o "$GO_VENDOR_DIR"

VENDOR_HASH=$(nix hash path "$GO_VENDOR_DIR")
echo "New hash: $VENDOR_HASH"

if [ "$CURRENT_HASH" == "$VENDOR_HASH" ]; then
  echo "Go vendor hash is up to date"
  exit 0
elif [ "$1" != "--update" ]; then
  echo "Vendor hashes do not match, rerun the script with --update"
  exit 1
fi

cat <<EOF >"$VENDOR_HASH_FILE"
"$VENDOR_HASH"
EOF
