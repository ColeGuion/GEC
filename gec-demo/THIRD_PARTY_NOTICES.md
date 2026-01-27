# Third-Party Notices

This project includes and/or depends on third-party software and system libraries.
The following notices are provided for attribution and license awareness.

> Note: Some dependencies may be provided by the OS/toolchain at build or runtime
> (e.g., system packages). If you believe an entry is missing, please open an issue
> or submit a PR.

---

## ONNX Runtime

- Usage: Native inference runtime uses ONNX Runtime for executing ONNX graphs.
- Typical artifacts: `libonnxruntime.so` (or equivalent shared library)

License: MIT License (ONNX Runtime)

---

## SentencePiece

- Usage: Tokenization (SentencePiece model / processor)
- Typical artifacts: SentencePiece headers / library and `spiece.model`

License: Apache License 2.0 (SentencePiece)

---

## nlohmann/json

- Usage: JSON parsing (commonly distributed as `json.hpp`)
- Typical artifacts: `json.hpp`

License: MIT License (nlohmann/json)

---

## ICU (International Components for Unicode)

- Usage: Unicode handling / character processing
- Typical artifacts: ICU development libraries (e.g., `libicu-dev`)

License: ICU License (ICU project)

---

## Unicode Headers / Unicode Data

- Usage: Unicode interfaces and utilities used by ICU and/or Unicode-related code
- Typical artifacts: headers such as `alphaindex.h`, `appendable.h`, `basictz.h`, `ustring.h`, etc.

License: Unicode License (Unicode Consortium)

---

## GNU C Library (glibc) — Header Snippets

- Usage: Toolchain/system header usage
- Typical artifacts: header such as `floatn-common.h` (part of glibc)

License: GNU Lesser General Public License (LGPL) for glibc (with additional terms/exceptions for certain components)

---

## Additional Notes

- This repository may also include build tooling and system packages installed during CI or container builds.
- Licenses listed here are best-effort based on common upstream licensing for the named projects.
- Where possible, consult upstream repositories for the authoritative license text.
