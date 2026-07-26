---
phase: 05
slug: rich-content
status: partial
nyquist_compliant: false
wave_0_complete: true
created: 2026-07-26
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (Go 1.23) |
| **Config file** | none — Go standard testing |
| **Quick run command** | `go test -run 'TestTable|TestImage|TestHeader|TestFooter|TestPageSetup' -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test -run 'TestTable|TestImage|TestHeader|TestFooter|TestPageSetup' -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 2 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 05-01-T1 | 01 | 1 | API-04, API-05 | — | Shared WML types compile and marshal correctly | unit | `go build ./...` | — | ✅ green |
| 05-01-T2 | 01 | 1 | API-04 | T-05-04 | Tables — AddTable creates grid from strings, AddTableBuilder chain sets style/width/borders/shading/merge, empty data returns error | unit | `go test -run TestTable -v` | `table_test.go` | ✅ green |
| 05-01-T3 | 01 | 1 | API-05 | T-05-01, T-05-02, T-05-03 | Images — AddImageBytes creates media part + rel + content type + DrawingML inline; SetImageWidth/SetImageHeight mutate extent; DPI detection for PNG/JPEG | unit | `go test -run TestImage -v` | `image_test.go` | ✅ green |
| 05-02-T1 | 02 | 2 | API-07 | T-05-05, T-05-06 | Headers/footers — AddHeader/AddFooter create OPC parts with correct rel/CT/sectPr linking; all 3 variants supported; AddParagraph adds content; X() escape hatch | unit | `go test -run 'TestHeader|TestFooter' -v` | `header_test.go` | ⬜ pending (2 known bugs escalated) |
| 05-02-T2 | 02 | 2 | API-09 | T-05-07 | Page setup — SetOrientation swaps W/H; SetPaperSize sets twips; SetMargins sets PgMar; AddPageBreak creates paragraph with PageBreakBefore; chaining works | unit | `go test -run TestPageSetup -v` | `page_test.go` | ✅ green |
| 05-03-T1 | 03 | 2 | API-06 | T-05-09, T-05-10 | Lists — AddList with ordered/bulleted, 9-level depth, AddListFromSlice, AddNumberingDef, template numbering merge without ID collision | unit | `go test -run 'TestList' -v` | `list_test.go` | ✅ green |
| 05-03-T2 | 03 | 2 | API-08 | T-05-08 | Hyperlinks — AddHyperlink creates CT_Hyperlink with External relationship; returned Run chains formatting; multiple calls create distinct rIds | unit | `go test -run 'TestHyperlink' -v` | `list_test.go` | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `table_test.go` — Table builder API tests (AddTable, AddTableBuilder, borders, shading, merge, escape hatch)
- [x] `image_test.go` — Image embedding tests (AddImageBytes, sizing, DPI detection, content type)
- [x] `header_test.go` — Header/footer tests (creation, variants, content, escape hatch; 2 known bugs SKIP'd)
- [x] `page_test.go` — Page setup tests (orientation, paper size, margins, page break, chaining, section wrapper)
- [x] `list_test.go` — Existing list + hyperlink tests (10 passing)

*Existing infrastructure covers all phase requirements.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Header content persists after AddParagraph | API-07 | Implementation bug: header.go serializes at creation time but AddParagraph modifies in-memory CT_Hdr without re-serializing. The part is not marked modified after content changes. | Create header via AddHeader(HeaderDefault), add paragraph via h.AddParagraph("content"), save & re-open, verify header1.xml contains "content" in zip |
| Footer content persists after AddParagraph | API-07 | Same bug as header — footer.go AddParagraph does not trigger re-serialization of the footer part | Create footer via AddFooter(FooterDefault), add paragraph via f.AddParagraph("content"), save & re-open, verify footer1.xml contains "content" |
| FooterReference serializes correctly | API-07 | Implementation bug: CT_HdrFtrRef in internal/wml/document.go:220 has XMLName hardcoded to "headerReference" — both header and footer references serialize as `<w:headerReference>`. Word may not recognize footer references. | Create doc with AddFooter, inspect serialized XML in document.xml — footer reference should use `<w:footerReference>` element name |

*All other phase behaviors have automated verification.*

---

## Known Bugs (Escalated)

| Bug ID | File | Description | Severity | Status |
|--------|------|-------------|----------|--------|
| B-05-01 | `header.go` | Header/footer content not persisted — AddParagraph modifies in-memory struct but does not re-serialize the header/footer part | medium | escalated |
| B-05-02 | `internal/wml/document.go:220` | CT_HdrFtrRef.XMLName hardcoded to `"headerReference"` — footer references serialize as `<w:headerReference>` instead of `<w:footerReference>` | medium | escalated |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [ ] Escalated bugs resolved before `nyquist_compliant: true`

**Approval:** pending (2 escalated bugs remain)

---

## Validation Audit 2026-07-26

| Metric | Count |
|--------|-------|
| Gaps found | 4 |
| Resolved | 3 |
| Escalated | 1 (2 sub-bugs) |
