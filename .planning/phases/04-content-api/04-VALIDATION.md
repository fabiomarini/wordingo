---
phase: 04
slug: content-api
status: final
nyquist_compliant: true
wave_0_complete: true
created: 2026-07-26
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (Go stdlib) |
| **Config file** | none — Go stdlib, no test framework deps |
| **Quick run command** | `go test ./... -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~3 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 3 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 04-01-01 | 01 | 1 | API-01, API-02, QUAL-02 | T-04-01 / T-04-02 / T-04-03 / T-04-04 | Warnings for invalid input (color, size); typed structs prevent injection | unit | `go test -count=1 -run "TestAddParagraph\|TestSerializeBody\|TestDirtyFlag\|TestDocWarnings"` | ✅ | ✅ green |
| 04-01-02 | 01 | 1 | API-01, API-02, QUAL-02 | T-04-01 / — | Nil-safe RPr/PPr init; warning on invalid alignment/size/color | unit | `go test -count=1 -run "TestRunFormatting\|TestParagraphFormatting\|TestBuilderChain\|TestFormatWarnings"` | ✅ | ✅ green |
| 04-01-03 | 01 | 1 | API-01, API-02, QUAL-02 | — | Integration: builder chain, round-trip save/open, lazy body init | integration | `go test -count=1 -run "TestBuilderChainIntegration\|TestRoundTripSaveOpen\|TestAddParagraphNilBody\|TestAddRunExistingParagraph\|TestSetStyleChain"` | ✅ | ✅ green |
| 04-02-01 | 02 | 2 | API-03, QUAL-01 | T-04-05 / T-04-06 / T-04-07 | Style name in typed CT_PStyle.Val; no raw XML; D-06 enforced | unit | `go test -count=1 -run "TestSetStyle\|TestStyleValidation"` | ✅ | ✅ green |
| 04-02-02 | 02 | 2 | API-03, QUAL-01, QUAL-03 | — | Full content doc build/save/open; style parts byte-identity; X() hatch | integration | `go test -count=1 -run "TestFullContentDoc\|TestStylePartsUntouched\|TestFromTemplateWithStyle\|TestQualRunX\|TestQualParagraphX\|TestQualDocumentX"` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. Go test framework is stdlib — no install needed.

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 3s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-07-26
