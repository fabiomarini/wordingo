# Requirements: wordingo — Pure Go Word Document Library

**Defined:** 2026-07-25
**Core Value:** Create styled .docx documents in Go — from a template or from scratch — that open in Word looking exactly as designed
**Source:** `.planning/PRD.md` (authoritative)

## v1 Requirements

### OPC Package

- [x] **OPC-01**: Open .docx ZIP as OPC package — parse [Content_Types].xml, _rels/.rels, per-part .rels
- [x] **OPC-02**: Write valid OPC packages with canonical ZIP entry ordering (ECMA-376 Part 2 §9.1.4.2)
- [x] **OPC-03**: Namespace resolution by URI across 40+ OOXML namespaces; files from Word, LibreOffice, Google Docs parse identically
- [x] **OPC-04**: Pass-through preservation — unmodeled parts stored as raw bytes, re-emitted identically on save
- [x] **OPC-05**: Read Transitional + Strict conformance; write Transitional; match source conformance when editing
- [x] **OPC-06**: Relationship graph integrity — unique rId generation, no reuse, consistency validation on save
- [x] **OPC-07**: Safety limits — decompression bomb protection, entity expansion limits, path traversal rejection, max parts limit

### WML Schema

- [x] **WML-01**: Go structs for ~60 essential WordprocessingML types (document, body, p, r, t, pPr, rPr, styles, numbering, tbl, sectPr, hdr, ftr)
- [x] **WML-02**: Bidirectional marshal/unmarshal with correct namespace URIs
- [x] **WML-03**: Whitespace fidelity — xml:space="preserve" honored on read/write; run boundaries never merged implicitly
- [x] **WML-04**: Unknown child elements hoarded and re-emitted on save

### Style Engine

Three distinct operations, explicitly separated. Do not conflate.

- [ ] **STYLE-CLONE-01**: Copy template style dependency graph into target document — styles.xml + numbering.xml + fontTable.xml + theme.xml + settings.xml, with relationships and content types updated
- [ ] **STYLE-CLONE-02**: Named styles from template applicable to new content by name
- [ ] **STYLE-RESOLVE-01**: Resolve effective paragraph properties through the full chain: docDefaults → latentStyles → basedOn chain → direct formatting
- [ ] **STYLE-RESOLVE-02**: Resolve effective run properties through the full chain, including run styles
- [ ] **STYLE-RESOLVE-03**: Resolve theme colors and numbering definitions; detect circular basedOn references
- [ ] **STYLE-ROUNDTRIP-01**: Editing a document never rewrites style parts unless user explicitly modifies styles
- [ ] **STYLE-ROUNDTRIP-02**: Unmodified content formatting is byte-identical after edit + save

### Document Creation

- [ ] **CREATE-01**: Create() generates minimal valid .docx — default styles (Normal, Heading 1–9, Title), default theme, default font table, empty body with one section
- [ ] **CREATE-02**: Generated blank document passes Word validation without repair dialog
- [ ] **CREATE-03**: FromTemplate(path) clones template package, preserves style/theme/numbering parts, provides ready body
- [ ] **CREATE-04**: Pre-populated template support — open template, keep existing body content, insert at specified locations

### Content API

- [ ] **API-01**: Paragraphs with runs — text, bold, italic, underline, font, size, color, highlight
- [ ] **API-02**: Paragraph formatting — alignment, spacing, line spacing, indentation
- [ ] **API-03**: Named style application to paragraphs and runs via style engine
- [ ] **API-04**: Tables — rows, cells, shading, borders, widths, hMerge/vMerge, named table styles
- [ ] **API-05**: Images — PNG/JPEG with DrawingML anchors, explicit sizing
- [ ] **API-06**: Lists — ordered/bulleted, multi-level, numbering-definition-backed
- [ ] **API-07**: Headers/footers — default, first-page, odd/even variants per section
- [ ] **API-08**: Hyperlinks on runs
- [ ] **API-09**: Page setup — margins, orientation, paper size; page breaks

### Template Merge

- [ ] **MERGE-01**: Replace {{key}} in paragraphs, table cells, headers, footers
- [ ] **MERGE-02**: Placeholders split across multiple runs detected and merged, preserving first-fragment formatting
- [ ] **MERGE-03**: Missing keys surfaced in Warnings(), never silent corruption

### Editing

- [ ] **EDIT-01**: Insert paragraph before/after target; delete paragraph; delete table row
- [ ] **EDIT-02**: Replace text within a run without disturbing adjacent formatting
- [ ] **EDIT-03**: All edits honor STYLE-ROUNDTRIP

### API Quality

- [ ] **QUAL-01**: Open from io.ReaderAt, save to io.Writer — paths are convenience only
- [ ] **QUAL-02**: No panics; all failures as errors; Warnings() for non-fatal issues
- [ ] **QUAL-03**: Single public package; X() escape hatch to WML types on every wrapper

## v2 Requirements

Deferred behind v1 validation.

- **V2-01**: Field codes (PAGE, DATE, NUMPAGES, TOC)
- **V2-02**: Multiple sections with independent page setup
- **V2-03**: Image positioning and text wrapping
- **V2-04**: Go struct → table mapping (reflection)
- **V2-05**: Comments, bookmarks, footnotes, endnotes
- **V2-06**: Content controls (SDT), form fields
- **V2-07**: Tracked changes, watermarks, charts, equations

## Out of Scope

| Feature | Reason |
|---------|--------|
| PDF conversion | Separate product |
| .doc / .odt / .rtf / .xlsx / .pptx | Focus on .docx |
| CLI / watch / HTML rendering / MCP | Not a library's job |
| COM automation / CGO | Defeats the purpose |
| Porting existing tools | ISO/IEC 29500 authoritative; DocumentFormat.OpenXml conceptual reference only; no code ported |
| Full OOXML spec | ~60 types modeled; unknown preserved, not parsed |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| OPC-01..07 | Phase 1 | Pending |
| WML-01..04 | Phase 1 | Pending |
| CREATE-01, CREATE-02 | Phase 1 | Pending |
| STYLE-CLONE-01..02 | Phase 2 | Pending |
| STYLE-RESOLVE-01..03 | Phase 2 | Pending |
| STYLE-ROUNDTRIP-01..02 | Phase 3 | Pending |
| CREATE-03, CREATE-04 | Phase 3 | Pending |
| API-01..03 | Phase 4 | Pending |
| API-04..09 | Phase 5 | Pending |
| MERGE-01..03 | Phase 6 | Pending |
| EDIT-01..03 | Phase 6 | Pending |
| QUAL-01..03 | All phases | Pending |

**Coverage:** 33 v1 requirements, all mapped ✓

---
*Requirements defined: 2026-07-25 after PRD rewrite*
