# Grammar Error Correction (GEC) Service

A production-oriented Grammar Error Correction (GEC) system built with Go, C/C++, and ONNX Runtime.  
This project provides a web interface and API for submitting text and receiving grammar-corrected, marked-up output in real time.

The system uses a fine-tuned transformer model exported to ONNX and optimized for low-latency inference in a containerized docker environment.

---

## Features

- High-performance inference using ONNX Runtime (C backend)
- SentencePiece tokenization in native code
- Go HTTP server with REST API
- Interactive web interface for live grammar correction
- Dockerized deployment
- Git LFS–managed model artifacts
- Configurable logging

---

## Tech Stack

| Layer        | Technology |
|--------------|------------|
| Frontend     | HTML, CSS, JavaScript |
| Backend API  | Go |
| Inference    | C/C++ & ONNX Runtime |
| Tokenization | SentencePiece |
| Container    | Docker, Docker Compose |
| Model        | Transformer → ONNX |
| Build Tools  | Make, Shell |

---

## Repository Structure

```
gec-demo/
├── docker/ # Dockerfile and build docs
├── models/ # ONNX + tokenizer files (Git LFS)
├── src/
│ ├── cmd/ # Application entrypoint
│ ├── internal/ # Go application logic
│ └── native/ # C/C++ inference runtime
├── webpage/ # Frontend UI
├── scripts/ # Utilities and smoke tests
└── docker-compose.yml
```


---

## Prerequisites

- Docker ≥ 20.x
- Docker Compose v2
- Git LFS

Install Git LFS:

```bash
git lfs install
```

---

## Cloning the Repository

This project uses Git LFS for model files.

Clone with:

```bash
git clone https://github.com/<your-username>/gec-demo.git
cd gec-demo
git lfs pull
```

Verify the GEC model files exist:
```bash
ls models/GecModel/*.onnx
```

---

## Environment Setup

Create a `.env` file in the project root:
```env
LOG_LEVEL=2
```

Example file:

```bash
cp .env.example .env
```


> Never commit `.env` to version control.

---

## Running the Project

### Using Docker Compose

```bash
docker compose up --build
```

Server will start at:

```
http://localhost:8089
```

---

## Web Interface

After starting the server, open:

```
http://localhost:8089
```

The UI allows you to:

* Enter text
* Submit for correction
* View marked-up grammar suggestions

---

## API Usage

### POST `/api/gec`

#### Request

```json
{
  "text": "we shood buy an car."
}
```

#### Response

```json
{
  "character_count": 20,
  "contains_profanity": false,
  "corrected_text": "We should buy a car.",
  "error_character_count": 9,
  "service_time": 0.603218595,
  "text_markups": [
    {
      "index": 0,
      "length": 2,
      "message": "Change the capitalization “We”",
      "category": "GRAMMAR_SUGGESTION"
    },
    {
      "index": 3,
      "length": 5,
      "message": "Possible spelling mistake found.",
      "category": "SPELLING_MISTAKE"
    },
    {
      "index": 13,
      "length": 2,
      "message": "Did you mean “a”?",
      "category": "GRAMMAR_SUGGESTION"
    }
  ]
}
```

---

## Logging

Log levels:

| Level | Name     |
| ----- | -------- |
| 0     | CRITICAL |
| 1     | ERROR    |
| 2     | WARNING  |
| 3     | INFO     |
| 4     | DEBUG    |

Set via:

```bash
LOG_LEVEL=4 docker compose up --build
```

---

## Testing

Run smoke tests:

```bash
./scripts/smoke_test.sh
```

Format code:

```bash
./scripts/format.sh
```

---

## Performance

Typical performance on development machine:

* Average latency: ~XX ms
* Tokenization: ~X ms
* Inference: ~X ms
* Memory: ~X MB

(See `ARCHITECTURE.md` for details.)

---

## 🧠 Model Information

* Architecture: Transformer (Seq2Seq)
* Format: ONNX
* Tokenizer: SentencePiece
* Storage: Git LFS

See `MODEL_CARD.md` for training and evaluation details.

---

## License

This project is licensed under the MIT License.
See `LICENSE` for details.

Third-party licenses are listed in `THIRD_PARTY_NOTICES.md`.

---

## Author

Cole Guion
Software / ML Engineer