---
phase: 3
slug: 03-document-model
status: approved
nyquist_compliant: true
wave_0_complete: true
created: 2026-07-25
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (Go 1.23+ stdlib testing) |
| **Config file** | none — Go convention |
| **Quick run command** | `go test . -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test . -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 2 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------------|-----------|-------------------|-------------|--------|
| 03-01-T1 | 01 | 1 | STYLE-ROUNDTRIP-01, STYLE-ROUNDTRIP-02 | countWriter wraps pkg.Save; no side effects on style parts | unit | `go test . -run 'TestCreate\|TestSaveFile\|TestXEscapeHatch' -count=1` | ✅ | ✅ green |
| 03-01-T2 | 01 | 1 | STYLE-ROUNDTRIP-01, STYLE-ROUNDTRIP-02 | parseDocument never calls MarkModified; lazy parts untouched | unit | `go test . -run 'TestParagraph\|TestOpen_Missing' -count=1` | ✅ | ✅ green |
| 03-01-T3 | 01 | 1 | STYLE-ROUNDTRIP-01, STYLE-ROUNDTRIP-02 | Per-part byte diff proves unmodified parts byte-identical | unit | `go test . -run 'TestRoundTrip\|TestLazyLoading\|TestClose\|TestSave_File' -count=1` | ✅ | ✅ green |
| 03-02-T1 | 02 | 2 | CREATE-03, CREATE-04, STYLE-ROUNDTRIP-01 | CloneStyles byte-copies style parts; body cleared/replaced entirely | unit | `go build ./...` | ✅ | ✅ green |
| 03-02-T2 | 02 | 2 | CREATE-03, CREATE-04, STYLE-ROUNDTRIP-01 | Round-trip diff verifies style parts byte-identical after clone | unit | `go test . -run 'TestFromTemplate\|TestOpenTemplate\|TestTemplate\|TestBuildEmptyBodyXML' -count=1` | ✅ | ✅ green |

*Status: ✅ green · ❌ red · ⚠️ flaky*

---

## Per-Requirement Coverage

| Requirement | Plans | Test File | Tests | Status |
|-------------|-------|-----------|-------|--------|
| STYLE-ROUNDTRIP-01 | 03-01, 03-02 | `roundtrip_test.go`, `template_test.go` | TestRoundTrip_Blank, TestRoundTrip_SingleParagraph, TestRoundTrip_MultiHeading, TestRoundTrip_HeaderFooter, TestRoundTrip_AllFixtures, TestFromTemplate_RoundTrip, TestOpenTemplate_RoundTrip, TestLazyLoading | ✅ COVERED |
| STYLE-ROUNDTRIP-02 | 03-01 | `roundtrip_test.go` | TestRoundTrip_Blank, TestRoundTrip_SingleParagraph, TestRoundTrip_MultiHeading, TestRoundTrip_HeaderFooter, TestRoundTrip_Synthetic_*, TestRoundTrip_AllFixtures | ✅ COVERED |
| CREATE-03 | 03-02 | `template_test.go` | TestFromTemplate_EmptyBody, TestFromTemplate_RoundTrip, TestFromTemplate_And_OpenTemplate_ReaderVariants, TestTemplateFromBlankDoc | ✅ COVERED |
| CREATE-04 | 03-02 | `template_test.go` | TestOpenTemplate_BodyPreserved, TestOpenTemplate_RoundTrip, TestFromTemplate_And_OpenTemplate_ReaderVariants | ✅ COVERED |

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 2s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-07-25
