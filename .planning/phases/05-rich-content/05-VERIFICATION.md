---
phase: 05-rich-content
verified: 2026-07-26T10:30:00Z
status: passed
score: 32/32 must-haves verified
behavior_unverified: 0
overrides_applied: 0
gaps: []
deferred: []
behavior_unverified_items: []
human_verification: []
---

# Phase 5: Rich Content Verification Report

**Phase Goal:** Complete business documents: tables, images, headers/footers, lists, hyperlinks, page setup
**Verified:** 2026-07-26T10:30:00Z
**Status:** PASSED
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | CT_P.Hyperlink field added for hyperlink support | ✓ VERIFIED | `internal/wml/document.go:56` — `Hyperlink []*CT_Hyperlink` field on CT_P |
| 2 | CT_R.Drawing field added for image support | ✓ VERIFIED | `internal/wml/document.go:87` — `Drawing *CT_Drawing` field on CT_R |
| 3 | DrawingML inline types in internal/wml/drawing.go compile and marshal | ✓ VERIFIED | `internal/wml/drawing.go` — 25 DrawingML struct types; `go build ./...` passes |
| 4 | AddTable(data [][]string) creates grid table from string data | ✓ VERIFIED | `table.go:247-294` — creates CT_Tbl with TblPr, TblGrid, GridCol, rows/cells, appends to Body.Tbl |
| 5 | AddTableBuilder() returns builder for complex tables | ✓ VERIFIED | `table.go:299-313` — returns TableBuilder with initial CT_Tbl, appends to Body.Tbl |
| 6 | Table/Row/Cell builders set style, width, borders, shading, merge | ✓ VERIFIED | `table.go:38-240` — SetTableStyle, SetWidth, SetBorders, SetShading on TableBuilder; SetBorders on RowBuilder; SetText, SetShading, SetBold, SetWidth, MergeRight, MergeDown on CellBuilder |
| 7 | AddImage(path) and AddImageBytes create media part + rel + content type | ✓ VERIFIED | `image.go:180-286` — AddImage reads file, AddImageBytes creates MarkModified part, rel with relImage, content type override, DrawingML inline |
| 8 | Image returns *Run with DrawingML inline element | ✓ VERIFIED | `image.go:285` — returns `&Run{ct: r, para: ...}` where r has CT_R.Drawing field |
| 9 | AddHeader creates word/header1.xml part with relationship and sectPr link | ✓ VERIFIED | `header.go:105-157` — creates CT_Hdr, serializes, addHelperPart with relHeader/ctHeader, appends CT_HdrFtrRef to SectPr |
| 10 | AddFooter creates word/footer1.xml part with relationship and sectPr link | ✓ VERIFIED | `header.go:162-214` — symmetric to AddHeader with relFooter/ctFooter |
| 11 | All three header/footer variants (default, first, even) supported | ✓ VERIFIED | `format.go:58-99` — HeaderVariant/FooterVariant enums with String() outputting "default"/"first"/"even" |
| 12 | Header/Footer types expose AddParagraph for content | ✓ VERIFIED | `header.go:43-54` — Header.AddParagraph, Footer.AddParagraph append CT_P to header/footer |
| 13 | OpenTemplate clones template headers/footers with rId fixup | ✓ VERIFIED | `template.go:118-196` — scans srcSectPr HdrFtrRef/FtrRef, copies part bytes, allocates fresh rId via dstRels.NextRID(), adds content type override |
| 14 | SetOrientation(Landscape) swaps PgSz W/H on default sectPr | ✓ VERIFIED | `page.go:27-50` — Section.SetOrientation swaps W/H for landscape when W<H |
| 15 | SetPaperSize(PaperA4) sets correct twips dimensions | ✓ VERIFIED | `page.go:54-64` — Section.SetPaperSize sets PgSz.W and PgSz.H |
| 16 | SetMargins(t,r,b,l) sets PgMar fields | ✓ VERIFIED | `page.go:68-80` — Section.SetMargins sets Top, Right, Bottom, Left on PgMar |
| 17 | AddPageBreak() creates paragraph with PageBreakBefore | ✓ VERIFIED | `page.go:124-139` — creates CT_P with PPr.PageBreakBefore = &CT_OnOff{} |
| 18 | Section wrapper provides same API as Document-level methods | ✓ VERIFIED | `page.go:10-80` — Section type with SetOrientation, SetPaperSize, SetMargins; Document delegates via Section() |
| 19 | AddList(true) creates numbering def with abstractNum + num entries | ✓ VERIFIED | `list.go:195-219` — AddList scans existing numbering, creates CT_AbstractNum with 9 levels + CT_Num, writes numbering.xml. Tested in TestList_Ordered (PASS) |
| 20 | AddList(false) creates bulleted list with bullet numFmt | ✓ VERIFIED | `list.go:156-183` — makeLevels(false) uses numFmt=bullet with Unicode chars. Tested in TestList_Bulleted (PASS) |
| 21 | AddItem creates paragraph with correct NumPr (numId + ilvl) | ✓ VERIFIED | `list.go:36-60` — AddItem creates CT_P with PPr.NumPr.ILvl and PPr.NumPr.NumId. Tested in TestList_MultiLevel (PASS) |
| 22 | AddListFromSlice creates flat list in one call | ✓ VERIFIED | `list.go:224-233` — AddList + AddItem for each string. Tested in TestList_AddListFromSlice (PASS) |
| 23 | AddNumberingDef for custom definitions | ✓ VERIFIED | `list.go:242-279` — user-specified numFmt and start for all 9 levels. Tested in TestList_AddNumberingDef (PASS) |
| 24 | Full 9-level depth (0-8) supported | ✓ VERIFIED | `list.go:156-183` — makeLevels generates 9 CT_Lvl entries. Tested in TestList_Ordered (counts 9 `<w:lvl `) (PASS) |
| 25 | Auto-numbering merges with template without collision | ✓ VERIFIED | `list.go:131-152` — findMaxAbstractNumID/findMaxNumID scan existing entries. Tested in TestList_TemplateNumberingPreserved (PASS) |
| 26 | para.AddHyperlink creates CT_Hyperlink with CT_R child | ✓ VERIFIED | `hyperlink.go:18-46` — creates CT_Hyperlink{ID: rId, R: []*CT_R{run}}. Tested in TestHyperlink_Basic (PASS) |
| 27 | AddHyperlink creates External relationship | ✓ VERIFIED | `hyperlink.go:29-35` — rel with TargetMode: "External". Tested in TestHyperlink_Basic (checks TargetMode="External") (PASS) |
| 28 | Hyperlink Run chains formatting (SetBold, SetColor) | ✓ VERIFIED | `hyperlink.go:45` — returns `&Run{ct: run, para: p}`. Tested in TestHyperlink_FormattingChaining (PASS) |
| 29 | Tables render after all paragraphs (v1 per D-24) | ✓ VERIFIED | `internal/wml/document.go:27-33` — CT_Body serializes P before Tbl in struct field order |
| 30 | nextImageID/nextHeaderID/nextFooterID init to 1 in all constructors | ✓ VERIFIED | `wordingo.go:71-73` (Create), `open.go:43-45` (OpenReader), `template.go:57-59` (FromTemplateReader), `template.go:235-237` (OpenTemplateReader) |
| 31 | Content type and relationship constants available | ✓ VERIFIED | `create.go:46-56` — relHeader, relFooter, relImage, relHyperlink, ctHeader, ctFooter, ctPng, ctJpeg |
| 32 | All tests pass: go test ./... -count=1 | ✓ VERIFIED | All 5 packages pass (wordingo, opc, style, wml, xmlutil) |

**Score:** 32/32 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/wml/drawing.go` | 25 DrawingML struct types | ✓ VERIFIED | CT_Drawing, CT_Inline, CT_Anchor, CT_Extent, CT_EffectExtent, CT_DocPr, CT_CNvGraphicFramePr, CT_GraphicFrameLocks, CT_Graphic, CT_GraphicData, CT_Pic, CT_NonVisualPicProps, CT_CNvPr, CT_CNvPicPr, CT_PicLocks, CT_BlipFill, CT_Blip, CT_Stretch, CT_FillRect, CT_SpPr, CT_Xfrm, CT_Point2D, CT_PositiveSize2D, CT_PresetGeometry, CT_AvLst |
| `internal/wml/hyperlink.go` | CT_Hyperlink type | ✓ VERIFIED | CT_Hyperlink with ID, Anchor, R, Raw fields |
| `table.go` | TableBuilder, RowBuilder, CellBuilder, AddTable, AddTableBuilder | ✓ VERIFIED | Full fluent builder API with chaining, X() escape hatch |
| `image.go` | AddImage, AddImageBytes, EMU helpers, DPI detection, SetImageWidth/SetImageHeight | ✓ VERIFIED | EMU conversion, JFIF/pHYs DPI, DrawingML inline builder, sizing |
| `header.go` | Header/Footer types, AddHeader, AddFooter, addHelperPart | ✓ VERIFIED | OPC part creation, relationship, content type, sectPr linking |
| `page.go` | Section type, page setup methods, AddPageBreak | ✓ VERIFIED | SetOrientation, SetPaperSize, SetMargins, Section wrapper |
| `list.go` | AddList, AddListFromSlice, AddNumberingDef, ListBuilder | ✓ VERIFIED | Auto-numbering defs, 9-level depth, template merge |
| `hyperlink.go` | Paragraph.AddHyperlink | ✓ VERIFIED | CT_Hyperlink creation, External relationship, Run chaining |
| `list_test.go` | List + hyperlink test suite | ✓ VERIFIED | 10 test cases (7 list + 3 hyperlink) all passing |
| `format.go` | HeaderVariant, FooterVariant, PageOrientation, PaperSize, TableBorders, BorderDef | ✓ VERIFIED | All enum types and structs present |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| CT_P.Hyperlink | CT_Hyperlink in hyperlink.go | struct field on CT_P | ✓ WIRED | `internal/wml/document.go:56` links to CT_Hyperlink type |
| CT_R.Drawing | CT_Drawing in drawing.go | struct field on CT_R | ✓ WIRED | `internal/wml/document.go:87` links to CT_Drawing type |
| AddHeader/AddFooter | sectPr via CT_HdrFtrRef | relationship rId | ✓ WIRED | `header.go:149-153` (header) and `header.go:206-210` (footer) append refs with Type + ID |
| OpenTemplate clone | fresh dst rId via dstRels.NextRID | part copy + rel allocation | ✓ WIRED | `template.go:118-196` — source rel lookup, part copy, new rId, content type override, sectPr link fixup |
| ListBuilder | CT_NumPr on paragraphs | numId + ilvl linking | ✓ WIRED | `list.go:47-55` — NumPr.ILvl and NumPr.NumId set on each AddItem paragraph |
| AddHyperlink | External relationship | relHyperlink + rId | ✓ WIRED | `hyperlink.go:29-35` — relationship with TargetMode="External" |
| Image part | document.xml.rels | relImage relationship | ✓ WIRED | `image.go:227-237` — NextRID, append rel with Type=relImage, Target=media/ |
| Page setup | CT_SectPr mutation | Section wrapper | ✓ WIRED | `page.go:27-80` — methods mutate PgSz/PgMar on Section which wraps Body.SectPr |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|-------------------|--------|
| AddTable | data [][]string input | function parameter | ✓ FLOWING | Rows/cells built from input; empty cells handled gracefully |
| AddImageBytes | data []byte input | function parameter | ✓ FLOWING | Media part created via MarkModified with actual bytes |
| Header/footer content | CT_P appended to CT_Hdr/CT_Ftr | AddParagraph method | ✓ FLOWING | Paragraph text flows into serialized XML part |
| List items | text string per item | AddItem parameter | ✓ FLOWING | Text set as CT_Text.Value inside CT_P.R[0].T |
| Hyperlink target | uri string | function parameter | ✓ FLOWING | URI stored as relationship Target with TargetMode="External" |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Build compiles | `go build ./...` | exit 0 | ✓ PASS |
| Vet passes | `go vet ./...` | exit 0 | ✓ PASS |
| All tests pass | `go test ./... -count=1` | exit 0, all 5 pkgs ok | ✓ PASS |
| List ordered | `go test -run TestList_Ordered -v` | PASS | ✓ PASS |
| List bulleted | `go test -run TestList_Bulleted -v` | PASS | ✓ PASS |
| List multi-level | `go test -run TestList_MultiLevel -v` | PASS | ✓ PASS |
| List from slice | `go test -run TestList_AddListFromSlice -v` | PASS | ✓ PASS |
| List numbering def | `go test -run TestList_AddNumberingDef -v` | PASS | ✓ PASS |
| List template merge | `go test -run TestList_TemplateNumberingPreserved -v` | PASS | ✓ PASS |
| Hyperlink basic | `go test -run TestHyperlink_Basic -v` | PASS | ✓ PASS |
| Hyperlink formatting | `go test -run TestHyperlink_FormattingChaining -v` | PASS | ✓ PASS |
| Hyperlink multiple | `go test -run TestHyperlink_MultipleCalls -v` | PASS | ✓ PASS |
| Header/footer round-trip | `go test -run TestRoundTrip_HeaderFooter -v` | PASS | ✓ PASS |
| Full content doc | `go test -run TestFullContentDoc -v` | PASS | ✓ PASS |
| All fixtures round-trip | `go test -run TestRoundTrip_AllFixtures -v` | PASS | ✓ PASS |

### Probe Execution

No probes declared in PLAN/SUMMARY. n/a.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-----------|-------------|--------|----------|
| API-04 | 05-01 | Tables — rows, cells, shading, borders, widths, hMerge/vMerge, named table styles | ✓ SATISFIED | `table.go` — AddTable/AddTableBuilder, TableBuilder/RowBuilder/CellBuilder with all features |
| API-05 | 05-01 | Images — PNG/JPEG with DrawingML anchors, explicit sizing | ✓ SATISFIED | `image.go` — AddImage/AddImageBytes, DPI detection, DrawingML inline, SetImageWidth/SetImageHeight |
| API-06 | 05-03 | Lists — ordered/bulleted, multi-level, numbering-definition-backed | ✓ SATISFIED | `list.go` — AddList, AddListFromSlice, AddNumberingDef, 9-level depth, template merge |
| API-07 | 05-02 | Headers/footers — default, first-page, odd/even variants per section | ✓ SATISFIED | `header.go` — AddHeader/AddFooter with HeaderVariant/FooterVariant, OPC parts, sectPr linking |
| API-08 | 05-03 | Hyperlinks on runs | ✓ SATISFIED | `hyperlink.go` — Paragraph.AddHyperlink creates CT_Hyperlink with External relationship |
| API-09 | 05-02 | Page setup — margins, orientation, paper size; page breaks | ✓ SATISFIED | `page.go` — SetOrientation, SetPaperSize, SetMargins, AddPageBreak, Section wrapper |

### Anti-Patterns Found

| File | Pattern | Severity | Status |
|------|---------|----------|--------|
| (none) | — | — | No TODO/FIXME/XXX/placeholder markers found in any phase 5 files |

### Human Verification Required

None. All truths verified through codebase evidence.

### Gaps Summary

No gaps found. Phase goal fully achieved.

## Detailed Verification

### Tables (API-04)

- **AddTable(data [][]string):** Creates CT_Tbl with TblPr, TblGrid, GridCol, cells. Appends to Body.Tbl. Returns TableBuilder for further customization.
- **AddTableBuilder():** Returns empty TableBuilder with CT_Tbl pre-appended to Body.Tbl.
- **TableBuilder:** SetTableStyle(name), SetWidth(w, type), SetBorders(b), SetShading(val, fill), Row(idx) grows slice, X() escape hatch.
- **RowBuilder:** SetBorders(b) applies to all cells, Cell(idx) grows slice.
- **CellBuilder:** SetText(text), SetShading(val, fill), SetBold(b), SetWidth(w, type), MergeRight() (GridSpan=2), MergeDown() (VMerge=restart).
- **Tables-at-end:** CT_Body struct serializes `P` (paragraphs) before `Tbl` (tables), guaranteeing paragraphs-first output order.

### Images (API-05)

- **AddImage(path):** os.ReadFile → detectContentType → AddImageBytes.
- **AddImageBytes(name, data, ct):** MarkModified for word/media/imageN.ext, content type override, relImage relationship from document.xml, DPI detection (JPEG JFIF APP0, PNG pHYs), auto-size 3in default, builds full DrawingML inline (wp:inline → a:graphic → pic:pic → a:blipFill → a:blip → r:embed), appends run with drawing to last paragraph.
- **Run.SetImageWidth/SetImageHeight:** Mutates CT_Drawing.Inline.Extent.Cx/Cy and SpPr.Xfrm.Ext Cx/Cy.

### Headers/Footers (API-07)

- **HeaderVariant/FooterVariant:** Enums (Default, First, Even) with String() output.
- **AddHeader(variant):** Creates CT_Hdr, serializes, addHelperPart creates word/headerN.xml part with relHeader relationship + ctHeader content type override, appends CT_HdrFtrRef to sectPr with variant type + rId.
- **AddFooter(variant):** Symmetric using relFooter/ctFooter, appends FtrRef.
- **Header/Footer.AddParagraph(text):** Appends CT_P to header/footer with text run.
- **OpenTemplate clone:** Scans source sectPr HdrFtrRef/FtrRef, resolves rel target, copies part bytes, allocates fresh rIds via dstRels.NextRID(), adds content type overrides, creates fixed-up refs.

### Page Setup (API-09)

- **PageOrientation:** Portrait/Landscape enum.
- **PaperSize constants:** LetterW=12240/H=15840, A4W=11906/H=16838, LegalW=12240/H=20160.
- **SetOrientation:** Swaps W/H when W<H for landscape, W>H for portrait.
- **SetPaperSize:** Sets PgSz.W/H in twips.
- **SetMargins:** Sets PgMar Top/Right/Bottom/Left in twips.
- **Section wrapper:** Forward-compatible API for v1 single-section.
- **AddPageBreak:** Creates paragraph with PageBreakBefore = &CT_OnOff{}.
- **Paragraph.SetPageBreakBefore(bool):** Explicit control.

### Lists (API-06)

- **AddList(ordered):** Scans existing numbering.xml, creates CT_AbstractNum with 9 levels (decimal/bullet numFmt, `%N.`/Unicode bullet chars per level), creates CT_Num linking to it, writes numbering.xml. Returns ListBuilder.
- **AddListFromSlice(items, ordered):** Convenience: AddList + AddItem per item at level 0.
- **AddNumberingDef(fmt, start):** Custom numFmt across all 9 levels.
- **ListBuilder.AddItem(text, level):** Creates paragraph with PPr.NumPr (numId + ilvl). Level bounds 0-8 enforced.
- **ID collision avoidance:** findMaxAbstractNumID/findMaxNumID scan all existing entries. numId=0 reserved for Word ListNumber.

### Hyperlinks (API-08)

- **Paragraph.AddHyperlink(text, uri):** Allocates fresh rId via NextRID(), creates relationship with Type=relHyperlink + TargetMode="External", creates CT_Hyperlink with r:id and CT_R child, returns *Run for formatting chaining.
- No URI dedup — each call creates new rId (~50 bytes overhead).

### Shared Infrastructure

- **CT_P.Hyperlink field:** `internal/wml/document.go:56` — `Hyperlink []*CT_Hyperlink`.
- **CT_R.Drawing field:** `internal/wml/document.go:87` — `Drawing *CT_Drawing`.
- **Constants:** `create.go:46-56` — relHeader/Footer/Image/Hyperlink, ctHeader/Footer/Png/Jpeg.
- **Sequence counters:** `nextImageID`, `nextHeaderID`, `nextFooterID` initialized to 1 in all constructors (Create, OpenReader, FromTemplateReader, OpenTemplateReader).
- **Threat model mitigations:**
  - T-05-01 (path traversal): os.ReadFile via AddImage — mitigated by Go stdlib.
  - T-05-02 (EMU overflow): int64 max / 914400 > 1e13 inches — accepted.
  - T-05-03 (decompression bomb): OPC-07 MaxPartBytes (128MB) — mitigated.
  - T-05-04 (grid vs cell mismatch): missing cells render empty in Word; Warnings().
  - T-05-05 (rId collision): dstRels.NextRID() for fresh allocation — mitigated.
  - T-05-06 (CT override): content type overrides for cloned parts — mitigated.
  - T-05-07 (orientation swap): Word reads orientation from W/H ratio — accepted.
  - T-05-08 (hyperlink URI injection): encoding/xml escapes XML — accepted.
  - T-05-09 (numId exhaustion): ~2B lists to exhaust int64 — accepted.
  - T-05-10 (template numId collision): scans all existing IDs — mitigated (Pitfall 3).
  - T-05-SC (package installs): zero external deps — mitigated.

---

_Verified: 2026-07-26T10:30:00Z_
_Verifier: the agent (gsd-verifier)_
