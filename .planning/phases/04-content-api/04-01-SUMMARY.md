---
phase: 04-content-api
plan: 01
type: execute
subsystem: content-api
tags: [paragraph, run, formatting, builder, warnings]
key-files:
  - wordingo.go
  - paragraph.go
  - run.go
  - format.go
  - open.go
  - format_test.go
metrics:
  files-changed: 5
  new-files: 2
  tests-added: 23
  test-status: PASS
---

## Summary

Added paragraph/run builder API with inline and paragraph formatting.

### Changes

- **wordingo.go**: Added `dirty bool` and `warnings []string` to Document; `serializeBody()`, `warn()`, `AddParagraph()` methods; Warnings() merge; WriteTo hooks into serializeBody; Paragraphs() returns doc-linked Paragraphs
- **paragraph.go**: Added `doc *Document` field; `AddRun`, `SetAlignment`, `SetSpacing`, `SetIndent`, `SetStyle`, `SetFormatting` methods with nil-safe PPr init
- **run.go** (NEW): `Run` type wrapping `*wml.CT_R` with `SetBold`, `SetItalic`, `SetUnderline`, `SetFont`, `SetSize`, `SetColor`, `SetHighlight`, `SetStyle`, `SetFormatting`, `X()` methods
- **format.go** (NEW): `Alignment` type with 4 constants + String(), `RunFormat`, `ParFormat`, `ParSpacing`, `ParIndent` structs
- **open.go**: No changes needed (zero-value fields suffice)
- **format_test.go** (NEW): 23 tests covering all setters, builder chains, warning accumulation, nil-state panics, round-trip integration, lazy body init

### Deviations

None.

### Self-Check: PASSED

- [x] Run type with all formatting setters + X()
- [x] Paragraph with AddRun + SetAlignment/SetSpacing/SetIndent/SetFormatting/SetStyle
- [x] Document.AddParagraph + dirty-flag body serialization
- [x] RunFormat, ParFormat, ParSpacing, ParIndent, Alignment types in format.go
- [x] All setters return *T for chaining; warnings accumulate; nil-state panics
- [x] serializeBody called in WriteTo before pkg.Save
- [x] All tests passing (go test ./... -count=1)
