---
phase: 6
slug: merge-edit-v1-0
status: approved
nyquist_compliant: true
wave_0_complete: true
created: 2026-07-26
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) |
| **Config file** | none — Go convention |
| **Quick run command** | `go test -run 'Test(Merge|Edit|Body|UC)' -count=1 .` |
| **Full suite command** | `go test -count=1 ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test -run 'Test(Merge|Edit|Body|UC)' -count=1 .`
- **After every plan wave:** Run `go test -count=1 ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** ~5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 06-01-T1 | 01 | 1 | MERGE-01 | T-06-01 | XML chardata auto-escaped by encoding/xml | unit | `go test -run TestMerge_Body$ -count=1 .` | ✅ | ✅ green |
| 06-01-T1 | 01 | 1 | MERGE-01 | — | Table cell {{key}} replacement | unit | `go test -run TestMerge_TableCell$ -count=1 .` | ✅ | ✅ green |
| 06-01-T1 | 01 | 1 | MERGE-01 | — | Header/footer {{key}} replacement | unit | `go test -run TestMerge_HeaderFooter$ -count=1 .` | ✅ | ✅ green |
| 06-01-T1 | 01 | 1 | MERGE-01 | — | nil opts defaults to all parts | unit | `go test -run TestMerge_NilOpts$ -count=1 .` | ✅ | ✅ green |
| 06-01-T1 | 01 | 1 | MERGE-01 | — | ScopedParts controls which parts scanned | unit | `go test -run TestMerge_ScopedParts$ -count=1 .` | ✅ | ✅ green |
| 06-01-T1 | 01 | 1 | MERGE-02 | T-06-02 | Split-run concat, first-fragment formatting preserved | unit | `go test -run TestMerge_SplitRun$ -count=1 .` | ✅ | ✅ green |
| 06-01-T1 | 01 | 1 | MERGE-02 | T-06-02 | Non-text runs preserved during split-run merge | unit | `go test -run TestMerge_SplitRunWithNonTextRuns$ -count=1 .` | ✅ | ✅ green |
| 06-01-T2 | 01 | 1 | MERGE-03 | — | Missing key surfaced in Warnings() | unit | `go test -run TestMerge_MissingKey$ -count=1 .` | ✅ | ✅ green |
| 06-01-T2 | 01 | 1 | MERGE-03 | — | Key not found in document warned | unit | `go test -run TestMerge_KeyNotFoundInDocument$ -count=1 .` | ✅ | ✅ green |
| 06-02-T1 | 02 | 2 | EDIT-01 | T-06-05 | InsertBefore by pointer identity | unit | `go test -run TestEdit_InsertBefore$ -count=1 .` | ✅ | ✅ green |
| 06-02-T1 | 02 | 2 | EDIT-01 | — | InsertAfter by pointer identity | unit | `go test -run TestEdit_InsertAfter$ -count=1 .` | ✅ | ✅ green |
| 06-02-T1 | 02 | 2 | EDIT-01 | — | DeleteParagraph removes by pointer, warn if not found | unit | `go test -run TestEdit_DeleteParagraph$ -count=1 .` | ✅ | ✅ green |
| 06-02-T1 | 02 | 2 | EDIT-01 | — | DeleteParagraph warns on cross-doc ref, no panic | unit | `go test -run TestEdit_DeleteParagraphNotFound$ -count=1 .` | ✅ | ✅ green |
| 06-02-T2 | 02 | 2 | EDIT-01 | — | DeleteRow removes correct row, returns nil | unit | `go test -run TestEdit_DeleteRow$ -count=1 .` | ✅ | ✅ green |
| 06-02-T2 | 02 | 2 | EDIT-01 | T-06-04 | DeleteRow OOB returns error, not panic | unit | `go test -run TestEdit_DeleteRowOutOfBounds$ -count=1 .` | ✅ | ✅ green |
| 06-02-T2 | 02 | 2 | EDIT-02 | T-06-03 | SetText replaces full run text | unit | `go test -run TestEdit_SetText$ -count=1 .` | ✅ | ✅ green |
| 06-02-T2 | 02 | 2 | EDIT-02 | T-06-03 | ReplaceText substring replacement | unit | `go test -run TestEdit_ReplaceText$ -count=1 .` | ✅ | ✅ green |
| 06-02-T2 | 02 | 2 | EDIT-02 | — | ReplaceText no-op when old not found | unit | `go test -run TestEdit_ReplaceTextNoMatch$ -count=1 .` | ✅ | ✅ green |
| 06-02-T1 | 02 | 2 | EDIT-03 | — | Style parts byte-identical after edits | unit | `go test -run TestEdit_StyleRoundTrip$ -count=1 .` | ✅ | ✅ green |
| 06-02-T3 | 02 | 2 | EDIT-03, QUAL-01 | — | BodyElement accessor + backward compat | unit | `go test -run TestBody_BodyAccessor$ -count=1 .` | ✅ | ✅ green |
| 06-02-T3 | 02 | 2 | QUAL-02 | — | NilBody returns nil, not panic | unit | `go test -run TestBody_NilBody$ -count=1 .` | ✅ | ✅ green |
| 06-02-T3 | 02 | 2 | QUAL-03 | — | BodyParagraph.X(), BodyTable.X() escape hatch | unit | `go test -run TestBody_XEscapeHatch$ -count=1 .` | ✅ | ✅ green |
| 06-02-T3 | 02 | 2 | MERGE-01..03, EDIT-01..03 | — | UC1: Basic template merge end-to-end | acceptance | `go test -run TestUC1_TemplateMerge$ -count=1 .` | ✅ | ✅ green |
| 06-02-T3 | 02 | 2 | MERGE-03 | — | UC2: Merge with missing keys warned | acceptance | `go test -run TestUC2_MergeWithMissingKeys$ -count=1 .` | ✅ | ✅ green |
| 06-02-T3 | 02 | 2 | EDIT-01, EDIT-02 | — | UC3: Edit after merge end-to-end | acceptance | `go test -run TestUC3_EditOperations$ -count=1 .` | ✅ | ✅ green |
| 06-02-T3 | 02 | 2 | EDIT-01, MERGE-01 | — | UC4: Table merge + row delete end-to-end | acceptance | `go test -run TestUC4_TableEditAndMerge$ -count=1 .` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-07-26
