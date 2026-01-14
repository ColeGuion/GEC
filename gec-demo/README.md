[GitHub Repo](https://github.com/ColeGuion/GEC)  
[Webpage](http://172.21.188.179:5500/gec-demo/webpage/src/)  


# To Do
- Fix any `TODO:` comments
- Make docker portion
- Test website
  - **Make website editable after checking**
- Set Markups Category: `GRAMMAR_SUGGESTION`, `SPELLING_MISTAKE`, `PROFANITY`
- Handle `"github.com/sthorne/go-hunspell"` outside of linux machines
- Update `CONFIG_PATH` in `inference.c`
  - Fix other paths to become local to workspace
  - `spellChecker.go`: Uses an environment variable `GEC_ROOT` for the repo root
    - Set GEC_ROOT once (dev, prod, Docker, CI)
```dockerfile
ENV GEC_ROOT=/app
```


# API Contract
POST /api/gec
body: { "text": "" }

Response Output:
```json
{
  "corrected_text": "We should buy a car.",
  "character_count": 20,
  "error_character_count": 9,
  "contains_profanity": false,
  "service_time": 2.94,
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
      "category": "Spelling Mistake"
    },
    {
      "index": 13,
      "length": 2,
      "message": "Did you mean “a”?",
      "category": "GRAMMAR_SUGGESTION"
    }
  ],
  "gibberish_scores": [
    {
      "index": 0,
      "length": 20,
      "score": {
        "clean": 68.77234,
        "mild": 24.709293,
        "noise": 5.989139,
        "wordSalad": 0.5292266
      }
    },
  ]

}
```


# Repo Layout
```php
gec-demo/
  README.md
  LICENSE
  Makefile                      # top-level: build/run/test/docker
  docker/
    Dockerfile
    docker-compose.yml

  src/                          # all backend + native code
    cmd/
      gec-server/
        main.go                 # starts HTTP server on :8089, serves /api/gec + static files
    internal/
      api/
        serve.go                # routes + handlers (POST /api/gec, /healthz, /info)
      gec/
        gec.go                  # request processing, calls cgo, builds response
        findDiff.go             # FindDifferences()
        structs.go              # request/response structs
        utils.go
      speechtagger/
        *.go                    # speech tagging logic
        GobData/
          weights.gob
          tags.gob
      logging/
        print.go                # your print.Debug/Info/Error etc.

    native/
      gec_runtime/
        include/
          inference.h
          logger.h
          timer.h
          sentencepiece_wrapper.h
          wp_tokenizer.h
        src/
          inference.c
          timer.c
          logger.c
          sentencepiece_wrapper.cpp
          wp_tokenizer.cpp
        third_party/
          json.hpp              # vendored single header
        config/
          config.json           # points to model paths (relative to repo/docker)
        Makefile                # builds libgec.a
        build/                  # output artifacts
          libgec.a

  models/
    GecModel/
      encoder_model.onnx
      decoder_model.onnx
      decoder_with_past_model.onnx
      spiece.model
      tokenizer.json
      config.json
      generation_config.json
      special_tokens_map.json
      tokenizer_config.json
    GibbModel/
      ...onnx + tokenizer files...

  webpage/
    src/
      index.html
      script.js
      style.css
    dist/                       # optional: if you later add a bundler
    README.md                   # how to run just the frontend (optional)

  scripts/
    smoke_test.sh               # curl localhost:8089/api/gec and prints response
    format.sh                   # gofmt + clang-format (optional)
```