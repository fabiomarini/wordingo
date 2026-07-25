# PRD: wordingo — Pure Go Word Document Library

**Status:** Draft v1
**Date:** 2026-07-25
**License:** MIT
**Language:** Go 1.23+ (stdlib only, zero external dependencies)

---

## 1. Vision

A pure Go library, importable by any Go application, that creates Microsoft Word (.docx) documents from scratch or from existing templates — blank or pre-populated — while preserving the styles of every document object: titles, headings, paragraphs, runs, tables, lists, headers, and footers.

No C# runtime. No COM interop. No LibreOffice wrapper. No external Go dependencies. One `go get` and it works.

```go
import "github.com/fabiomarini/wordingo"

doc, _ := wordingo.FromTemplate("brand-template.docx")
doc.AddParagraph("Quarterly Report").Style("Title")
doc.AddParagraph("Revenue grew 25%").Style("Normal")
doc.Save("q3-report.docx")
// q3-report.docx opens in Word with the template's fonts, colors, and heading styles intact.
```

---

## 2. Problem Statement

Go has no production-grade, permissively-licensed library for Word document creation with style preservation.

| Existing option | License | Fatal flaw |
|---|---|---|
| unioffice/unidoc | AGPL / commercial $1K+/yr | License incompatible with commercial and permissive OSS use |
| gomutex/godocx | MIT | v0.x, no style resolution, no template support |
| mmonterroca/docxgo | MIT | 111 stars, single maintainer, unproven style claims |
| nguyenthenguyen/docx | MIT | Text replacement only; no document model, no styles |
| fumiama/go-docx | AGPL | License contaminates importers |

Consequence: Go teams shell out to LibreOffice, run python-docx via subprocess, or pay unioffice. All three are operational liabilities — slow startup, runtime dependencies, or licensing cost.

The missing capability is not "write XML to a ZIP." Any library can do that. The missing capability is **fidelity**: documents that open in Word looking exactly as the template author designed — correct fonts, heading styles, colors, table styles, list numbering — because the style system (styles.xml, numbering.xml, fontTable.xml, theme.xml) was handled correctly.

---

## 3. Goals

### G1 — Create from scratch
Generate a minimal, valid .docx containing sensible default styles, theme, and settings. Opens in Word 2016+ and LibreOffice without repair dialogs.

### G2 — Create from template (blank or pre-populated)
Clone a reference .docx's complete style dependency graph (styles, numbering, font table, theme, settings) into a new document. New content references the template's named styles and renders identically.

### G3 — Style preservation on edit
Open an existing .docx, modify content, save. All untouched parts, styles, and formatting survive byte-identically. Unknown elements and unmodeled parts pass through untouched.

### G4 — Style-aware content API
Paragraphs, runs, tables, lists, images, headers/footers created through the API apply named styles resolved through the full OOXML inheritance chain (docDefaults → latentStyles → basedOn → direct formatting).

### G5 — Template merge
Replace `{{placeholder}}` markers in paragraphs and table cells with data, preserving surrounding run formatting.

### G6 — Zero dependencies, MIT license
Go standard library only (`archive/zip`, `encoding/xml`, `image/*`). MIT license. Static binaries, cross-compiles everywhere, importable into any project without license review.

---

## 4. Non-Goals

| Excluded | Reason |
|---|---|
| PDF conversion | Layout engine is a separate product; use pandoc/LibreOffice |
| .doc (binary OLE2) | Different spec entirely; convert via LibreOffice first |
| .odt / .rtf / .pptx / .xlsx | Focus. OPC layer designed for reuse, but formats are out of scope for v1 |
| HTML rendering / live preview | A viewer is a separate product, not a library concern |
| CLI tool | This is a library. Downstream users build their own CLIs |
| Word COM automation / CGO | Defeats the purpose |
| Collaborative editing | Single-writer document processing |
| VBA / macros | Security surface; preserved as opaque parts on round-trip only |
| Full OOXML spec coverage | ~200 WML types modeled; unknown elements preserved, not parsed |

---

## 5. Users and Use Cases

### Primary user: Go backend developer building document generation

**UC1 — Branded report generation (template + merge)**
A service generates weekly client reports. Designer maintains `report-template.docx` in Word with corporate styles. Service opens template, merges `{{client}}`, `{{period}}`, `{{total}}` into placeholders, appends data tables using template table styles, saves. Output is indistinguishable from hand-authored.

**UC2 — Document from scratch with defaults**
A CI pipeline generates release notes as .docx. No template available. Library creates a blank document with professional default styles (Calibri, heading hierarchy, default theme), adds headings, paragraphs, bullet lists, saves.

**UC3 — Edit existing document**
A compliance tool opens an existing contract .docx, replaces one clause paragraph, saves. Every other part — embedded fonts, custom XML, tracked changes markup, glossary entries — survives untouched.

**UC4 — Pre-populated template**
An HR system uses a template that already contains boilerplate sections, headers with logos, and footers with page numbers. The library opens it, inserts offer details into specific locations, saves. Headers, footers, and section formatting preserved.

---

## 6. Functional Requirements

### FR-1: OPC Package Layer (`opc`)
The foundation. .docx is a ZIP of XML parts connected by relationships.

- **FR-1.1** Open .docx as OPC package: parse `[Content_Types].xml`, `_rels/.rels`, per-part `.rels`
- **FR-1.2** Create valid OPC packages with canonical ZIP entry ordering (content types first, rels second, per ECMA-376 Part 2 §9.1.4.2)
- **FR-1.3** Namespace handling by URI, not prefix — files from Word, LibreOffice, Google Docs parse identically (40+ OOXML namespaces)
- **FR-1.4** Pass-through preservation: unmodeled parts stored as raw bytes, re-emitted identically on save
- **FR-1.5** Read both Transitional and Strict conformance; write Transitional by default; match source conformance when editing
- **FR-1.6** Relationship graph integrity: unique rId generation, no reuse, consistency validation on save
- **FR-1.7** Safety limits: decompression bomb protection, entity expansion limits, path traversal rejection, max parts limit

### FR-2: WML Schema Types (`wml`)
- **FR-2.1** Hand-written Go structs for ~60 essential WordprocessingML types: document, body, paragraph, run, text, paragraph properties, run properties, styles, numbering, tables, sections, headers, footers
- **FR-2.2** Bidirectional marshal/unmarshal with correct namespace URIs
- **FR-2.3** Whitespace fidelity: `xml:space="preserve"` handled on read and write; run boundaries never merged implicitly
- **FR-2.4** Unknown child elements hoarded (`xml:",any"`-style capture) and re-emitted on save

### FR-3: Style Engine (`style`)
Three distinct operations, explicitly separated:

- **FR-3.1 STYLE-CLONE** — Copy a template's complete style dependency graph into a target document: styles.xml + numbering.xml + fontTable.xml + theme.xml + settings.xml, with relationships and content types updated
- **FR-3.2 STYLE-RESOLVE** — Compute effective paragraph and run properties by walking the inheritance chain: docDefaults → latentStyles → style (basedOn chain, cycle-detected) → paragraph direct formatting → run style → run direct formatting. Includes theme color resolution and numbering definition resolution
- **FR-3.3 STYLE-ROUNDTRIP** — Editing an existing document never rewrites style parts unless the user explicitly modifies styles. Formatting of unmodified content is byte-identical after save

### FR-4: Blank Document Creation
- **FR-4.1** `Create()` generates a minimal valid .docx: default styles (Normal, Heading 1–9, Title), default theme, default font table, empty body with one section
- **FR-4.2** Generated document passes Word validation without repair dialog

### FR-5: Template Document Creation
- **FR-5.1** `FromTemplate(path)` clones the template package, preserves all style/theme/numbering parts, and provides a body ready for content
- **FR-5.2** Pre-populated template support: open template, keep its existing body content, allow insertion at specified locations
- **FR-5.3** Named styles from the template are applicable to new content by name (`SetStyle("Heading1")`)

### FR-6: Content API (`wordingo` root package)
Fluent, Go-idiomatic builders over WML types:

- **FR-6.1** Paragraphs with runs: text, bold, italic, underline, font family, size, color, highlight
- **FR-6.2** Paragraph formatting: alignment, spacing before/after, line spacing, indentation, keep-with-next
- **FR-6.3** Named style application to paragraphs and runs, resolved via the style engine
- **FR-6.4** Tables: rows, cells, cell shading, borders, column widths, merged cells (hMerge/vMerge), named table styles
- **FR-6.5** Images: PNG/JPEG embedding with DrawingML anchors, explicit sizing
- **FR-6.6** Lists: ordered and bulleted, backed by numbering definitions, multi-level
- **FR-6.7** Headers and footers: default, first-page, and odd/even variants per section
- **FR-6.8** Hyperlinks on runs (external relationships)
- **FR-6.9** Page setup: margins, orientation, paper size per section
- **FR-6.10** Page breaks

### FR-7: Template Merge
- **FR-7.1** `Merge(data map[string]string)` replaces `{{key}}` in paragraphs, table cells, headers, footers
- **FR-7.2** Placeholders split across multiple runs (Word does this arbitrarily) are detected and merged correctly, preserving the formatting of the first run fragment
- **FR-7.3** Missing keys produce a warning list (`doc.Warnings()`), never silent corruption

### FR-8: Editing Existing Documents
- **FR-8.1** Insert paragraph before/after a target paragraph
- **FR-8.2** Delete paragraph / delete table row
- **FR-8.3** Replace text within a run without disturbing adjacent run formatting
- **FR-8.4** All edits honor STYLE-ROUNDTRIP (FR-3.3)

### FR-9: I/O Flexibility
- **FR-9.1** Open from `io.ReaderAt`, save to `io.Writer` — file paths are convenience, not requirement
- **FR-9.2** No panics in library code; all failures returned as errors
- **FR-9.3** `doc.Warnings() []string` exposes non-fatal issues (dropped nothing, unresolved keys, unknown relationship types logged but preserved)

---

## 7. Non-Functional Requirements

| Requirement | Target |
|---|---|
| Dependencies | Zero external. Stdlib only |
| License | MIT |
| Go version | 1.23+ |
| Binary size contribution | < 5 MB (wordprocessingML types only, no xlsx/pptx codegen) |
| Open performance | < 50 ms for a 100-page document (lazy part loading) |
| Round-trip fidelity | Unmodified parts byte-identical; modified parts semantically identical |
| Compatibility | Word 2016, 2019, 2021, M365, LibreOffice 7+, Google Docs (import) |
| Conformance | ISO/IEC 29500 Transitional on write; read Transitional + Strict |
| Thread safety | Individual Document not shared across goroutines; separate Documents are independent |
| Security | ZIP bomb, entity expansion, path traversal, SSRF (external relationships never auto-fetched) |

---

## 8. API Sketch

Single public package. Internal packages (`opc`, `wml`, `style`) are not imported by users.

```go
package wordingo

// Creation
func Create() (*Document, error)
func FromTemplate(path string) (*Document, error)
func FromTemplateReader(r io.ReaderAt, size int64) (*Document, error)

// Open existing
func Open(path string) (*Document, error)
func OpenReader(r io.ReaderAt, size int64) (*Document, error)

type Document struct { /* wraps opc.Package + style.Resolver */ }

func (d *Document) AddParagraph(text string) *Paragraph
func (d *Document) AddTable(rows, cols int) *Table
func (d *Document) AddImage(r io.Reader, format ImageFormat, widthEMU, heightEMU int) error
func (d *Document) Merge(data map[string]string) error
func (d *Document) Paragraphs() []*Paragraph          // body content access
func (d *Document) Warnings() []string
func (d *Document) Save(path string) error
func (d *Document) WriteTo(w io.Writer) (int64, error)
func (d *Document) Close() error

type Paragraph struct { /* wraps *wml.CT_P */ }
func (p *Paragraph) SetStyle(name string) *Paragraph
func (p *Paragraph) AddRun(text string) *Run
func (p *Paragraph) SetAlignment(a Alignment) *Paragraph
func (p *Paragraph) InsertAfter(text string) *Paragraph   // editing

type Run struct { /* wraps *wml.CT_R */ }
func (r *Run) SetBold(b bool) *Run
func (r *Run) SetItalic(b bool) *Run
func (r *Run) SetFont(name string) *Run
func (r *Run) SetSize(pts float64) *Run
func (r *Run) SetColor(hex string) *Run
func (r *Run) SetText(s string) *Run

// Escape hatch: every wrapper exposes its WML type
func (p *Paragraph) X() *wml.CT_P
```

Design rules: builder chaining returns the receiver; zero values are valid; no constructors with 5+ arguments; errors from I/O and parsing, never panics.

---

## 9. Architecture

```
wordingo/            ← public API (Document, Paragraph, Run, Table, ...)
  ├── internal/opc/     ← ZIP + content types + relationships + namespaces
  ├── internal/wml/     ← ~60 WordprocessingML struct types (value objects)
  ├── internal/style/   ← clone, resolve, roundtrip engines
  └── internal/xmlutil/ ← URI-based namespace normalization over encoding/xml
```

Layering is strict: `opc` knows nothing about WML; `wml` knows nothing about OPC; `style` depends on both; public API orchestrates all three.

Key patterns (validated in research):
- **Wrapper-over-schema** — fluent API over raw WML structs; `X()` escape hatch
- **Lazy part loading** — parse parts on first access; raw bytes otherwise
- **Unknown element hoarding** — round-trip safety for unmodeled XML
- **URI-based namespace registry** — immune to prefix variation across producers

---

## 10. Success Metrics

| Metric | Target | How measured |
|---|---|---|
| Word opens generated docs without repair | 100% of test corpus | Golden-file CI against Word 2019/2021/M365 validation |
| Round-trip fidelity (unmodified parts) | Byte-identical | CI: open → save → diff parts |
| Template style fidelity | Rendered output visually matches template | Rendered-HTML comparison of same content in template vs output |
| Placeholder merge correctness | 100% of split-run placeholder cases | Dedicated test corpus of hostile templates |
| External dependencies | 0 | `go list -m all` = stdlib only |
| Adoption signal | Library importable and usable in < 10 lines | Getting-started example compiles and runs |

---

## 11. Risks and Mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| `encoding/xml` namespace fragility causes silent parse failures | High | URI-based registry built in Phase 1; test corpus spans Word/LibreOffice/Google Docs producers |
| Style resolution edge cases (circular basedOn, latent styles, theme-dependent props) break rendering | High | Style engine isolated in Phase 2 with real-template test corpus before any content API exists |
| Scope creep toward full OOXML feature surface | Medium | Non-goals are explicit; v2 features deferred behind validation gate |
| Zero-dep constraint makes XML layer expensive (~500 LOC custom) | Medium | Accepted consciously; documented trade-off. Revisit only if registry exceeds ~800 LOC |
| Unknown-part data loss on edit | Medium | Pass-through preservation in Phase 1, round-trip diff tests in CI from day one |
| Word version compatibility surprises | Medium | Compatibility matrix CI: Word 2016/2019/2021/M365 + LibreOffice + Google Docs import |

---

## 12. Milestones

| Milestone | Contents | Exit criteria |
|---|---|---|
| **M1 — Foundation** | opc package, wml types, blank document | Blank .docx opens in Word without repair |
| **M2 — Style Engine** | STYLE-CLONE + STYLE-RESOLVE | Doc created from template renders with template styles |
| **M3 — Round-trip** | Document model, open/read/save | Existing doc edited + saved with zero unintended diffs |
| **M4 — Content API** | Paragraphs, runs, formatting, named styles | Full text document created programmatically |
| **M5 — Rich Content** | Tables, images, headers/footers, lists, hyperlinks, page setup | Complete business document from template |
| **M6 — Merge & Edit** | {{placeholder}} merge, insert/delete, v1.0 release | UC1–UC4 all green; tagged v1.0.0 |

Post-v1.0 (deferred): field codes, multi-section docs, comments, bookmarks, footnotes, content controls, tracked changes, watermarks, charts, equations, Go struct→table mapping.

---

## 13. Design References

This project is built from scratch against the public ISO/IEC 29500 (OOXML) and ECMA-376 specifications. Conceptual references only:

- **ISO/IEC 29500 + ECMA-376** — authoritative spec for OPC packaging, WordprocessingML types, style inheritance, and conformance classes
- **Microsoft `DocumentFormat.OpenXml`** (.NET SDK) — reference for part architecture and the style inheritance model; studied for behavior, no code ported
- **`{{placeholder}}` merge UX** — adopted because it is the proven, designer-friendly interface for template-driven generation

wordingo is a Go library for applications that need documents as a library call — not a CLI, not a renderer, not a port of any existing tool.

---

*PRD authored 2026-07-25.*
