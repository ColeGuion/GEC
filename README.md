[cJSON](https://github.com/DaveGamble/cJSON?tab=readme-ov-file#building)


**TODO: Make misspellings either be the type `SPELLING_MISTAKE` or `PROFANITY`**
Allowed Types:
- `GRAMMAR_SUGGESTION`
- `SPELLING_MISTAKE`
- `PROFANITY`
- `GIBBERISH`



# If built with a GPU/CPU
- Update go `init()`
- Update C inference initialization






# How to make CGo Flags
[ChatGPT](https://chatgpt.com/c/69602bed-50c0-8327-91cf-e37de1bcc3b1)
Plan Steps:
1. Pick a linking strategy (recommended: build a static library `libgec.a` from `src/gec/*`).
2. Add a `#cgo CFLAGS` include path so `#include "inference.h"` resolves.
3. Add `#cgo LDFLAGS` to link the static library (and C++ stdlib if you have `.cpp` sources).
4. (Optional) Create a tiny C wrapper API so cgo only touches C symbols.


# Prompt
I want to create a project that runs my GEC model. This will be a project for my portfolio for other employers to see and test out my GEC model. 

Components:
- Webpage: Full page to write a story. Includes a check button which will send the current text to the endpoint. Once it gets the result it should markup the text accordingly by highlighting the specific text with its correlated grammar message as a tooltip when hovered above. The page should also include a drop down box that lists which example test to input and run (i.e. "Example #1", "Example #2", "Subject-Verb Agreement", "Punctuation", "Capitalization").
- Backend Endpoint: Runs GEC model and return result along with an array of where markups are needed in the text


Code Folders
- `/src`: Contains endpoint server code in Go and the code to run inference on my model in C and C++.
- `/webpage`: Contains code for my webpage



I want all of this in a docker container so that it is easy for someone else to view and build locally. It should startup the webpage and the service to accept requests for the model.

Create a step-by-step plan to create this project. Mention suggestions to improve my code quality for employers or make it easier for them to build and run the model to see it in action

# ======= MORE =======
Should run on CPU or GPU(cuda) based on machine / config.
Should I keep some stuff private? Like my model?








# Step by Step Plan

## API Contract
POST /api/gec
body: { "text": "" }

Response Output:
```json
{
  "corrected_text": "We should buy a car.",
  "contains_profanity": false,
  "character_count": 20,
  "error_character_count": 9,
  "text_markups": [
    {
      "index": 0,
      "length": 2,
      "message": "Change the capitalization “We”",
      "type": "GRAMMAR_SUGGESTION"
    },
    {
      "index": 3,
      "length": 5,
      "message": "Possible spelling mistake found.",
      "type": "Spelling Mistake"
    },
    {
      "index": 13,
      "length": 2,
      "message": "Did you mean “a”?",
      "type": "GRAMMAR_SUGGESTION"
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


Repo Layout + Build System
/src
  /cmd/server          # Go main
  /internal/api        # handlers, validation, types
  /internal/gec        # Go wrapper around C ABI
  /gec_runtime         # C/C++ inference code + headers
/webpage
  /src                 # JS/TS, CSS
  /dist                # built assets (generated)
/scripts
Dockerfile
docker-compose.yml
Makefile
README.md
LICENSE


C++ Inference Runner
Goal: Keep the model runtime isolated and language-agnostic.




Go Service: Endpoint + Static Site + Health Checks


Webpage UX: “story editor” + highlight tooltips


Dockerize It


Concrete Milestone Breakdown (so you can execute fast)

1.	Scaffold repo (folders, Makefile, README skeleton)
2. Define API types in Go (request/response structs)
3. Create C ABI wrapper + gec_cli smoke test
4. Go wrapper calls C ABI and returns edits JSON
5. Implement /api/gec, /healthz, /readyz
6. Build webpage basic editor + fetch + render highlights
7. Multi-stage Dockerfile + docker-compose.yml
8. Add tests + lint + CI
9. Final polish: screenshots/gif demo in README



