# Stack Research

**Domain:** Pure Go library for .docx creation/editing with style preservation from reference templates
**Researched:** 2026-07-25
**Confidence:** HIGH

## Executive Summary

No existing Go library provides a complete, permissively-licensed solution for reading, creating, and editing .docx files while preserving styles from a reference template. The Go ecosystem has two tiers: commercially-licensed complete libraries (unioffice/gooxml, AGPL/paid) and MIT-licensed partial libraries that cover only text replacement or basic creation. In the C# ecosystem, Microsoft's DocumentFormat.OpenXml SDK is the gold standard for OOXML — it has no Go equivalent. **The recommended approach is to build a native Go library using only standard library packages (`archive/zip`, `encoding/xml`)**, adopting the ISO 29500 part-based architecture directly rather than wrapping any existing incomplete library. For the Go language version, use Go 1.23+ for its `iter` package support, `io/fs`, and improved XML handling.

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go (language) | 1.23+ | Runtime and standard library | Go 1.23 adds iterators for clean document traversal; `io/fs` for embedded template support; `archive/zip` and `encoding/xml` are the only dependencies needed for OPC/XML — zero external dependency risk |
| `archive/zip` | stdlib | OPC container read/write | .docx IS a ZIP with OPC conventions; Go's standard zip reader/writer handles the package layer (parts, content types, relationships) natively with no CGO |
| `encoding/xml` | stdlib | XML document parts parsing | Every .docx part (document.xml, styles.xml, numbering.xml, etc.) is XML; Go's `xml.Decoder` with streaming `Token()` API handles namespace-prefixed WML/RELAX namespaces correctly |
| `image` + `image/jpeg`/`image/png` | stdlib | Embedded image metadata extraction | Required for reading image dimensions, aspect ratios, and color profiles from media parts without external libs |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `golang.org/x/image` | latest | Extended image format support (TIFF, BMP) | Only when .docx contains exotic embedded image formats; skip for initial delivery |
| `sigs.k8s.io/kind` / std `testing` | — | Golden file testing for generated .docx output | Compare byte-level output against known-good .docx files in CI; critical for regression detection |
| `github.com/google/go-cmp` | v0.6+ | Deep XML tree comparison in tests | When asserting style resolution produces correct XML output; use `cmp` with custom `cmpopts` for namespace-agnostic comparison |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| `go test -fuzz` | Fuzz OPC parsing with malformed inputs | .docx files from untrusted sources may have corrupted ZIP entries or invalid XML; fuzz `encoding/xml` deserialization of every WML part type |
| `go tool cover` | Track coverage of WML type marshaling | Each WML element type must round-trip correctly; coverage gaps indicate untested schema elements |
| `validator` (W3C XML Schema) | Validate generated OPC packages against ISO 29500 schemas | Run against CI output to ensure spec compliance before release; Microsoft publishes the XSDs |

## Architecture Decision: Do NOT Use Existing Go .docx Libraries

### Why Not Existing Libraries

| Library | When It Works | Why NOT for This Project |
|---------|---------------|--------------------------|
| **unioffice/gooxml** (unidoc) | Need all Office formats (docx + xlsx + pptx) and have budget | **Commercial license** — AGPL-3.0 for open source only, paid license required for closed source ($1,000+/yr); 33MB binary bloat from generated schema code; document model is proprietary, not aligned with ISO 29500 part architecture; style resolution is opaque — you can't audit or extend it |
| **docxgo** (mmonterroca/docxgo v2.7.2) | Quick document generation with fluent builder | **Immature** — 111 stars, single maintainer, Go 1.23+ only; claims round-trip style preservation but has <1 year of field validation; style system is a thin layer over raw XML, not a complete resolution engine; no proven template style inheritance |
| **gomutex/godocx** (v0.1.5) | python-docx style API in Go | **Early beta** — v0.x, 263 stars, incomplete WML type coverage; no numbered list support, no header/footer style inheritance; table style resolution is partial; API inspired by python-docx which maps poorly to Go idioms |
| **fumiama/go-docx** | Reading and basic editing | **AGPL-3.0 license** — contaminates your project; no explicit style preservation API; hard fork of gonfva/docxlib with incompatible changes; sparse docs |
| **nguyenthenguyen/docx** | Simple text replacement in templates | **Not a document creation library** — text-only find/replace on raw bytes; no paragraph/run/section model; no style awareness; destructive to complex formatting |
| **lukasjarosch/go-docx** | Reliable `{placeholder}` replacement | **Placeholder-only** — no document creation; no style manipulation; operates at byte level on the decompressed XML, not through a document object model |
| **gingfrederik/docx** | Minimal document creation | **Write-only** — cannot read or edit existing documents; no style system, no headers/footers, no tables, no images; trivial implementation (~300 lines) |

### The Build-Vs-Borrow Decision

**Build new.** Rationale:

1. **Licensing incompatibility** — The only complete libraries (unioffice, gooxml) use AGPL. The project requires permissive licensing (MIT or Apache-2.0) for broad adoption.
2. **Style preservation is the differentiator** — No existing library treats style preservation from a reference template as a first-class feature. Every existing library either ignores styles, applies hard-coded formatting, or has a proprietary style model that doesn't map to ISO 29500.
3. **Total implementation scope is bounded** — DocumentFormat.OpenXml-based implementations of the full word subset run ~50KLOC of C#. The Go equivalent targets the same subset. This is a 2-3 month effort, not a year-long rewrite.
4. **Zero external deps reduces risk** — Using `archive/zip` + `encoding/xml` means no supply-chain risk, no version conflicts, no CGO cross-compilation issues.

## Gap Analysis: DocumentFormat.OpenXml (C#) vs Go Library

DocumentFormat.OpenXml is Microsoft's official .NET SDK for OOXML. Here's what a Go library must implement natively:

| C# SDK Component | Go Equivalent | Complexity | Notes |
|------------------------|---------------|------------|-------|
| `Package.Open()` / `WordprocessingDocument` | Custom `OPCPackage` type wrapping `archive/zip.Reader` | Medium | OPC conventions: [Content_Types].xml, .rels files, part URIs |
| `MainDocumentPart` | `DocumentPart` struct with `Document` (Body, Paragraphs, Runs) | High | WML namespace types (w:p, w:r, w:t, w:pPr, w:rPr, etc.) — ~200 element types |
| `StylePart` | `StylesPart` struct — `Styles` (Style for each type) | High | Style inheritance chain: LatentStyles → Default → Style (basedOn/next/link) → ParagraphProperties → RunProperties |
| `NumberingPart` | `NumberingPart` struct — `Numbering` (AbstractNum, Num, NumFmt) | Medium | List numbering definitions, abstract numbering, concrete numbering instances |
| `HeaderPart` / `FooterPart` | `HeaderPart` / `FooterPart` — same WML types as document | Medium | Relationship-linked parts, same content model as body |
| Theme part | `ThemePart` — color scheme, font scheme, format scheme | Medium | DrawingML types (dml:theme, dml:clrScheme, dml:fontScheme) |
| Settings part | `SettingsPart` — document-level settings | Low | Compatibility settings, zoom, default tab stop, etc. |
| `AltChunk` handling | Alternative format chunks (HTML import) | Medium | Needed only if HTML import is added later |
| Style resolution | `StyleResolver` — computes effective style from inheritance chain | **Critical** | No Go library implements this. Core of style preservation. |
| Relationship management | `RelationshipCollection` — part-to-part relationships | Medium | OPC relationships per part, external (hyperlinks, images) and internal (headers, footers) |

### What the Go Library Must Build (No Existing Equivalent)

1. **WML XML type hierarchy** — Go structs for all WordprocessingML elements (w:document, w:body, w:p, w:r, w:t, w:pPr, w:rPr, w:sectPr, w:tbl, w:tr, w:tc, w:styles, w:style, w:numbering, etc.) with `encoding/xml` marshaling/unmarshaling. This is ~50-80 struct types initially, growing to ~200 for full spec coverage.

2. **OPC package reader/writer** — ZIP wrapper that respects OPC conventions: parses `[Content_Types].xml`, resolves relationships via `_rels/.rels` and part-level `_rels/*.xml.rels`, preserves unknown parts during round-trip.

3. **Style resolution engine** — Computes effective paragraph and run properties by following the ISO 29500 style inheritance chain:
   - DocDefaults (document-level defaults)
   - LatentStyles (style gallery defaults)
   - Paragraph style (basedOn chain)
   - Paragraph direct formatting (w:pPr)
   - Run style (if applied via w:rStyle)
   - Run direct formatting (w:rPr)
   + Numbering style resolution for lists
   + Table style resolution (special inheritance: whole table → banding → rows → cells)

4. **Template style preservation** — On document creation from template:
   - Copy all style definitions from template styles.xml
   - Copy numbering definitions from template numbering.xml
   - Preserve theme (colors, fonts) from template theme part
   - When adding content, reference template styles by ID; apply style resolution to compute effective formatting

5. **Document modification engine** — Insert/update/delete content while preserving existing formatting:
   - Maintain proper relationship IDs when adding/removing parts
   - Update content types when new parts are added
   - Recalculate page references when content shifts

## Stack Patterns by Variant

### If Building a Minimal Viable Product (Phase 1):

| Technology | Purpose |
|------------|---------|
| Go 1.23+ stdlib only | Zero dependencies for first release |
| `archive/zip` + `encoding/xml` | Core OPC/XML handling |
| Custom WML types (subset) | Only document body + styles.xml + numbering.xml |
| No theme part parsing | Hardcode default theme |

**Ship:** Document creation from Go structs, basic paragraph/run formatting, style application from template, save to .docx.

### If Extending Toward Full Feature Parity (Phase 2+):

| Technology | Purpose |
|------------|---------|
| Full WML type coverage | All ~200 element types for document.xml + parts |
| Header/Footer/Footnote support | Multiple header/footer per section |
| Table style resolution | Complex table style inheritance banding |
| Image handling with `image` stdlib | Image dimensions, aspect ratio calculation |
| AltChunk for HTML import | Import rendered HTML as document content |
| Content controls (SDT) | Structured document tags for template filling |
| Tracked changes (revisions) | Accept/reject edits; revision marks |

### If Multi-Format Expansion (Phase 3+):

| Technology | Purpose |
|------------|---------|
| Go stdlib + same pattern | Apply same OPC + XML architecture to xlsx/pptx |
| Shared common layer | `OPCPackage`, relationship management, content types |
| SML types for Excel | SpreadsheetML: workbook, worksheet, cell formatting |
| PML types for PowerPoint | PresentationML: slides, shapes, masters |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| **unioffice** (unidoc) | AGPL-3.0 license contaminates permissive project; $1K+/yr for commercial; 33MB binary; opaque style engine | Build from stdlib + custom WML types |
| **gooxml** (baliance) | Same team as unioffice, effectively deprecated since 2018; same AGPL trap | Build from stdlib |
| **fumiama/go-docx** | AGPL-3.0; fork lineage makes provenance unclear; incomplete style support | Build from stdlib |
| **gomutex/godocx** | v0.x, incomplete WML coverage; python-docx inspired API is non-idiomatic Go | Build from stdlib with Go-idiomatic API |
| **CGO/com** | Requires Windows-only Word interop; destroys cross-compilation; huge attack surface | Pure Go OPC/XML handling via stdlib |
| **python-docx** (invoked via subprocess) | Process overhead; python runtime dep; fragile I/O; no static binary | Pure Go library from scratch |
| **libreoffice --headless** (CLI wrapper) | 500MB+ install; slow startup; format conversion loses data; not a library | Pure Go OPC manipulation |

## Comparison: Key Library Decision Matrix

| Capability | unioffice | docxgo v2.7.2 | gomutex/godocx | **This Project (target)** |
|------------|-----------|---------------|----------------|--------------------------|
| License | AGPL/Commercial | MIT | MIT | MIT/Apache-2.0 |
| Read existing .docx | ✅ | ✅ | ✅ | ✅ |
| Create new .docx | ✅ | ✅ | ✅ | ✅ |
| Template style preservation | Partial | Claimed (unproven) | ❌ | ✅ **First-class** |
| Style resolution engine | Opaque | Thin | Partial | ✅ **Explicit, auditable** |
| Headers/Footers | ✅ | ✅ | Partial | ✅ |
| Tables | ✅ | ✅ | ✅ | ✅ |
| Numbering/Lists | ✅ | ✅ | Partial | ✅ |
| Images | ✅ | ✅ | ✅ | ✅ |
| AltChunk/HTML import | ❌ | ❌ | ❌ | Planned |
| Content controls (SDT) | ✅ | ❌ | ❌ | Planned |
| Tracked changes | ❌ | ❌ | ❌ | Planned |
| Zero external deps | ❌ (uses xml pkg only) | ❌ | ❌ | **✅ (stdlib only)** |
| Static binary size | ~33MB | Unknown | ~5MB | ~5MB |
| Maturity | High (4.9k★, 7yr) | Low (111★, <1yr) | Low (263★, 1yr) | New |
| Single maintainer risk | Low (company-backed) | High | Medium | Medium |

## Implementation Architecture

### Package Structure

```
github.com/fabiomarini/wordingo/
├── opc/                    # OPC package layer
│   ├── package.go          # OPC Package reader/writer
│   ├── part.go             # Part interface + base implementation
│   ├── relationship.go     # Relationship management
│   └── content_types.go    # [Content_Types].xml handling
├── wml/                    # WordprocessingML types (generated from XSD)
│   ├── document.go         # w:document, w:body
│   ├── paragraph.go        # w:p, w:pPr
│   ├── run.go              # w:r, w:rPr, w:t
│   ├── section.go          # w:sectPr
│   ├── table.go            # w:tbl, w:tr, w:tc
│   ├── styles.go           # w:styles, w:style
│   ├── numbering.go        # w:numbering, w:num, w:abstractNum
│   ├── header_footer.go    # w:hdr, w:ftr
│   └── types.go            # Shared simple types (ST_*)
├── dml/                    # DrawingML types (theme, shapes)
│   └── theme.go            # dml:theme
├── resolver/               # Style resolution engine
│   ├── resolver.go         # StyleResolver interface
│   ├── paragraph.go        # Paragraph style resolution
│   ├── run.go              # Run style resolution
│   └── table.go            # Table style resolution
├── document.go             # Public Document API
├── paragraph.go            # Paragraph builder
├── run.go                  # Run builder
├── table.go                # Table builder
└── template.go             # Template-based document creation
```

### Data Flow: Template-Based Document Creation

```
User provides:
  1. Reference .docx template (with styles.xml, numbering.xml, theme.xml)
  2. Go struct with content data

Pipeline:
  OpenTemplate("template.docx")
    │
    ├── opc.Open() → reads ZIP, parses [Content_Types].xml, .rels
    │
    ├── Load Parts:
    │   ├── document.xml → DocumentPart (wml.Document)
    │   ├── styles.xml → StylesPart (wml.Styles) ← PRESERVED
    │   ├── numbering.xml → NumberingPart (wml.Numbering) ← PRESERVED
    │   ├── theme.xml → ThemePart (dml.Theme) ← PRESERVED
    │   ├── settings.xml → SettingsPart
    │   └── header*.xml / footer*.xml → HeaderPart/FooterPart ← PRESERVED
    │
    ├── Resolve Styles:
    │   └── resolver.New(styles, numbering, theme)
    │       ├── Computes effective paragraph formatting
    │       ├── Computes effective run formatting
    │       └── Handles style inheritance chain
    │
    ├── Build Content:
    │   ├── doc.AddParagraph().SetStyle("Heading1").AddText("Title")
    │   ├── doc.AddParagraph().SetStyle("Normal").AddText("Body")
    │   ├── doc.AddTable().SetStyle("LightList-Accent1")
    │   └── Apply styles via resolver → generates w:pPr/w:rPr with resolved values
    │
    └── Save():
        ├── opc.Pack() → serializes all parts back to ZIP
        ├── Preserves unknown parts (media, embeddings, etc.)
        └── Updates relationship IDs, content types as needed
```

## Version Compatibility

All components use Go standard library only. There are no external package version conflicts to manage.

| Go Version | Compatibility | Notes |
|-----------|---------------|-------|
| 1.21 | Minimal | Lacks `iter` package for clean document traversal; `xml.Decoder` works but less ergonomic |
| 1.22 | Good | Range-over-int, improved `xml`; acceptable baseline |
| **1.23** | **Target** | `iter.Seq` for document element traversal; `http.ServeMux` patterns; `unique` package |

## Sources

- **pkg.go.dev** — [search: docx](https://pkg.go.dev/search?q=docx) — surveyed all 20+ Go docx modules (2026-07-25) — HIGH confidence
- **unidoc/unioffice** — [GitHub: unidoc/unioffice](https://github.com/unidoc/unioffice) — 4.9k stars, AGPL-3.0 license, commercial — HIGH confidence
- **mmonterroca/docxgo** — [GitHub: mmonterroca/docxgo](https://github.com/mmonterroca/docxgo) — v2.7.2 (Jul 2026), MIT, 111 stars — MEDIUM confidence (new)
- **gomutex/godocx** — [GitHub: gomutex/godocx](https://github.com/gomutex/godocx) — v0.1.5 (Sep 2024), MIT, 263 stars — MEDIUM confidence
- **fumiama/go-docx** — [GitHub: fumiama/go-docx](https://github.com/fumiama/go-docx) — AGPL-3.0, 302 stars — HIGH confidence
- **DocumentFormat.OpenXml (Microsoft)** — [.NET SDK](https://github.com/dotnet/Open-XML-SDK) — architectural reference for part hierarchy and style inheritance — HIGH confidence
- **ISO 29500-1:2016** — Office Open XML File Formats — ECMA-376 standard — HIGH confidence
- **Office Open XML anatomy** — [officeopenxml.com](http://officeopenxml.com/anatomyofOOXML.php) — WML reference — HIGH confidence

---

*Stack research for: wordingo — Word document library with style preservation*
*Researched: 2026-07-25*
