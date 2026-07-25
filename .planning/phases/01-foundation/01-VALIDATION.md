---
phase: 01
slug: foundation
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-07-25
---

# Phase 01 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (Go standard library `testing`) |
| **Config file** | go.mod (module `github.com/fabiomarini/wordingo`, go 1.23) |
| **Quick run command** | `go test ./... -count=1` |
| **Full suite command** | `go build ./... && go vet ./... && go test ./... -count=1 -v` |
| **Estimated runtime** | ~3 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -count=1`
- **After every plan wave:** Run `go build ./... && go vet ./... && go test ./... -count=1 -v`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 3 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 01-01-T1 | 01 | 1 | OPC-01 | T-01-04 | External rels warned, never fetched (no http import) | unit | `go test ./internal/opc/ -run TestOpen -count=1` | ✅ | ✅ green |
| 01-01-T1 | 01 | 1 | OPC-05 | — | Transitional + Strict conformance recorded on open | unit | `go test ./internal/opc/ -run 'TestOpen\|TestOpenStrictConformance' -count=1` | ✅ | ✅ green |
| 01-01-T1 | 01 | 1 | OPC-07 | T-01-01/02/03 | Decompression bomb, path traversal, max parts, backslash rejection | unit | `go test ./internal/opc/ -run TestSafety -count=1` | ✅ | ✅ green |
| 01-01-T1 | 01 | 1 | OPC-07 | T-01-06/07 | DOCTYPE rejection (billion-laughs), depth cap 512 | unit | `go test ./internal/xmlutil/ -run 'TestSafeDecoder_RejectsDOCTYPE\|TestSafeDecoder_DepthLimit' -count=1` | ✅ | ✅ green |
| 01-01-T1 | 01 | 1 | OPC-06 | T-01-04/05 | Monotonic rId high-water allocator, no reuse after delete | unit | `go test ./internal/opc/ -run TestRelationshipsNextRID -count=1` | ✅ | ✅ green |
| 01-01-T1 | 01 | 1 | QUAL-02 | — | No panics on corrupt/garbage input; all failures as errors | unit | `go test ./internal/opc/ -run TestNoPanicCorruptInput -count=1` | ✅ | ✅ green |
| 01-01-T2 | 01 | 1 | OPC-02 | — | Canonical ZIP entry ordering per ECMA-376 §9.1.4.2 | unit | `go test ./internal/opc/ -run 'TestCanonicalOrder\|TestRoundTrip' -count=1` | ✅ | ✅ green |
| 01-01-T2 | 01 | 1 | OPC-04 | T-01-09 | Unmodeled parts byte-identical via raw pass-through (synthetic) | unit | `go test ./internal/opc/ -run 'TestRoundTrip\|TestSave' -count=1` | ✅ | ✅ green |
| 01-01-T2 | 01 | 1 | OPC-04 | T-01-09 | Real-fixture round-trip byte-identity (6 Word .docx fixtures) | integration | `go test ./internal/opc/ -run TestRoundTripRealFixtures -count=1` | ✅ | ✅ green |
| 01-01-T2 | 01 | 1 | OPC-06 | T-01-05 | Dangling target fails closed, zero bytes written to writer | unit | `go test ./internal/opc/ -run 'TestSave/dangling' -count=1` | ✅ | ✅ green |
| 01-02-T1 | 02 | 1 | OPC-03 | T-01-08 | URI-based resolution, 40+ namespaces, prefix-agnostic parse | unit | `go test ./internal/xmlutil/ -run 'TestPrefixFor\|TestNormalizeURI' -count=1 && go test ./internal/wml/ -run TestPrefixAgnostic -count=1` | ✅ | ✅ green |
| 01-02-T1 | 02 | 1 | OPC-03 | T-01-08 | Real-producer fixture open (cross-producer identity) | integration | `go test ./internal/opc/ -run TestOpenRealFixtures -count=1` | ✅ | ✅ green |
| 01-02-T1 | 02 | 1 | WML-04 | — | RawXML token-blob capture/replay, attribute-order preserved | unit | `go test ./internal/xmlutil/ -run TestRawXML_RoundTrip -count=1` | ✅ | ✅ green |
| 01-02-T2 | 02 | 1 | WML-01 | — | ~60 essential CT_* types (actual: 78, floor: 55) | unit | `go test ./internal/wml/ -run TestExportedTypeCount -count=1` | ✅ | ✅ green |
| 01-02-T2 | 02 | 1 | WML-02 | — | Bidirectional marshal/unmarshal, canonical namespace URIs | unit | `go test ./internal/wml/ -run TestRoundTrip -count=1 && go test ./internal/xmlutil/ -run TestEncoder -count=1` | ✅ | ✅ green |
| 01-02-T2 | 02 | 1 | WML-03 | — | xml:space="preserve" honored on read/write | unit | `go test ./internal/wml/ -run TestWhitespace -count=1` | ✅ | ✅ green |
| 01-02-T2 | 02 | 1 | WML-03 | — | Run boundaries never merged (2 runs in → 2 runs out) | unit | `go test ./internal/wml/ -run TestRunBoundaries -count=1` | ✅ | ✅ green |
| 01-02-T2 | 02 | 1 | WML-04 | — | Unknown child elements hoarded and re-emitted verbatim | unit | `go test ./internal/wml/ -run TestHoard -count=1` | ✅ | ✅ green |
| 01-03-T1 | 03 | 2 | CREATE-01 | T-01-09/10 | Full default part set (9 parts), styles, sectPr via Create() | integration | `go test . -run TestCreate -count=1` | ✅ | ✅ green |
| 01-03-T1 | 03 | 2 | QUAL-01 | — | io.ReaderAt/io.Writer I/O + SaveFile convenience path | integration | `go test . -run 'TestCreate\|TestSaveFile' -count=1` | ✅ | ✅ green |
| 01-03-T1 | 03 | 2 | QUAL-03 | — | X() escape hatch returns underlying *opc.Package | unit | `go test . -run TestXEscapeHatch -count=1` | ✅ | ✅ green |
| 01-03-T2 | 03 | 2 | CREATE-02 | — | Blank.docx opens in Word/LibreOffice without repair dialog | manual | — | — | ⚠️ manual |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. Go standard `testing` package was bootstrapped in Plan 01-01 Task 0. No additional framework installation needed.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Generated blank.docx opens in Word + LibreOffice without repair dialog | CREATE-02 | Requires Microsoft Word GUI to verify absence of repair dialog — cannot be automated programmatically | 1. Run `go test . -run TestCreate -count=1` (all green). 2. Generate `.planning/tmp/blank.docx` via test helper. 3. Open in Microsoft Word 2016/2019/2021/M365 — expect NO repair dialog, Normal + Heading 1–9 + Title visible in Styles gallery. 4. Open in LibreOffice Writer — expect clean open, no errors. 5. If repair dialog appears, paste exact message (likely missing part or malformed styles value). |

---

## Validation Audit 2026-07-25

| Metric | Count |
|--------|-------|
| Gaps found | 7 |
| Resolved | 7 |
| Escalated | 0 |

### Gaps Filled by Nyquist Auditor

| Gap | Requirement | Test Added | File |
|-----|-------------|------------|------|
| 1 | OPC-03 | TestOpenRealFixtures | internal/opc/opc_test.go |
| 2 | OPC-04 | TestRoundTripRealFixtures | internal/opc/opc_test.go |
| 3 | WML-01 | TestExportedTypeCount (enhanced — AST-based count = 78) | internal/wml/wml_test.go |
| 4 | WML-03 | TestRunBoundaries | internal/wml/wml_test.go |
| 5 | QUAL-01 | TestSaveFile | create_test.go |
| 6 | QUAL-02 | TestNoPanicCorruptInput | internal/opc/opc_test.go |
| 7 | QUAL-03 | TestXEscapeHatch | create_test.go |

### Test Count Summary

| Package | Tests | Subtests | Status |
|---------|-------|----------|--------|
| wordingo (root) | 3 | 0 | ✅ green |
| internal/opc | 11 | 24 | ✅ green (1 skip: TestRoundTripHostileFixture — hostile fixture absent) |
| internal/wml | 10 | 3 | ✅ green |
| internal/xmlutil | 10 | 12 | ✅ green |
| **Total** | **34** | **39** | **✅ all pass** |

---

## Validation Sign-Off

- [x] All automated tasks have `<automated>` verify (Task 01-03-T2 is `checkpoint:human-verify` by design — manual-only)
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references (none — existing infra sufficient)
- [x] No watch-mode flags (`-count=1` enforces fresh run, no caching)
- [x] Feedback latency < 3s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-07-25
