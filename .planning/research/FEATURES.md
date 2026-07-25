# Feature Research: Word Document (.docx) Library

**Domain:** Pure Go library for creating and editing Word .docx documents with style preservation
**Researched:** 2026-07-25
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features that any production-grade .docx library must provide. Missing these = unusable.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Read .docx content | Users need to extract text and structure from existing documents | HIGH | OPC (ZIP) parsing + XML namespace handling. Must handle complex relationships. |
| Create new .docx from scratch | Users need to generate documents programmatically | MEDIUM | Requires assembling valid OPC package with required parts (document.xml, styles.xml, relationships) |
| Write text with formatting (bold, italic, underline, font, size, color) | Core text formatting for any document generation | MEDIUM | Per-run (w:rPr) formatting. Must handle direct formatting vs style inheritance. |
| Paragraph formatting (alignment, spacing, indentation) | Block-level layout control | MEDIUM | w:pPr properties. Complex due to interplay with styles. |
| Apply named styles (Heading1, Normal, custom) | Users expect document structure via styles, not just direct formatting | MEDIUM | Style resolution: style → style (basedOn) → docDefaults. Must handle style hierarchy. |
| Save to .docx | Write valid OPC package | MEDIUM | ZIP archive with correct content types + rels for each part. Must be ISO 29500 compliant. |
| Tables with rows and cells | Data presentation is a core Word use case | HIGH | Complex OOXML: tblGrid, tblPr, trPr, tcPr. Cell merging (hMerge/vMerge), borders, widths. |
| Images (PNG/JPG) | Documents need graphics | MEDIUM | Must handle image parts, relationships, drawingML, extent/sizing. |
| Lists (ordered/bulleted) | Numbered and bulleted lists are basic document elements | MEDIUM | Requires numbering definitions (abstractNum/num), list formatting properties. |
| Page setup (margins, orientation, size) | Document-level layout | MEDIUM | Section properties (sectPr). Page dimensions, margins, orientation, headers/footers reference. |
| Headers and Footers | Standard document structure | HIGH | Separate document parts. Must handle first-page, odd/even variants. Link to previous. |
| Page breaks | Basic document flow control | LOW | Simple w:br element with type="page". |
| Hyperlinks | Text and image hyperlinks | MEDIUM | Requires relationship IDs, external or internal targets, optional tooltip. |

### Differentiators (Competitive Advantage)

Features that set this library apart from godocx, unioffice, and nguyenthenguyen/docx.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Style preservation when editing existing docs** | Open a .docx, modify text, save — styles remain intact. No existing Go library does this correctly. | **CRITICAL/HIGH** | The core differentiator. Must preserve w:rPr, w:pPr, w:tblPr from input. Don't strip unknown XML. Use XML round-tripping. |
| **Template-based creation (reference-style inheritance)** | Create documents using styles from a reference .docx. Users bring branded templates; output matches their brand. | HIGH | Load styles.xml from template, apply to new document body. Must handle style hierarchy, numbering, fonts table. |
| **Content preservation (unknown XML passthrough)** | Edit a doc without losing content controls, custom XML, or legacy features the library doesn't explicitly know about. | HIGH | Preserve all w:r/w:p children the library doesn't understand. Write-via-XML not write-via-deserialize. |
| **Go struct → document mapping** | Map Go structs to document content (tables, paragraphs) without manual element-by-element construction. | MEDIUM | Reflection-based or interface-based mapping. JSON/YAML struct tags define document structure. |
| **Insert/edit/delete content in existing docs** | Not just create — modify existing documents. Replace text, add paragraphs, delete sections. | HIGH | Must navigate document body, target specific paragraphs/runs, modify in place without corrupting. |
| **{{placeholder}} template merge** | Users design template in Word, library fills variables. Drives most real-world doc generation use cases. | MEDIUM | Replace `{{key}}` in paragraphs, table cells, headers/footers. Content controls alternative. |
| **Tables with full styling** | Table style application, cell formatting (shading, borders, width), row/column operations. | HIGH | Beyond basic tables. Style from gallery, cell shading, borders per-cell, vertical alignment, merged cells. |
| **Document sections** | Multiple sections with independent page setup, headers/footers, columns. | HIGH | Section breaks, per-section page layout, per-section headers/footers. Complex OOXML. |
| **Comments** | Review annotations. Round-trip existing comments, add new ones. | MEDIUM | Comment parts, comment references in body. Author, date, text. |
| **Bookmarks** | Named locations for navigation and cross-references. | MEDIUM | Bookmark start/end in body. Target by name. |
| **Field codes (TOC, PAGE, DATE)** | Auto-updating fields. Table of contents, page numbers, date fields. | HIGH | Field instructions (instrText). Word updates these on open. Library writes correct field codes. |
| **RTL and i18n support** | Arabic, Hebrew, CJK, Hindi documents. Multi-language documents. | MEDIUM | Complex script properties, bidirectional text, per-script fonts, locale tracking. |
| **Pure Go, no CGO** | Static binary, cross-compilation, easy CI integration. | LOW | Use archive/zip, encoding/xml. No libxml2, no COM, no .NET dependency. |
| **Document validation** | Check generated documents before delivery. Warn about missing required parts, broken relationships. | MEDIUM | Validate OPC package structure, required parts present, relationship targets exist. |
| **GTK-free headless operation** | Works in Docker, CI, Lambda — no display, no Office install. | LOW | By design: Go + standard libs only. |

### Anti-Features (Commonly Requested, Often Problematic)

Features to explicitly avoid or defer.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| **PDF conversion** | Users often want .docx → PDF output | Requires layout engine (page rendering). Massive scope. Separate project. | Document-specific PDF tool (e.g., unoconv, pandoc, LibreOffice headless) |
| **Real-time rendering / live preview** | "See the document as I build it" | Requires HTML/CSS renderer, bidirectional sync, UI. Entirely different product. | Use an external renderer for preview; generate .docx with pure Go for delivery |
| **Word COM automation** | "Just use Microsoft Word to do it" | Requires Windows + Word install. CGO. Fragile. Defeats headless purpose. | Pure OPC/OOXML manipulation. This is the entire point of the library. |
| **.doc format support** | Legacy .doc compatibility | Binary format, completely different spec (OLE2). Massive scope. | Stick to .docx (ISO 29500). Provide migration docs. |
| **.odt / .rtf support** | Multi-format output | Different specs. Dilutes focus. | Keep .docx only. Users convert via pandoc if needed. |
| **Collaborative editing** | Multiple users editing same doc | Requires operational transforms, conflict resolution, real-time sync. | Single-user document processing. Users manage versioning in git. |
| **Full formula/programming API** | "Let me run VBA-like code" | VBA is a complete runtime. Insane scope. | Users write Go code to use the library. That IS the scripting. |
| **Built-in web server / watch mode** | "Live preview while editing" | Different paradigm. CLI tool vs library concern. | Library exports document as immutable snapshot. External tools watch file system. |

## Mature-Tool Feature Gap Analysis

Feature mapping between what mature OOXML tooling (built on Microsoft's DocumentFormat.OpenXml) supports and what this Go library should support.

| Feature Area | Mature C# tooling | Go Library (target) | Gap Notes |
|-------------|--------------|---------------------|-----------|
| Create new .docx | ✅ | ✅ | Must match |
| Read text/structure | ✅ | ✅ | Must match |
| Edit existing docs | ✅ | ✅ | Must match |
| Style preservation | ✅ (built-in) | ✅ (core differentiator) | Go must explicitly design for this |
| Template merge ({{key}}) | ✅ | ✅ | Must match |
| Paragraph formatting | ✅ (framePr, tabs, indents) | ✅ (subset: alignment, spacing, indents) | Skip framePr v1 |
| Run formatting | ✅ (bold, italic, underline, font, size, color, position) | ✅ (bold, italic, underline, font, size, color) | Skip position half-pts v1 |
| Tables | ✅ (virtual cols, hMerge, styles) | ✅ (basic + style, merge) | Skip virtual cols v1. Need hMerge. |
| Images | ✅ (PNG/JPG/GIF/SVG) | ✅ (PNG/JPG) | Skip SVG/GIF v1 |
| Headers/Footers | ✅ (first, odd/even, link-to-prev) | ✅ (basic + variants) | Must match for template use |
| Sections | ✅ | ✅ | Must match |
| Styles (add/modify) | ✅ (full management) | ✅ (apply existing, create simple) | Skip modifying existing styles v1 |
| Lists | ✅ | ✅ | Must match |
| Hyperlinks | ✅ | ✅ | Must match |
| Comments | ✅ | In v2 | Defer to post-MVP |
| Footnotes | ✅ | In v2 | Defer |
| Watermarks | ✅ | In v2 | Defer |
| Bookmarks | ✅ | In v2 | Defer |
| TOC | ✅ | In v2 | Defer. Needs field code support. |
| Charts | ✅ | In v2 | Defer. Complex OOXML. |
| Content Controls (SDT) | ✅ | In v2 | Defer. Preserve on read though. |
| Form fields | ✅ | In v2 | Defer |
| Equations | ✅ (LaTeX) | In v2 | Defer |
| Diagrams | ✅ (mermaid) | No | Out of scope |
| Tracked changes / revisions | ✅ | In v2 | Defer |
| OLE objects | ✅ | No | Out of scope |
| Shape/Textbox | ✅ | In v2 | Defer |
| Dump to JSON / batch | ✅ | In v2 | Defer |
| Watch/live preview | ✅ | No | Out of scope (anti-feature) |
| Document validation | ✅ | In v2 | Defer |
| Refresh (TOC/page numbers) | ✅ | In v2 | Defer, needs external Word |
| MCP server / AI integration | ✅ | In v3 | Defer |

## Feature Dependencies

```
[OPC Package: read/write .docx as ZIP]
    └──requires──> [Content Type handling]
    └──requires──> [Relationship handling]

[Style preservation]
    └──requires──> [OPC Package]
    └──requires──> [XML round-trip: preserve unknown elements]
    └──requires──> [Style resolution: basedOn, defaults, hierarchy]

[Document creation]
    └──requires──> [OPC Package]
    └──requires──> [Minimal required parts: document.xml, styles.xml, rels]

[Template-based creation]
    └──requires──> [OPC Package]
    └──requires──> [Style resolution]
    └──requires──> [Clone styles.xml + numbering from template]

[Paragraph + Run formatting]
    └──requires──> [Document creation (for new docs)]
    └──requires──> [OPC Package (for edit)]

[Tables]
    └──requires──> [Paragraph + Run formatting (for cell content)]
    └──requires──> [Style resolution (table styles)]

[Headers/Footers]
    └──requires──> [OPC Package]
    └──requires──> [Section support]
    └──enhances──> [Template-based creation]

[Images]
    └──requires──> [OPC Package (add part + relationship)]
    └──requires──> [DrawingML: inline anchors]

[Template merge {{placeholder}}]
    └──requires──> [Paragraph + Run formatting (find and replace in runs)]
    └──requires──> [Table support (replace in cells)]
    └──enhances──> [Template-based creation]

[Go struct → document mapping]
    └──requires──> [Document creation]
    └──requires──> [Tables]
    └──enhances──> [Template merge]

[Field codes (TOC, PAGE)]
    └──requires──> [Paragraph support (field char runs)]
    └──requires──> [Field instruction syntax: fldCode]
    └──conflicts──> None, but Word opens and updates; library must write correct XML

[Comments]
    └──requires──> [OPC Package]
    └──requires──> [Paragraph support (commentReference anchor)]

[Bookmarks]
    └──requires──> [Paragraph support (bookmarkStart/End)]
```

### Dependency Notes

- **Style preservation is the foundation** of the entire differentiator strategy. Without it, this library is just another godocx clone. Every read operation must round-trip the XML, preserving properties the library doesn't understand.
- **OPC Package layer** is the most critical foundation. All features depend on correct ZIP structure, content types, and relationships. Get this wrong and no generated document opens in Word.
- **Template-based creation** depends on style resolution and template reading. Must be designed from the start, not bolted on.
- **Tables and images** have no hard ordering dependency on each other but both need the OPC package layer.
- **V2 features** (comments, bookmarks, footnotes) share the same OPC + paragraph dependency. They add new XML parts but don't change the core architecture.
- **Field codes** are tricky: they span multiple runs (fldChar begin/separate/end). The library must write valid field instruction XML that Word can interpret and update on open.

## MVP Definition

### Launch With (v1.0)

Minimum viable product — what validates the concept. Core differentiator: style-preserving read + template-merge write.

- [x] **OPC package layer** — Read/write .docx as ZIP with content types + relationships. Foundation of everything.
- [ ] **Read .docx with style preservation** — Open existing doc, extract paragraph/run content, preserve all XML properties the library doesn't touch. Critical: write-back must not strip unknown elements.
- [ ] **Create new .docx from scratch** — Minimal valid document with heading, normal text, and basic formatting.
- [ ] **Paragraph and run formatting** — Bold, italic, underline, font, size, color, alignment, spacing.
- [ ] **Named styles** — Apply heading styles (H1-H3), Normal, and custom styles. Style resolution from template.
- [ ] **Template-based creation** — Open reference .docx, clone styles/numbering, create new body with template's styling.
- [ ] **Tables** — Create tables with rows/cells, basic borders, cell text formatting, cell shading.
- [ ] **Images** — Embed PNG/JPG images with drawingML anchors.
- [ ] **Headers and Footers** — Simple header/footer per section. Default page number.
- [ ] **Lists** — Ordered and unordered lists with numbering definitions.
- [ ] **Hyperlinks** — External hyperlinks on text runs.
- [ ] **Page setup** — Margins, orientation, paper size via section properties.
- [ ] **{{placeholder}} template merge** — Replace `{{key}}` markers in paragraph runs, table cells. Works on template-loaded docs.
- [ ] **Go struct → table mapping** — Map slices/structs to table rows with column mapping tags.

### Add After Validation (v1.x)

Features to add once core is working and users validate the approach.

- [ ] **Insert/delete paragraphs** — Edit existing docs: add paragraph before/after target, delete paragraph.
- [ ] **Insert/delete runs in paragraphs** — Split existing runs, insert new formatted text.
- [ ] **Document sections** — Multiple sections with independent page setup. Section breaks.
- [ ] **Inline formatting in template merge** — `{{key}}` inside mixed-format runs preserves surrounding formatting.
- [ ] **Table row insert/delete** — Add rows to existing tables, delete rows.
- [ ] **Image sizing/positioning** — Absolute positioning, text wrapping styles.
- [ ] **Field codes** — PAGE, DATE, NUMPAGES fields. Table of Contents field (TOC \o). Preserve existing fields on read.
- [ ] **Document properties** — Author, title, subject, keywords metadata.
- [ ] **Page breaks** — Explicit page breaks.
- [ ] **Better style resolution** — BasedOn chain resolution, style defaults inheritance, latent styles.

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] **Comments** — Read/write threaded comments. Preserve on round-trip.
- [ ] **Footnotes and endnotes** — Read/write, preserve on edit.
- [ ] **Bookmarks** — Read/write, target by name.
- [ ] **Content Controls (SDT)** — Read/preserve SDT boundaries. Write structured document tags.
- [ ] **Form fields** — Text, checkbox, dropdown form fields. Fill and save.
- [ ] **Tracked changes / revisions** — Read change tracking markup. Accept/reject logic.
- [ ] **Watermarks** — Text and image watermarks.
- [ ] **Shapes and Textboxes** — DrawingML shapes, textboxes with formatting.
- [ ] **Charts** — Simple chart types (bar, line, pie) with data.
- [ ] **Equations** — OMML equation support.
- [ ] **Dump to JSON / batch replay** — Serialize document structure to workflow-friendly JSON.
- [ ] **Document validation** — Validate OPC structure, required parts, relationship integrity.
- [ ] **RTL / i18n** — Full complex-script support, per-script fonts, Bidi properties.
- [ ] **No-install OPC validation** — End-to-end test suite that opens generated docs in a checker.

### Out of Scope (Forever)

- **PDF conversion** — Use separate tool (pandoc, LibreOffice, wkhtmltopdf).
- **Real-time rendering** — Not a library concern. Use an external renderer.
- **.doc format** — Use LibreOffice for conversion.
- **.odt / .rtf** — Not this library's format.
- **COM interop / Word automation** — Defeats the purpose.
- **Collaborative editing** — Single-user processing.
- **Mermaid diagrams** — User renders externally and embeds as image.
- **OLE objects** — Embedded spreadsheets, etc. Rarely needed.
- **Watch / live preview server** — Different product.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| OPC package layer | HIGH (foundation) | HIGH | P1 |
| Read with style preservation | CRITICAL | HIGH | P1 |
| Create new .docx | HIGH | MEDIUM | P1 |
| Paragraph + run formatting | HIGH | MEDIUM | P1 |
| Named styles | HIGH | MEDIUM | P1 |
| Template-based creation | CRITICAL | HIGH | P1 |
| Tables (basic) | HIGH | HIGH | P1 |
| Images (basic) | MEDIUM | MEDIUM | P1 |
| Headers/Footers | HIGH | HIGH | P1 |
| Lists | HIGH | MEDIUM | P1 |
| Hyperlinks | MEDIUM | MEDIUM | P1 |
| Page setup | MEDIUM | MEDIUM | P1 |
| {{placeholder}} merge | CRITICAL | MEDIUM | P1 |
| Go struct → table mapping | HIGH | MEDIUM | P1 |
| Insert/delete paragraphs | HIGH | MEDIUM | P2 |
| Insert/delete runs | MEDIUM | MEDIUM | P2 |
| Multiple sections | MEDIUM | HIGH | P2 |
| Table row insert/delete | HIGH | MEDIUM | P2 |
| Field codes | MEDIUM | HIGH | P2 |
| Document properties | LOW | LOW | P2 |
| Comments | MEDIUM | MEDIUM | P3 |
| Footnotes | LOW | MEDIUM | P3 |
| Bookmarks | MEDIUM | LOW | P3 |
| Content Controls | LOW | HIGH | P3 |
| Tracked changes | LOW | HIGH | P3 |
| Watermarks | LOW | MEDIUM | P3 |
| Charts | LOW | HIGH | P3 |
| Equations | LOW | HIGH | P3 |
| Document validation | MEDIUM | MEDIUM | P3 |

**Priority key:**
- P1: Must have for v1 launch
- P2: Should have, add when core validated
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | unioffice (commercial) | godocx (MIT) | nguyenthenguyen/docx | This Library (target) |
|---------|----------------------|-------------|---------------------|----------------------|
| License | Commercial (paid) | MIT | MIT | MIT/Apache2 |
| Pure Go | Yes | Yes | Yes | Yes |
| Read .docx | ✅ | Limited | ✅ (.xml only) | ✅ |
| Create .docx | ✅ | ✅ | ❌ | ✅ |
| Edit existing | ✅ | ❌ | ✅ (replace only) | ✅ |
| Style preservation | Partial | ❌ | ❌ | ✅ (core) |
| Template-based creation | ✅ (use-template) | ❌ | ❌ | ✅ (core) |
| Headers/Footers | ✅ | ❌ | ✅ (replace) | ✅ |
| Tables | ✅ | ✅ | ❌ | ✅ |
| Images | ✅ | ❌ | ✅ (replace) | ✅ |
| Lists | ✅ | ❌ | ❌ | ✅ |
| Field codes | ✅ | ❌ | ❌ | v2 |
| Comments | ❌ | ❌ | ❌ | v2 |
| Performance | HIGH (30K rows/3.9s) | MEDIUM | LOW | TBD |
| Binary size | 33MB (all formats) | Small | Small | Small (docx only) |
| Price | License key required | Free | Free | Free |

## Sources

- **Mature C# OOXML tooling:** DocumentFormat.OpenXml-based feature surfaces (paragraph, run, table, style, section, header/footer, comment, field, etc.) as documented in the .NET Open XML SDK and ECMA-376
- **unioffice:** GitHub repository (https://github.com/unidoc/unioffice) — Readme features list, examples at unioffice-examples
- **godocx (gomutex):** GitHub repository (https://github.com/gomutex/godocx) — Readme, API docs at pkg.go.dev
- **nguyenthenguyen/docx:** GitHub repository (https://github.com/nguyenthenguyen/docx) — Readme, simple text replacement library
- **OOXML / ISO 29500:** ECMA-376 Office Open XML standard, Part 1 (Fundamentals), Part 2 (OPC), Part 3 (Markup Compatibility), Part 4 (Transitional) — referenced for OPC structure understanding

---

*Feature research for: wordingo — Word Document Library*
*Researched: 2026-07-25*
