# Plan Steps:

1. Decide what your container must contain at runtime (Go server binary, `libgec.a`-linked runtime, models, data files, webpage assets, and any shared libs like ONNX Runtime)
2. Create a **multi-stage Dockerfile**: one stage builds `libgec.a` + Go binary, the final stage is a slim runtime image
3. Make your code use **repo-relative paths** inside the container (set `GEC_ROOT=/app`, `MODEL_ROOT=/app/models`, etc.)
4. Add `.dockerignore` so you don’t accidentally ship huge build junk (and optionally handle models with bind-mounts)
5. Run with `docker run` (or `docker compose`) and verify with your `/healthz` + `/api/gec` + webpage



# Multi-stage Dockerfile

This assumes:

* You build native runtime via `make -C src/native/gec_runtime`
* You build Go server via `go build ./src/cmd/gec-server`
* Your server serves `/api/gec` and also serves static webpage files (you can serve `webpage/src` as static)





### Why I copy `/app/src` into runtime

Because your Go code currently reads spellcheck/tagger assets from the repo layout (`src/internal/.../data`).
If you later switch those to `go:embed`, you can remove that copy and the image gets smaller.

---

# Make your server serve the webpage in Docker

Inside your Go server, add static serving (if you haven’t already):

```go
// serve static webpage files
http.Handle("/", http.FileServer(http.Dir("webpage/src")))
```

Or better, mount it at `/` and keep API at `/api/...`.

---

# 4) Build + run

From repo root:

```bash
docker build -f docker/Dockerfile -t gec-demo:latest .
docker run --rm -p 8089:8089 gec-demo:latest
```

Then test:

```bash
curl -s http://localhost:8089/healthz
curl -s http://localhost:8089/info
```

And open:

* `http://localhost:8089/` (webpage)
* `POST http://localhost:8089/api/gec`

---

# 5) Docker Compose (optional, but very nice for employers)

Create `docker/docker-compose.yml`:

```yaml
services:
  gec-demo:
    build:
      context: ..
      dockerfile: docker/Dockerfile
    ports:
      - "8089:8089"
    environment:
      PORT: "8089"
      GEC_ROOT: "/app"
      MODEL_ROOT: "/app/models"
    # Optional: keep models outside the image (useful if models are huge)
    # volumes:
    #   - ../models:/app/models:ro
```

Run:

```bash
docker compose -f docker/docker-compose.yml up --build
```

---

# 6) The one thing you should verify now (before you call it “dockerized”)

Inside the container, your runtime must be able to load ONNX Runtime.

If the container starts but inference fails with something like:

* `error while loading shared libraries: libonnxruntime.so: cannot open shared object file`

That means you need to implement Approach A or B in the Dockerfile to ship ORT.

---

## Quick check: where is ONNX Runtime coming from on your machine?

Tell me **one** of these and I’ll give you the exact final Dockerfile (no options / no guessing):

1. “I installed onnxruntime via apt” (which package?)
2. “I downloaded a prebuilt onnxruntime Linux tarball” (where is `libonnxruntime.so` located?)
3. “I build onnxruntime from source” (install prefix?)
4. “I vendor ORT into the repo” (preferred)

Once we lock that down, you’ll have a truly one-command: `docker compose up --build` demo that anyone can run.
