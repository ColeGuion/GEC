#!/usr/bin/env bash
set -euo pipefail
PORT="${PORT:-8089}"
BASE_URL="http://localhost:${PORT}"

echo "Running GEC smoke test against ${BASE_URL}"

# -------- Check server health --------
echo "→ Checking /healthCheck"
curl -sf "${BASE_URL}/healthCheck" >/dev/null
echo "✓ Server is healthy"

# -------- Test /api/gec --------
echo "→ Testing /api/gec"

REQ='{
  "text": "we should buy car."
}'

RESP=$(curl -s \
  -H "Content-Type: application/json" \
  -d "$REQ" \
  "${BASE_URL}/api/gec")

echo "Response:"
echo "$RESP" | jq .

# -------- Validate response shape --------
echo "$RESP" | jq -e '.corrected_text' >/dev/null
echo "$RESP" | jq -e '.text_markups' >/dev/null

echo "✓ Response contains corrected_text and text_markups"

echo "Smoke Test: ✓ PASSED ✓"
