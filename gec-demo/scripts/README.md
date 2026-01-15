# NOTES:
```
set -euo pipefail
```

Turns on three safety features in bash scripts. Together, they make scripts fail fast, loudly, and correctly instead of silently doing the wrong thing.
1. `-e`: Exit on Error
   - If any command exits with a non-zero status, the script immediately stops.
2. `-u`: Treat unset variables as errors
   - If you reference an undefined variable, the script exits immediately.
3. `-o`: Fail pipelines correctly
   - Normally, pipelines only return the exit code of the last command.



# `smoke_test.sh`

Tests if the system works. For this project, that means:

* Is the Go server running?
* Does `/api/gec` accept input?
* Does it return valid JSON with expected fields?



# `format.sh`

A **format script** enforces **consistent style** across languages.

This is *huge* for code reviews.

For your repo, it should:

* `gofmt` all Go code
* `clang-format` all C / C++ code
* fail loudly if tools are missing


```bash
chmod +x scripts/format.sh
```

### How you’ll use it

```bash
./scripts/format.sh
```

Or wire it into Makefile:

```makefile
format:
	./scripts/format.sh
```

---

# Why these scripts matter (portfolio perspective)

| Script          | What it tells employers                     |
| --------------- | ------------------------------------------- |
| `smoke_test.sh` | “This service is testable and safe to run”  |
| `format.sh`     | “I care about code quality and consistency” |

Together, they:

* reduce onboarding friction
* make CI trivial
* show professional habits

---

# Optional next upgrades (easy wins)

If you want to go further:

* Add `make smoke`
* Add `make format`
* Add a GitHub Action that runs both
* Add a `--fail-on-diff` mode to `format.sh` for CI

If you want, I can:

* wire these into your Makefile
* create a `.clang-format`
* add a GitHub Actions workflow that runs **build + smoke test**

Just say the word.

# Run Scripts
```bash
./scripts/smoke_test.sh
./scripts/format.sh
```