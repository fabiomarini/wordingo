---
phase: 03-document-model
verified: 2026-07-25T23:45:00Z
status: passed
score: 12/12
behavior_unverified: 0
overrides_applied: 0
---

# Phase 3: Document Model — Verification Report

**Phase Goal:** Library opens, reads, and saves existing documents with zero unintended diffs; creates documents from templates (blank or pre-populated)
**Verified:** 2026-07-25T23:45:00Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| #  | Truth | Status | Evidence |
| -- | ----- | ------ | -------- |
| 1  | Open(path) reads existing .docx, OpenReader(r, size) reads from io.ReaderAt | ✓ VERIFIED | `open.go:16-41` — Open delegates to OpenReader, calls `opc.Open` + `parseDocument`. Tests: TestRoundTrip_Blank, _SingleParagraph, _MultiHeading, _HeaderFooter, _Synthetic* all PASS. |
| 2  | Document.Paragraphs() returns typed Paragraph wrappers with Style(), Text(), X() accessors | ✓ VERIFIED | `paragraph.go:11-35` — Paragraph struct with nil-safe Style/Text/X. `wordingo.go:105-114` — Paragraphs() iterates Body.P. Tests: TestParagraph_StyleNilChain, _TextEmptyRuns, _X, TestRoundTrip_MultiHeading all PASS. |
| 3  | Save(path)/WriteTo(w)/Close() work correctly | ✓ VERIFIED | `wordingo.go:64-79,118-122` — WriteTo uses countWriter + pkg.Save; Save creates file + calls WriteTo; Close nils pkg+doc. Tests: TestSave_File, TestClose, TestCloseThenParagraphs all PASS. |
| 4  | Unmodified parts byte-identical after open→no-op→save (STYLE-ROUNDTRIP-01/02) | ✓ VERIFIED | `roundtrip_test.go:21-65` — diffParts compares all non-manifest parts byte-for-byte. All round-trip tests (8 variants, 4 fixtures) PASS. Style parts untouched on save. |
| 5  | Lazy part loading — body eagerly parsed, supporting parts never marked modified | ✓ VERIFIED | `roundtrip_test.go:435-471` — TestLazyLoading verifies body non-nil, header1.xml exists w/ IsModified()==false, styles.xml not modified. PASS. |
| 6  | FromTemplate/FromTemplateReader — empty body + template styles (CREATE-03) | ✓ VERIFIED | `template.go:20-54` — CloneStyles + buildEmptyBodyXML + parseDocument. `template_test.go:144-181` — TestFromTemplate_EmptyBody verifies 0 paragraphs + style parts present. PASS. |
| 7  | OpenTemplate/OpenTemplateReader — body preserved + styles (CREATE-04) | ✓ VERIFIED | `template.go:62-133` — CloneStyles + source body P+Tbl preserved, clean sectPr. `template_test.go:203-238` — TestOpenTemplate_BodyPreserved verifies 3 paragraphs, correct text. PASS. |
| 8  | OpenTemplate strips sectPr HdrFtrRef/FtrRef (no dangling rIds — Pitfall 4) | ✓ VERIFIED | `template.go:108-114` — freshDoc uses defaultSectPr() with no header/footer refs. `template_test.go:230-237` — explicit assertion HdrFtrRef/FtrRef empty. PASS. |
| 9  | Style parts byte-identical after FromTemplate→Save round-trip | ✓ VERIFIED | `template_test.go:183-257` — TestFromTemplate_RoundTrip, TestOpenTemplate_RoundTrip use assertPartsMatchExcept, body excluded (expected diff). Style parts match. PASS. |
| 10 | buildEmptyBodyXML produces valid XML with correct sectPr defaults | ✓ VERIFIED | `create.go:211-228` — CT_Document with empty body, defaultSectPr(). `template_test.go:309-350` — TestBuildEmptyBodyXML_Valid verifies 0 paragraphs, W=12240, H=15840, margins, no HdrFtrRef. PASS. |
| 11 | Existing Create() tests pass after refactor | ✓ VERIFIED | `create_test.go:14-181` — TestCreate, TestSaveFile, TestXEscapeHatch all PASS after Create()→(*Document,error) sig change and Save→WriteTo rename. |
| 12 | FromTemplate body clearing replaces entirely (not in-place edit — Pitfall 3) | ✓ VERIFIED | `template.go:47` — `dst.MarkModified("word/document.xml", buildEmptyBodyXML())` — complete replacement. TestFromTemplate_EmptyBody verifies 0 paragraphs. PASS. |

**Score:** 12/12 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `open.go` | Open, OpenReader, parseDocument | ✓ VERIFIED | 67 lines, substantive implementation. Reads via Part.Open → SafeDecoder → Unmarshal. |
| `paragraph.go` | Paragraph struct + Style/Text/X | ✓ VERIFIED | 35 lines, full nil-safe accessors. |
| `template.go` | FromTemplate/FromTemplateReader/OpenTemplate/OpenTemplateReader | ✓ VERIFIED | 133 lines, CloneStyles + body policy. |
| `wordingo.go` | doc field, WriteTo, Save(path), Close, Paragraphs | ✓ VERIFIED | 122 lines. WriteTo uses countWriter. Close nils refs. |
| `create.go` | ptrInt64, defaultSectPr, buildEmptyBodyXML, newTemplateTarget | ✓ VERIFIED | 228 lines. buildEmptyBodyXML + defaultSectPr + newTemplateTarget. |
| `roundtrip_test.go` | Round-trip diff tests, lazy loading, fixture generation | ✓ VERIFIED | 703 lines. 8 round-trip tests, lazy loading, edge cases. |
| `template_test.go` | All template tests (7) | ✓ VERIFIED | 379 lines. 7 tests all PASS. |
| `create_test.go` | Updated Create/SaveFile/X tests | ✓ VERIFIED | 181 lines. All PASS after refactor. |
| `testdata/roundtrip/blank.docx` | Blank document fixture | ✓ VERIFIED | 10000 bytes, generated by TestGenerateFixtures. |
| `testdata/roundtrip/single-paragraph.docx` | Single paragraph fixture | ✓ VERIFIED | 1802 bytes. |
| `testdata/roundtrip/multi-heading.docx` | 3-paragraph fixture | ✓ VERIFIED | 1856 bytes. |
| `testdata/roundtrip/header-footer.docx` | Header/footer fixture | ✓ VERIFIED | 2163 bytes. |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| `open.go:parseDocument` | `opc.Part.Open` → `xmlutil.NewSafeDecoder` → `xml.Unmarshal` → `*wml.CT_Document` | Part lookup, decode chain | ✓ WIRED | `open.go:46-67` — reads pkg.Parts["word/document.xml"], SafeDecoder, Decode into CT_Document |
| `Document.Paragraphs()` | `d.doc.Body.P` → `[]*Paragraph` | Iteration + wrapping | ✓ WIRED | `wordingo.go:105-114` — nil-guarded iteration |
| `Document.WriteTo` | `pkg.Save(w)` | countWriter wrapper | ✓ WIRED | `wordingo.go:64-68` — countWriter → pkg.Save |
| `Document.Save(path)` | `os.Create` → `WriteTo` | File open + delegate | ✓ WIRED | `wordingo.go:71-79` |
| `Document.Close()` | nil pkg, nil doc | Assignment | ✓ WIRED | `wordingo.go:118-122` |
| `diffParts(t, original, saved)` | Per-part byte diff | ZIP unzip + byte compare | ✓ WIRED | `roundtrip_test.go:21-65` — used in all round-trip tests |
| `FromTemplateReader` | `opc.Open` → `newTemplateTarget` → `CloneStyles` → `buildEmptyBodyXML` → `parseDocument` | Full orchestration | ✓ WIRED | `template.go:34-54` |
| `OpenTemplateReader` | `opc.Open` → `newTemplateTarget` → `CloneStyles` → source body read → `defaultSectPr()` → re-encode → `parseDocument` | Full orchestration | ✓ WIRED | `template.go:76-133` — preserves P+Tbl, fresh sectPr |
| `buildEmptyBodyXML` | `defaultSectPr` → `xmlutil.NewEncoder` → encoded bytes | XML encode | ✓ WIRED | `create.go:211-228` |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| `open.go:parseDocument` | `pkg.Parts["word/document.xml"]` | opc.Part.Open (ZIP file) | ✓ Real ZIP data decoded | ✓ FLOWING |
| `paragraph.go:Style()` | `ct.PPr.PStyle.Val` | Parsed CT_Document → CT_P | ✓ Reads from actual WML parse | ✓ FLOWING |
| `paragraph.go:Text()` | `ct.R[*].T.Value` | Parsed CT_Document → CT_R → CT_Text | ✓ Reads from actual WML parse | ✓ FLOWING |
| `template.go:OpenTemplateReader` | `srcDoc.Body.P`, `.Tbl` | Source ZIP document.xml | ✓ Reads actual source body, re-encodes | ✓ FLOWING |
| `template.go:FromTemplateReader` | `buildEmptyBodyXML()` | Built from defaults | ✓ Fresh CT_Document, no hardcoded stubs | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| All round-trip tests | `go test . -run 'TestRoundTrip' -v -count=1` | 13/13 PASS | ✓ PASS |
| All template tests | `go test . -run 'TestFromTemplate|TestOpenTemplate|TestTemplate' -v -count=1` | 7/7 PASS | ✓ PASS |
| Existing Create tests | `go test . -run 'TestCreate|TestSaveFile|TestXEscapeHatch' -v -count=1` | 3/3 PASS | ✓ PASS |
| Build | `go build ./...` | PASS | ✓ PASS |
| Vet | `go vet ./...` | PASS | ✓ PASS |

### Probe Execution

**SKIPPED** — no probes declared in PLANs, no conventional `scripts/*/tests/probe-*.sh` files. Phase is pure Go library code verified by standard test suite.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| STYLE-ROUNDTRIP-01 | 03-01, 03-02 | Editing never rewrites style parts unless user explicitly modifies | ✓ SATISFIED | Round-trip tests prove byte-identity of unmodified style parts. D-06 enforced (no MarkModified on style parts). |
| STYLE-ROUNDTRIP-02 | 03-01 | Unmodified content formatting byte-identical after save | ✓ SATISFIED | Per-part byte diff in all round-trip tests. diffParts verifies byte-identity. |
| CREATE-03 | 03-02 | FromTemplate clones template, preserves styles, provides ready body | ✓ SATISFIED | FromTemplate/FromTemplateReader implemented. TestFromTemplate_EmptyBody verifies 0 paras + styles present. TestFromTemplate_RoundTrip verifies style part byte-identity. |
| CREATE-04 | 03-02 | Pre-populated template — keep existing body content, insert at specified locations | ✓ SATISFIED | OpenTemplate/OpenTemplateReader preserves body. "Insert at locations" deferred to Phase 4 (read-only Phase 3 per D-01). TestOpenTemplate_BodyPreserved verifies 3 paragraphs preserved. |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| *(none)* | — | — | — | No TBD/FIXME/XXX/HACK/placeholder markers in phase files. No empty return stubs. No console.log patterns. |

### Human Verification Required

**None.** All behaviors exercised by passing tests. No visual/real-time/external-service items.

### Gaps Summary

**No gaps found.** All 12/12 truths verified. All 4 requirements satisfied. All artifacts exist, substantive, wired, and data flows real.

---

_Verified: 2026-07-25T23:45:00Z_
_Verifier: the agent (gsd-verifier)_
