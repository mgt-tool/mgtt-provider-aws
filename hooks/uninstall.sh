#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/.."
rm -rf bin/
echo "✓ aws provider cleaned up"
