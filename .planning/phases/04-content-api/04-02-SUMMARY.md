---
phase: 04-content-api
plan: 02
type: execute
subsystem: content-api
tags: [style, named-styles, validation, warnings]
key-files:
  - wordingo.go
  - paragraph.go
  - run.go
  - style_test.go
metrics:
  files-changed: 3
  new-files: 1
  tests-added: 13
  test-status: PASS
---

## Summary

Added named style application (API-03) and public API polish (QUAL-01, QUAL-03).

### Changes

- **wordingo.go**: Added `checkStyleNames()` method that reads available style IDs from `word/styles.xml` and validates all paragraph/run style references at Save time. Unknown and XML-illegal style names produce warnings. Called inside `serializeBody()` before body encoding.
- **paragraph.go**: `SetStyle("")` clears style reference (sets PStyle to nil). Empty-string clear produces no warning.
- **run.go**: `SetStyle("")` clears run style reference (sets RStyle to nil).
- **style_test.go** (NEW): 16 tests covering style application, clearing, chaining, warnings for unknown/illegal names, nil-receiver panic, style parts byte-identity, round-trip integration, FromTemplate+SetStyle integration.

### Deviations

None.

### Self-Check: PASSED

- [x] Paragraph.SetStyle sets pStyle reference; empty string clears it
- [x] Run.SetStyle sets rStyle reference; empty string clears it
- [x] Clean style names set without warning; empty string clears reference
- [x] Unknown style names produce deferred warning at Save time (D-12)
- [x] XML-illegal characters in style name produce warning
- [x] Style parts never MarkModified — byte-identity verified (D-06)
- [x] Resolver stays internal — no public accessor (D-07)
- [x] QUAL-01: OpenReader/WriteTo I/O paths work with content
- [x] QUAL-03: Run.X() escape hatch returns *wml.CT_R
- [x] Full content doc integration test passes
- [x] All existing tests pass (go test ./... -count=1)
