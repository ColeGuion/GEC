#!/usr/bin/env bash
set -euo pipefail

: "${HF_REPO_ID:?Set HF_REPO_ID (e.g. username/gec-model)}"
: "${MODEL_DIR:=/models/GecModel/}"
: "${HF_HOME:=/models/.cache/huggingface}"

echo "MODEL_DIR: ${MODEL_DIR}"
echo "HF_REPO_ID: ${HF_REPO_ID}"
echo "HF_HOME: ${HF_HOME}"

mkdir -p "$MODEL_DIR"
mkdir -p "$HF_HOME"

# Only download if dir is empty (or missing expected file)
if [ -z "$(ls -A "$MODEL_DIR" 2>/dev/null || true)" ]; then
  echo "Model not found in ${MODEL_DIR}. Downloading from Hugging Face..."

  # Auth: HF_TOKEN env var is supported by huggingface_hub tooling. :contentReference[oaicite:6]{index=6}
  if [ -z "${HF_TOKEN:-}" ]; then
    echo "ERROR: HF_TOKEN is not set (needed for private repo download)."
    exit 1
  fi

  # Download the whole repo snapshot to a local folder.
  # huggingface-cli download supports --local-dir. :contentReference[oaicite:7]{index=7}
  huggingface-cli download "$HF_REPO_ID" \
    --local-dir "$MODEL_DIR" \
    --token "$HF_TOKEN"

  echo "Download complete."
else
  echo "Model already present in ${MODEL_DIR}; skipping download."
fi

exec /app/gec-server
