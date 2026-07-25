# Pitfalls Research

**Domain:** Pure Go .docx creation/editing library (OPC / ISO 29500)
**Researched:** 2026-07-25
**Overall confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: Treating .docx as a Flat XML File Instead of an OPC Package

**What goes wrong:**
New implementers treat .docx as a single XML file. They unzip, parse `word/document.xml`, modify text, re-zip — and produce a file Word rejects with "corrupt" or silently loses content. They miss relationship parts (`_rels/.rels`, `word/_rels/document.xml.rels`), the `[Content_Types].xml` manifest, and parts like `word/styles.xml`, `word/settings.xml`, `word/numbering.xml`, `word/fontTable.xml`, `word/webSettings.xml`, headers, footers, footnotes, endnotes, comments, and glossary.

**Why it happens:**
The OPC (Open Packaging Conventions) model is unintuitive — a .docx is a ZIP archive of *parts* connected by *relationships*. The main document XML is just one part. Most Go developers approach it as "XML in a ZIP" and skip the packaging layer. The spec (ECMA-376 Part 2) is 120+ pages; few read it.

**How to avoid:**
Implement strict OPC package abstraction from day 1:
- `Package` struct wrapping `zip.Reader`/`zip.Writer` with part enumeration
- `Relationship` graph traversal (source → target, typed by URI)
- `ContentTypes` tracking for every part
- Validation: every part must have a content type; every relationship must resolve
- Write tests against real .docx files opened/saved by Word

**Warning signs:**
- Output opens in Word with "repair" dialog
- Images/headers/footers missing in output
- Word opens file but styles are wrong
- `zip: not a valid zip file` panic (sign of broken ZIP structure from issue #27 in nguyenthenguyen/docx)

**Phase to address:**
Phase 1 (OPC Core) — must be right before any content manipulation.

---

### Pitfall 2: Copying C# `System.IO.Packaging` Patterns Directly to Go

**What goes wrong:**
C# Open XML SDK uses `Package.Open()`, `PackagePart`, and `PackageRelationship` — an object model where parts expose streams. Porting this 1:1 to Go produces overengineered interfaces that fight Go idioms. The C# SDK also has `using` blocks for stream lifecycle, implicit relationship resolution via `Part.GetRelationship()`, and LINQ to XML for querying. Go has no LINQ, no `using`, and `encoding/xml` is struct-tag based, not query-based.

Consequences: either massive code that feels like C# translated line-by-line, or a Go API that frustrates users because it breaks the Go convention of simple structs + methods.

**Why it happens:**
Reference implementations are C#. Natural inclination is to port class-for-class. But C# Open XML SDK was designed for .NET idioms (LINQ, IDisposable, generics) that don't exist in Go.

**How to avoid:**
Design the Go API as Go-first, not a transliteration:
- Use `io.ReaderAt`/`io.Writer` for ZIP I/O (standard Go patterns)
- Expose XML as decoded structs, not streaming parts
- No inheritance — use composition with interfaces
- No `PackagePart` class — use a `Part` struct with `Name`, `ContentType`, and `Data []byte`
- Relationship graph as a simple `map[string][]Relationship` on the Package

**Warning signs:**
- Code has classes named `PackagePart`, `PackageRelationship`, `OpenSettings`
- Heavy use of interfaces where Go would use concrete structs
- Panic-prone nil checks because C# null patterns don't map to Go zero values
- `Stream`/`Seeker` patterns that make no sense for in-memory XML

**Phase to address:**
Phase 0 (Architecture & API Design) — establish Go API contract before any implementation.

---

### Pitfall 3: Ignoring XML Namespace Prefix Registration and Atomization

**What goes wrong:**
The WordprocessingML namespace (`http://schemas.openxmlformats.org/wordprocessingml/2006/main`) appears in dozens of element names. If you use `encoding/xml` with struct tags like `xml:"w:p"`, Go's decoder requires the prefix `w:` to match exactly what's in the file. But Word and other producers use varying prefixes (`w:`, `wne:`, or even no prefix on some elements). Files with different prefix mappings silently fail to parse — no error, just empty structs.

Additionally, many parts use *multiple* namespaces in the same document (w:, r:, wp:, wp14, mc:, a:, etc.). Missing or swapping namespace aliases causes partial document loss.

**Why it happens:**
`encoding/xml` is namespace-aware but prefix-sensitive by default. The Open XML spec defines 40+ namespaces. Most Go XML libraries assume a single namespace per document.

**How to avoid:**
- Build a namespace registry that maps URIs to prefixes, not the reverse
- When decoding, normalize to canonical prefixes before matching struct tags
- Use `xml:"{http://...}localName"` syntax (Go 1.2+) instead of prefix-based tags to avoid prefix dependency
- Register all known namespaces from ECMA-376 Parts 1 and 4
- For unknown namespaces, preserve raw XML rather than dropping

**Warning signs:**
- `encoding/xml` silently skips elements (unmarshal into empty struct with no error)
- Style elements disappear after round-trip
- Word complains "content is unreadable" on files using alternate prefixes
- Drawings/images vanish (the `a:` namespace elements were dropped)

**Phase to address:**
Phase 1 (OPC Core) — namespace handling built into the XML layer from the start.

---

### Pitfall 4: Style Resolution — Treating Styles as a Flat List

**What goes wrong:**
When applying styles from a template, implementers copy `styles.xml` verbatim and expect formatting to work. But style resolution in WordprocessingML uses a complex inheritance chain:
1. `latentStyles` (default style definitions)
2. `docDefaults` (document-level run/property defaults)
3. `style` definitions with `basedOn` chains (e.g., "Heading 1" inherits from "Heading" which inherits from "Normal")
4. `style` with `next` (paragraph following inherits next style)
5. Paragraph properties (`pPr`) override style
6. Run properties (`rPr`) override paragraph properties
7. Direct formatting (inline `rPr`) overrides everything

If you import a style by copying `styles.xml` but don't include the `basedOn` chain, all styles collapse to defaults. If you copy `styles.xml` but not `numbering.xml`, numbered lists break. If you copy from a different template, style IDs that don't exist in the document cause Word to apply Normal style silently.

**Why it happens:**
Styles look like a simple key-value map (`styleId` → properties) in the XML. The inheritance is implicit — `basedOn` and `next` attributes create a DAG that's easy to miss. Also, styles reference numbering definitions via `numPr` → `numId` that must exist in `numbering.xml`.

**How to avoid:**
- Implement explicit style resolution: resolve `basedOn` chains in order (deepest ancestor first)
- Track `styleId` → definition mapping AND validate all references exist
- When copying styles from a template, also copy dependent parts (`numbering.xml`, `fontTable.xml`, `styles.xml` with full chain)
- After style copy, run validation: every `basedOn`, `next`, `numPr`, `rPr` reference must resolve
- Test with: Normal, Heading 1-9, List styles, Table styles, Character styles

**Warning signs:**
- Output text has wrong font/size even though template style is "applied"
- Numbered lists all show "1." (numId reference broken)
- Heading styles don't cascade (H2 inheriting from H1)
- Style name appears in Word but formatting is blank

**Phase to address:**
Phase 3 (Style Preservation) — requires OPC core (Phase 1) and basic document editing (Phase 2) to be stable first.

---

### Pitfall 5: Round-Trip Infidelity — Silent Data Loss on Unknown Elements

**What goes wrong:**
A library reads a .docx, the user edits one paragraph, saves. The output opens, but custom XML parts (`word/customXml/*`), `altChunk` references, `mailMerge` settings, `webSettings`, `glossaryDocument`, `activeX` controls, embedded fonts, or macro-enabled VBA projects are gone. The document was "edited" but lost significant data.

This is by far the most common complaint against Go .docx libraries (evident in unioffice issue #370 where "unsupported relationship type" messages get logged and parts dropped).

**Why it happens:**
Libraries parse only the parts they know about. When saving, they write only the parts in their internal model. Parts not in the model are omitted. The library "preserves the document" but silently drops everything not in its schema registry.

**How to avoid:**
- Implement "pass-through" preservation: any part not understood is read as raw bytes and written back identically
- Maintain a `map[string]Part` of all original parts; on save, merge modified parts with untouched parts
- Log warnings for unknown relationship types but NEVER drop them
- Track `[Content_Types].xml` entries — if a part exists but isn't in content types, Word may reject the file
- After save, verify byte-for-byte that untouched parts match originals

**Warning signs:**
- Document loses ActiveX controls after edit
- Mail merge fields stop working
- Custom XML data bindings break
- Embedded fonts missing
- VBA macros vanish (.docm → .docx corruption)
- `glossaryDocument` entries (Quick Parts, Building Blocks) disappear

**Phase to address:**
Phase 1 (OPC Core) — preservation mechanism is a package-level concern, not content-level.

---

### Pitfall 6: Transitional vs. Strict Conformance — Wrong Schema Target

**What goes wrong:**
A library writes files conforming to ISO 29500 Strict (the 2012+ standard) but users open them in Office 2007. Office 2007 only supports Transitional conformance (ECMA-376 1st edition). The file opens with "conversion" dialog or fails entirely.

Conversely, writing Transitional markup that uses Office 2007-era features (VML instead of DrawingML for shapes, legacy `wp:` positioning) produces bloated XML and may not render correctly in Office 2021+.

unioffice issue #401 documents this exact problem: "Office2007 is not supported" because the library targets a newer conformance class.

**Why it happens:**
The spec is confusing: ECMA-376 1st ed. (2006) = Transitional, ISO 29500:2008 + ECMA-376 5th ed. = Strict + Transitional with deprecations. Libraries pick one conformance class without documenting tradeoffs.

**How to avoid:**
- Default to **Transitional** (ISO 29500 Transitional) for maximum compatibility — Office 2007/2010/2013/2016/2019/2021 all support it
- Support reading both; validate conformance class on open via `xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"` (Transitional) vs `http://purl.oclc.org/ooxml/wordprocessingml/main` (Strict)
- When writing, use the same conformance class as the source document (or Target Transitional for new docs)
- Document the conformance class decision prominently

**Warning signs:**
- "Sorry, Word can't read this document" on Office 2007/2010
- "This document contains content that Word doesn't support" warnings
- Files pass validation in Oxygen but fail in Office
- `xmlns:w` value changes between read and write

**Phase to address:**
Phase 1 (OPC Core) — conformance class documented and enforced at package creation time.

---

### Pitfall 7: Breaking Template Style References by Copying Only `styles.xml`

**What goes wrong:**
Core project requirement: "preserve styles from a reference template document." Implementers copy `word/styles.xml` from the template to the output document. Result: styles appear to be there, but formatting is off. Fonts don't match. Colors differ. This happens because styles reference other parts:
- `styles.xml` references fonts in `word/fontTable.xml`
- `styles.xml` references numbering in `word/numbering.xml`
- `styles.xml` references themes in `word/theme/theme1.xml`
- `styles.xml` references colors in `word/theme/theme1.xml` (via `clrSchemeMapping`)
- `styles.xml` references `latentStyles` → default paragraph/run properties

**Why it happens:**
Developers see "style" as a single XML file. In reality, style resolution is distributed across the OPC package.

**How to avoid:**
Copy the entire dependency graph when adopting template styles:
1. Copy `word/styles.xml`
2. Copy `word/fontTable.xml`
3. Copy `word/numbering.xml`
4. Copy `word/theme/theme1.xml` (and any theme parts)
5. Copy `word/settings.xml` (contains `defaultTabStop`, `zoom`, etc.)
6. Copy relationship entries for all copied parts
7. Update content types for any new parts
8. Validate: open in Word, verify template style renders identically

**Warning signs:**
- Template uses custom font but output shows fallback font (fontTable not copied)
- Template uses theme colors but output uses default colors
- Template numbered lists show as plain text (numbering not copied)
- Template heading colors differ from source

**Phase to address:**
Phase 3 (Style Preservation).

---

### Pitfall 8: XML Whitespace Handling — Run Text Collapse

**What goes wrong:**
Word stores text in `<w:t>` elements. A paragraph "Hello World" is stored as `<w:r><w:t>Hello World</w:t></w:r>`. But text with leading/trailing spaces must have `xml:space="preserve"` on the `<w:t>` element. When Go's `encoding/xml` marshals a struct with `xml:",innerxml"`, it trims whitespace. The result: spaces vanish, text runs collapse, words concatenate.

Worse: Word breaks paragraphs into multiple runs (`<w:r>`) for tracking formatting changes, spelling suggestions, or just because. If the library simplifies runs by merging them, it breaks revision tracking. If it preserves runs but modifies text inside them without preserving `xml:space`, spacing breaks.

**Why it happens:**
`encoding/xml` treats whitespace in element content as insignificant by default (per XML spec). But WordprocessingML requires `xml:space="preserve"` for `<w:t>` and `<w:r>` to retain spaces. Also, the run structure is meaningful — runs delineate formatting spans.

**How to avoid:**
- Always set `xml:space="preserve"` on `<w:t>` when the text has leading/trailing spaces or multiple consecutive spaces
- On unmarshal, capture `xml:space` attribute and re-apply on marshal
- Preserve run boundaries — never merge runs unless explicitly asked (and only if no revision tracking is present)
- Use `xml:",innerxml"` with caution; prefer struct fields with `xml:",chardata"` for text content
- Normalize: read all text content into a flat buffer, but write back preserving original run structure

**Warning signs:**
- "Hello  World" (two spaces) becomes "Hello World" (one space)
- " Hello " becomes "Hello"
- Text formatting boundaries shift (bold span grows/shrinks)
- Revision marks appear on wrong text

**Phase to address:**
Phase 2 (Document Model) — text handling in the document data model.

---

### Pitfall 9: Relationship Reference Breakage on Part Rename

**What goes wrong:**
When editing a .docx, the library renames internal parts (e.g., rebasing `word/media/image1.png` after insertion). But `document.xml.rels` still references `media/image1.png` by its relationship ID (`rId5`). The renamed file path no longer matches. Images go missing. Alternatively, relationship IDs collide (`rId1` used twice) because the library assigns sequential numbers without checking existing IDs.

**Why it happens:**
Relationships are a graph, not a list. Each part has its own `.rels` file. Changing a part's path or content type requires updating:
1. The source part's `.rels` file (the `Target` attribute)
2. `[Content_Types].xml` (if content type changed)
3. Any other part that references it

Go libraries often track relationships in memory but forget to serialize updates to the `.rels` XML files.

**How to avoid:**
- Auto-generate relationship IDs with prefix + counter, checking against existing IDs in the `.rels` file
- On any part add/remove/rename, update exactly three things: the `.rels`, the content type manifest, and the part file
- Maintain a bidirectional `PartURI ↔ RelationshipID` map per part
- After any modification, run a consistency check: every `Target` in every `.rels` must resolve to an existing part; every part must have a content type
- Never reuse an `rId` even if the original part was deleted

**Warning signs:**
- Broken images after document edit
- "The relationship of the specified part is not valid" error
- Content type errors when opening
- `rId` conflicts causing only some images to appear

**Phase to address:**
Phase 1 (OPC Core) — relationship graph integrity enforced at the package layer.

---

### Pitfall 10: Porting C# LINQ-to-XML Queries to Go Without a Query Layer

**What goes wrong:**
C# OOXML tools use LINQ to XML heavily — `doc.Descendants(W.p).Where(e => e.Attribute(W.styleId)?.Value == "MyStyle")`. Go's `encoding/xml` unmarshals into structs. Direct translation produces code that walks the entire document tree manually with recursion + type assertions. This is verbose, error-prone, and slow.

C# to Go porters end up writing XML walking helpers (`FindElement`, `Descendants`, `Ancestors`) that add a mini-XML-query library on top of `encoding/xml`. This code grows unmaintainable and rarely covers edge cases (XML namespaces, comments, processing instructions).

**Why it happens:**
Go deliberately has no LINQ equivalent. `encoding/xml` is a marshal/unmarshal library, not a query library. Developers underestimate how much C# code relies on LINQ-to-XML for navigation — it's not just queries, it's the primary way the C# SDK reads documents.

**How to avoid:**
Two valid approaches:
1. **Query layer built on `encoding/xml`**: Write safe `Descendants()` and `Element()` helpers using `xml.NewDecoder` in token mode, namespace-aware, with limited memory footprint. Accept that this is a distinct subproject.
2. **Unmarshal to Go structs**: Define Go structs mirroring the schema (like unioffice does with auto-generated types). Use tagged structs for direct unmarshal. More code generation up front, less query code at point of use.

Recommendation: Approach 2 for a library (predictable, type-safe), Approach 1 only for ad-hoc document inspection.

Either way, decide in Phase 0 and commit. Switching halfway causes a rewrite.

**Warning signs:**
- `FindDescendant()` helper functions proliferating across the codebase
- Same XML traversal logic duplicated in every feature module
- XML parsing scattered instead of centralized in document model layer
- Type assertions (`.(type)`) on XML tokens everywhere

**Phase to address:**
Phase 0 (Architecture) — XML query strategy is a foundational decision.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Store modified parts as raw `[]byte` without schema validation | Fast prototyping | Corrupt documents on subtle XML errors; missed namespace issues; debugging hell | Never in library code. Acceptable in CLI exploration scripts. |
| Skip `.rels` update on part add (rely on Word to fix) | Saves 20 lines of code | Files that Word opens with repair; relationship errors; support tickets | Never. Word is NOT a validator — it silently drops unresolvable relationships. |
| Merge all `<w:r>` elements on read (simplify internal model) | Simple text model | Breaks revision tracking, field codes, hyperlinks, comments, annotations, and output differs from source | Only if the library explicitly disclaims round-trip fidelity for tracked changes |
| Hardcode `Content_Types.xml` instead of deriving from content | Simpler save path | Missing content types for images, embedded objects, OLE, ActiveX; Word rejection | Only for a library that creates new documents from scratch (no editing existing) |
| Ignore unknown namespaces | Faster development | Silent data loss on documents with custom XML parts, legacy VML, or future extensions | Never. Use "pass-through" preservation. |
| Single global namespace map | Simple code | Documents with conflicting prefix aliases break; conformance class detection breaks | Only if you control both producer and consumer |
| Skip `altChunk` handling | ~500 lines less code | Documents with embedded HTML/Word fragments lose that content silently | Acceptable only in Phase 1 MVP; must be added before any "production" release |

---

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Unmarshal entire document to struct tree on every read | High memory, slow open for 100+ page docs | Lazy loading: only parse parts on access; cache parsed result | Documents >50 pages or with many images |
| Serialize entire package on every save (re-zip all parts) | Slow save for large files | Write individual parts to `zip.Writer` sequentially; track which parts changed and only write those | Document packages >50MB |
| In-memory DOM for all XML processing | OOM on large tables (10k+ rows) | Use streaming XML token reader for bulk content; DOM only for styles/layout | Spreadsheet-like documents in .docx (tables with 1000s of rows) |
| Re-compress images on every save | Quality loss, CPU waste | Store image blobs as-is; only compress new images | All image-containing documents |
| No part-level cache | Re-parsing the same `.rels` file 20 times per operation | Cache parsed `.rels` per part; invalidate only when relationships change | Documents with complex structure (many headers, footers, embedded content) |
| String concatenation for XML construction | Repeated allocations, GC pressure | Use `strings.Builder` or write directly to `xml.Encoder` | Generating documents with 1000+ paragraphs |

---

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| No ZIP bomb protection | OOM from decompression bomb | Limit total decompressed size (e.g., 512MB); check compression ratio per entry |
| XML entity expansion | Billion laughs / quadratic blowup attack | Limit XML entity depth; disable DTD processing (`xml.NewDecoder` → `Strict: true`) |
| Unvalidated relationship `Target` paths | Path traversal reading/writing outside package | Reject `Target` values with `..` or absolute paths; validate part names against OPC part name rules |
| Accepting any external relationship | SSRF via external document references | Validate relationship `TargetMode`; warn on `External` relationships; don't resolve automatically |
| Unrestricted image part content types | MIME type confusion, file upload mimicry | Validate content type matches magic bytes; reject executable types |
| No max parts limit | ZIP with 10k empty parts exhausts FD/file handles | Limit to 1000 parts (generous for ~400 page doc); fail fast |

---

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Silent corruption on unknown elements | Document breaks after edit, no error | Log warnings per dropped part; expose `Doc.Warnings() []string` |
| No validation after save | Library says OK, Word says corrupt | Run OPC validation after write (at least relationships + content types) |
| Unclear conformance class | User shares file, recipient can't open | Document the conformance class; provide `NewStrict()` and `NewTransitional()` constructors |
| Panic on malformed input | Library crashes the whole program | Return errors from every public function; never panic in library code |
| No `io.Reader`/`io.Writer` support | Forces file I/O only | Accept `io.ReaderAt` for reading, `io.Writer` for writing; optional `fs.FS` support |
| API that requires full re-parse for every edit | Perverse: "open, edit, save" re-parses whole package 3x | Document model is mutable in-memory; serialize once on save |

---

## "Looks Done But Isn't" Checklist

- [ ] **OPC Package:** Opens .docx but silently drops parts it doesn't understand. Verify: open a file with embedded fonts, ActiveX, glossary, mail merge, custom XML — re-save it. All intact?
- [ ] **Style Chain:** Copies `styles.xml` but not `basedOn` → `fontTable` → `numbering` → `theme` chain. Verify: open a template with Heading 2 (which inherits from Heading 1) — styles match?
- [ ] **Namespace Handling:** Parses elements with `w:` prefix only. Verify: open files produced by LibreOffice, Google Docs, Apache POI — they use different prefix conventions.
- [ ] **Relationship Integrity:** Adds an image and it shows. Verify: the relationship ID counter didn't collide with existing IDs; content type for image/png exists; the `.rels` Target is correct.
- [ ] **Round-trip Blanks:** Text round-trips correctly. Verify: leading/trailing spaces, double spaces, non-breaking spaces, tabs, and zero-width spaces all survive.
- [ ] **Conformance Class:** Creates new documents that open everywhere. Verify: open the output in Office 2007, 2010, 2016, 2019, 2021, LibreOffice, Google Docs, OnlyOffice.
- [ ] **XML Declaration:** Saves valid XML. Verify: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` is present, encoding matches, no BOM issues on Windows.
- [ ] **Part Ordering:** OPC requires parts in a specific order in the ZIP table. Verify: `[Content_Types].xml` is first, `.rels` is second, other parts follow — some ZIP tools fail otherwise (ECMA-376 Part 2 §9.1.4.2).

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Unknown elements dropped | HIGH — loss is permanent unless user has original file | 1. Implement pass-through preservation retroactively 2. For affected users: compare original vs output with XML diff, merge back |
| Relationship IDs collided | MEDIUM — Word may reject file | 1. Regenerate all rId values with unique prefix 2. Rebuild all `.rels` files 3. Re-serialize package |
| Template style chain broken | MEDIUM — output looks wrong | 1. Trace `basedOn` chain 2. Identify missing referenced parts 3. Copy them from template 4. Update content types |
| Wrong conformance class | LOW — re-save with correct class | 1. Detect source conformance 2. Remap namespace URIs 3. Re-serialize |
| Namespace prefix mismatch | LOW — re-decode with normalized prefixes | 1. Build prefix-agnostic decoder 2. Re-process document 3. Output with canonical prefixes |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Flat XML instead of OPC | Phase 1 (OPC Core) | Round-trip a .docx with images, headers, footer, styles — every part must be intact |
| C# → Go transliteration | Phase 0 (Architecture) | Code review: does API feel like Go (structs, errors, io.Reader/Writer) or C# (classes, streams, inheritance)? |
| Namespace prefix sensitivity | Phase 1 (OPC Core) | Read files from Word, LibreOffice, Google Docs — all must parse identically |
| Style resolution chain | Phase 3 (Style Preservation) | Apply template with Heading 2 (basedOn Heading 1) — verify cascade works |
| Unknown elements dropped | Phase 1 (OPC Core) | Round-trip file with ActiveX, VBA, custom XML, altChunk — all must survive |
| Transitional vs Strict | Phase 1 (OPC Core) | Open output in Office 2007, 2010, 2021 — all must open without error |
| Template style dependency graph | Phase 3 (Style Preservation) | Copy template with custom font, theme colors, numbering — output must match source |
| Whitespace/run collapse | Phase 2 (Document Model) | Test: "Hello  World" (double space), "  leading", "trailing  " — all survive |
| Relationship ID collision/breakage | Phase 1 (OPC Core) | Add 10 images, delete 3, add 5 more — no rId conflicts, all images visible |
| C# LINQ-to-XML query translation | Phase 0 (Architecture) | Decide XML query strategy before writing any content manipulation code |

---

## Sources

- **unioffice/issues/370** — "Word Fields not loaded" — demonstrates unsupported relationship type drops, silent field loss
- **unioffice/issues/401** — "Office2007 is not supported" — Transitional vs Strict conformance class incompatibility
- **nguyenthenguyen/docx/issues/27** — "panic: zip: not a valid zip file" — ZIP structure errors from naive implementation
- **nguyenthenguyen/docx/issues/40** — "Replace is very flaky" — inline text replacement weaknesses without proper XML structure awareness
- **nguyenthenguyen/docx/issues/30** — "Fattened replaced images" — image replacement without content type / relationship management
- **Eric White, OpenXmlSdkTs announcement (2026-04)** — notes that C# Open XML SDK class hierarchy doesn't port well, advocates for functional construction and LINQ-to-XML patterns
- **ECMA-376 Part 2 (OPC) 5th ed. December 2021** — OPC part naming rules, content type requirements, relationship structure, ZIP ordering requirements
- **ECMA-376 Part 1 (Fundamentals) 5th ed. December 2016** — WordprocessingML namespace definitions, style inheritance model, conformance classes
- **Microsoft "Introducing the Office (2007) Open XML File Formats" (AA338205)** — OPC package structure overview, relationship types, content types table
- **Open XML SDK for JavaScript / TypeScript (EricWhiteDev)** — demonstrates the difficulty of querying Open XML without LINQ; notes that recursive functional transforms are the cleanest approach
- **StackOverflow: Open XML SDK Pitfalls** — community reports on namespace handling, content type management, relationship ID issues
- **Go `encoding/xml` limitations** — standard library namespace registration is prefix-based, leading to fragility with alternate prefix conventions

---
*Pitfalls research for: wordingo .docx Library*
*Researched: 2026-07-25*
