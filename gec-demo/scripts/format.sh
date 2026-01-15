#!/usr/bin/env bash
set -euo pipefail

echo "Formatting code..."

# -------- Go formatting --------
if ! command -v gofmt >/dev/null; then
  echo "❌ gofmt not found"
  exit 1
fi

echo "→ Formatting Go files"
gofmt -w src

# -------- C / C++ formatting --------
if ! command -v clang-format >/dev/null; then
  echo "clang-format not found — skipping C/C++ formatting"
else
  echo "→ Formatting C/C++ files"
  find src/native/gec_runtime \
    \( -name "*.c" -o -name "*.h" -o -name "*.cpp" -o -name "*.hpp" \) \
    -exec clang-format -i {} +
fi

echo "Formatting Complete!"
