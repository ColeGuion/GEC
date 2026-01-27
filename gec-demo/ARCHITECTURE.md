# System Architecture — GEC Demo

This document describes the internal architecture and data flow of the Grammar Error Correction system.

---

## 📐 High-Level Overview

The system is organized as a multi-layer pipeline:

````

Browser UI
↓
Go HTTP Server
↓
Native Inference Runtime (C/C++)
↓
ONNX Runtime + SentencePiece
↓
Model Output

```

Each layer is optimized for separation of concerns and performance.

---

## 🧩 Component Diagram

```

+------------------+
|  Web Frontend    |
|  (HTML / JS)     |
+--------+---------+
|
| HTTP POST
v
+------------------+
|  Go API Server   |
|  (net/http)      |
+--------+---------+
|
| CGO / FFI
v
+------------------+
| Native Runtime   |
|  (C / C++)        |
+--------+---------+
|
| ONNX API
v
+------------------+
| ONNX Runtime     |
| SentencePiece    |
+--------+---------+
|
v
+------------------+
| GEC ONNX Model   |
+------------------+

```

---

## 🌐 Frontend Layer

**Location:** `webpage/`

Responsibilities:

- Collect user input
- Send requests to backend
- Render marked-up output
- Display diagnostics

Technologies:

- Vanilla JavaScript
- HTML/CSS
- Fetch API

---

## 🖥️ Backend API Layer

**Location:** `src/cmd/gec-server`, `src/internal/api`

Responsibilities:

- HTTP routing
- Input validation
- Request lifecycle management
- Response formatting
- Logging

The server exposes REST endpoints and manages concurrency.

---

## ⚙️ Native Inference Layer

**Location:** `src/native/gec_runtime`

Responsibilities:

- Load ONNX models
- Initialize ONNX Runtime sessions
- Manage memory buffers
- Tokenize input
- Run inference
- Decode outputs

Implemented in C/C++ for:

- Low-level memory control
- Reduced overhead
- Maximum inference throughput

---

## 🔗 Go ↔ C Integration

The backend communicates with the native runtime using CGO bindings.

Key design goals:

- Minimal copying
- Deterministic memory ownership
- Explicit error propagation
- Thread-safe inference calls

---

## 🧠 Model Layer

**Location:** `models/GecModel`

Contents:

- Encoder/decoder ONNX graphs
- Tokenizer configuration
- SentencePiece model
- Generation config

The model is versioned and stored using Git LFS.

---

## 🔄 Request Lifecycle

### 1. User Submission
User enters text in browser.

### 2. HTTP Request
Frontend sends POST `/grammar`.

### 3. Validation
Go server validates JSON payload.

### 4. Tokenization
Text is passed to native runtime and tokenized.

### 5. Inference
ONNX Runtime executes encoder/decoder graphs.

### 6. Decoding
Output tokens converted back to text.

### 7. Markup
Differences are computed and annotated.

### 8. Response
Formatted JSON is returned.

---

## 📊 Performance Considerations

### Memory

- Preloaded model sessions
- Reused inference buffers
- Avoided dynamic allocation in hot paths

### Latency

- Native tokenization
- Batched ONNX calls
- Minimal CGO overhead

### Concurrency

- Thread-safe inference sessions
- Goroutine-based request handling

---

## 🐳 Container Architecture

```

Docker Image
├── Go Binary
├── Native Runtime
├── ONNX Runtime
└── Model Files

```

Runtime configuration via environment variables:

- `LOG_LEVEL`
- `HF_TOKEN`
- `PORT`

---

## 📈 Observability

### Logging

- Structured logs
- Configurable verbosity
- Request timing

### Diagnostics

- Startup validation
- Model presence checks
- Error codes

---

## 🔒 Security Architecture

- Secrets injected at runtime
- No credentials baked into images
- Isolated container environment
- Limited filesystem access

---

## ♻️ Extensibility

Designed to support:

- Model swapping
- Multi-model routing
- GPU acceleration
- Batch inference
- Streaming inference

Minimal changes required for new models.

---

## 🚧 Known Limitations

- Single-model deployment
- No horizontal scaling
- No authentication layer
- CPU-only by default

These are intentional for portfolio scope.

---

## 📚 Related Documents

- `README.md`
- `MODEL_CARD.md`
- `API.md`
- `THIRD_PARTY_NOTICES.md`