# Architecture Research

**Domain:** Pure Go .docx library with style preservation
**Researched:** 2026-07-25
**Confidence:** HIGH

## Standard Architecture

### System Overview

Every .docx library handles the same fundamental layers. The OPC container format and OOXML schema define the boundaries:

```
┌─────────────────────────────────────────────────────────────────────┐
│                        API Layer (Public)                            │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────┐  ┌───────────┐  │
│  │ Document    │  │ Paragraph    │  │ Run        │  │ Table     │  │
│  │ (Open/Create│  │ (Add/Edit    │  │ (Text/     │  │ (Rows/    │  │
│  │  /Save)     │  │  /Delete)    │  │  Format)   │  │  Cells)   │  │
│  └──────┬──────┘  └──────┬───────┘  └─────┬──────┘  └─────┬─────┘  │
│         │                │                │               │        │
├─────────┴────────────────┴────────────────┴───────────────┴────────┤
│                      Style Resolution Layer                          │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │ StyleResolver: resolve run/paragraph properties through      │   │
│  │ style hierarchy (basedOn chain, theme, latent styles,        │   │
│  │ direct formatting override).                                 │   │
│  └──────────────────────────┬───────────────────────────────────┘   │
│                             │                                       │
├─────────────────────────────┴───────────────────────────────────────┤
│                      OPC Package Layer (ZIP)                         │
│  ┌───────────┐  ┌──────────────┐  ┌──────────────┐  ┌───────────┐  │
│  │ Package   │  │ Part         │  │ Relationship │  │ Content   │  │
│  │ (ZIP R/W) │  │ (XML Part    │  │ (Part-to-    │  │ Types     │  │
│  │           │  │  Abstraction)│  │  Part Links) │  │ (.xml)    │  │
│  └───────────┘  └──────────────┘  └──────────────┘  └───────────┘  │
└─────────────────────────────────────────────────────────────────────┘
         │                     │                     │
         ▼                     ▼                     ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      XML Part Files (ZIP entries)                     │
│                                                                      │
│  [Content_Types].xml    word/document.xml        word/styles.xml     │
│  word/numbering.xml     word/fontTable.xml       word/header1.xml    │
│  word/footer1.xml       word/theme/theme1.xml    docProps/core.xml   │
│  _rels/.rels            word/_rels/document.xml.rels                 │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Implementation Pattern |
|-----------|----------------|------------------------|
| **Package** | R/W .docx as ZIP archive; manage [Content_Types].xml and _rels/.rels; open/save lifecycle | Wraps Go `archive/zip`; maintains part index and relationship graph in memory |
| **Part** | Abstraction over a single XML part in the archive; each part has a content type, URI, and relationships to other parts | Interface with `ContentType()`, `URI()`, `Relationships()`, `Reader()`, `Writer()` |
| **ContentTypes** | Track part-to-content-type mapping; auto-register new parts; serialize [Content_Types].xml | In-memory map[string]string; rebuild on save |
| **Relationships** | Manage part-to-part relationship graph; `rId` tracking; relationship types | Directed graph with string IDs; ensure unique rId generation |
| **DocumentModel** | In-memory representation of document parts as Go structs; mirrors XML but with Go-idiomatic APIs | `Document` struct holds references to Body, Styles, Numbering, FontTable, Headers, Footers, etc. |
| **StyleResolver** | Resolve effective formatting for a paragraph/run: walk `basedOn` chain, merge latent/theme styles, apply direct formatting overrides | Recursive merge: Style.ParagraphProperties ⊕ Style.RunProperties, following basedOn. Cache resolved styles. |
| **NumberingResolver** | Resolve abstract numbering (`numPr`) → concrete numbering definition (`numFmt`, `lvlOverride`, `lvl`) | Map w:numId → w:abstractNumId → w:lvl list |
| **Paragraph** | High-level API for creating/editing paragraphs: runs, formatting, numbering, spacing | Wraps underlying CT_P XML; methods for AddRun(), SetStyle(), SetAlignment() |
| **Run** | High-level API for text runs: text content, font, size, bold/italic, color, underline | Wraps CT_R XML; fluent builder pattern for formatting properties |
| **Table** | High-level API for table creation: rows, cells, grid, borders, shading | Wraps CT_Tbl; row/column management with cell merging |
| **Serializer** | Convert in-memory model back to XML parts; ensure correct namespaces, preserve unknown elements during round-trip | XML marshal; skip/merge strategy for unknown elements encountered during read |
| **Deserializer** | Parse XML parts into in-memory model; handle OpenXML namespaces; parse style, numbering, font tables | XML unmarshal; namespace-aware; fallback on unknown elements (preserve for round-trip) |

## Recommended Project Structure

```
wordingo/
├── opc/                       # OPC package layer (ZIP archive I/O)
│   ├── package.go             # Package: open, create, save, close .docx
│   ├── part.go                # Part interface + implementations
│   ├── content_types.go       # [Content_Types].xml management
│   ├── relationships.go       # _rels/.rels management
│   ├── namespaces.go          # OpenXML namespace constants
│   └── package_test.go
│
├── docx/                      # Public API: document creation and editing
│   ├── document.go            # Document: Open(path), Create(path), Save(), Close()
│   ├── paragraph.go           # Paragraph: text, style, numbering, alignment
│   ├── run.go                 # Run: text content, formatting (bold, italic, font, size, color)
│   ├── table.go               # Table: grid, rows, cells, borders, shading
│   ├── section.go             # Section: page setup, headers, footers, columns
│   ├── header_footer.go       # Header/Footer management
│   ├── image.go               # Image embedding in runs
│   └── docx_test.go
│
├── style/                     # Style model and resolution
│   ├── style.go               # Style struct (paragraph/character/linked/table/list)
│   ├── resolver.go            # StyleResolver: resolve effective properties
│   ├── numbering.go           # Numbering definitions and abstract numbering
│   ├── font_table.go          # Font table handling
│   ├── theme.go               # Theme color resolution
│   └── resolver_test.go
│
├── schema/                    # Generated OOXML types (Go structs matching XSD)
│   ├── wordprocessingml/      # w: namespace types (CT_P, CT_R, CT_Style, CT_Tbl, etc.)
│   │   ├── types.go           # Auto-generated from OOXML XSD or handwritten
│   │   └── types_test.go
│   ├── drawingml/             # a: namespace types (colors, fonts, effects)
│   ├── math/                  # m: namespace (equations) — deferred
│   └── relationships/         # (unused; relationship types are string constants)
│
├── internal/                  # Internal utilities
│   ├── xmlutil/               # Namespace-aware XML marshal/unmarshal helpers
│   ├── ziputil/               # ZIP I/O utilities (buffered reads, path normalization)
│   ├── uuid/                  # UUID generation for part IDs
│   └── testutil/              # Test helpers (compare .docx files, extract XML for assertion)
│
├── cmd/                       # Optional CLI wrapper (deferred)
│   └── docxutil/              # Basic CLI for debugging (extract, inspect structure)
│
├── go.mod
└── go.sum
```

### Structure Rationale

- **opc/ separate from docx/:** OPC is format-agnostic. The same package can serve future .pptx or .xlsx support. Keeps ZIP handling, relationship tracking, and content type management isolated from Word-specific logic.
- **docx/ is the public API:** Users import `wordingo/docx` and interact with `Document`, `Paragraph`, `Run`. This is the idiomatic Go surface — no leaked OPC internals.
- **style/ separated from docx/:** Style resolution is the most complex non-trivial behavior in the library. Isolating it allows independent testing and makes the public API cleaner (docx methods call into style resolver, but users don't need to).
- **schema/ mirrors OOXML namespaces:** The w:, a:, m: namespace separation maps directly to OOXML schema files. Each namespace is a sub-package. Types are flat structs with XML tags — no methods. This is the same pattern unioffice uses.
- **internal/ for private helpers:** XML utilities, ZIP utilities, and test infrastructure are not part of the public API contract. Keeps the public surface small.

## Architectural Patterns

### Pattern 1: Wrapper-over-Schema (Two-Layer API)

**What:** Schema types (in schema/) are exact mirrors of OOXML XML elements — verbose, deeply nested, XML-tagged Go structs. The docx/ layer wraps these with Go-idiomatic methods. Users never interact with schema types directly; the `X()` escape hatch exposes them when needed.

**When to use:** Always. Direct exposure of schema types makes the API unusable (e.g., creating a paragraph requires constructing 4+ nested structs with XML namespace attributes).

**Trade-offs:**
- + Clean public API: `doc.AddParagraph().AddRun().SetText("Hello").SetBold(true)`
- + Escape hatch: `p.X()` returns underlying `CT_P` for advanced manipulation
- - Writing wrapper code is tedious but mechanical. See "Code Generation" below.

**Example:**
```go
// Public API (docx/) — user friendly
p := doc.AddParagraph()
p.SetStyle("Heading1")
r := p.AddRun()
r.SetText("Hello World")
r.SetBold(true)

// Underlying (schema/) — raw OOXML
type CT_P struct {
    XMLName xml.Name `xml:"w:p"`
    PPr     *CT_PPr  `xml:"w:pPr,omitempty"`
    Content []*CT_R  `xml:"w:r"`
    // ... many more fields
}

// Escape hatch when API is insufficient:
raw := p.X() // *CT_P
raw.PPr.PStyle = &CT_String{Val: "Heading1"}
```

### Pattern 2: Lazy Part Loading

**What:** When opening an existing .docx, don't deserialize every part into memory. Parse only [Content_Types].xml and _rels/.rels upfront; deserialize document.xml, styles.xml, numbering.xml on first access. Store other parts (headers, footers, images) as raw bytes until touched.

**When to use:** Always. A .docx can have dozens of parts. Parsing all at once is slow and wasteful for typical workflows (e.g., "change title text" shouldn't parse every header/footer).

**Trade-offs:**
- + Fast open times for large documents
- + Lower memory footprint
- - Simple API requires lazy initialization checks. Mitigate: use accessor methods (`doc.Styles()` that lazy-loads and caches).

**Example:**
```go
type Document struct {
    pkg    *opc.Package
    body   *DocumentBody   // lazy, loaded on first body access
    styles *Styles         // lazy, loaded on first styles access
    mu     sync.Mutex
}

func (d *Document) Styles() (*Styles, error) {
    d.mu.Lock()
    defer d.mu.Unlock()
    if d.styles != nil {
        return d.styles, nil
    }
    part, err := d.pkg.GetPart("/word/styles.xml")
    if err != nil {
        return nil, err
    }
    d.styles = parseStyles(part.Reader())
    return d.styles, nil
}
```

### Pattern 3: Immutable Schema with Mutable Wrappers

**What:** Schema types are value objects — no methods, no mutation. Wrapper types in docx/ hold a pointer to the schema type and mutate it through setters. On save, the wrappers serialize the modified schema types back to XML.

**When to use:** Core pattern for document editing. Separates the "what" (schema type state) from "how" (edit operations).

**Trade-offs:**
- + Simple serialization: just marshal the schema struct
- + Thread-safe boundaries (if wrappers synchronize)
- - Two objects for every document element. Acceptable given the complexity of OOXML.

**Example:**
```go
type Run struct {
    ct *CT_R          // owned schema type, modified in place
    doc *Document     // back-reference for style resolution
}

func (r *Run) SetBold(bold bool) *Run {
    if r.ct.RPr == nil {
        r.ct.RPr = &CT_RPr{}
    }
    if r.ct.RPr.B == nil {
        r.ct.RPr.B = &CT_OnOff{}
    }
    r.ct.RPr.B.Val = &bold
    return r
}

func (r *Run) SetText(text string) *Run {
    // Clear existing text elements, add new one
    r.ct.Content = nil
    r.ct.Content = append(r.ct.Content, &CT_Text{Content: text})
    return r
}
```

### Pattern 4: Round-Trip Preservation via Unknown Element Hoarding

**What:** When reading XML parts, store any element that doesn't map to a known schema type in a `[]Any` or `[]xml.Token` slice. On write, re-emit those unknown elements unchanged. This ensures the library doesn't corrupt documents that use features it doesn't model.

**When to use:** Required for production-grade document editing. Without this, editing and saving a document with any unsupported feature (custom XML markup, legacy fields, third-party extensions) destroys that content.

**Trade-offs:**
- + Safe round-trip. Documents survive edit-save cycles intact.
- + Future-proof against new OOXML features.
- - Slightly larger memory footprint. Storage for unknown elements is proportional to document complexity, not library features.
- - Harder to implement with `encoding/xml`. Mitigate: use `xml.Decoder` + `xml.Encoder` token stream approach, not struct-level marshal/unmarshal.

**Example:**
```go
type CT_P struct {
    XMLName xml.Name `xml:"w:p"`
    PPr     *CT_PPr  `xml:"w:pPr,omitempty"`

    // Known child types
    Runs    []*CT_R       `xml:"w:r"`
    Hyperlinks []*CT_Hyperlink `xml:"w:hyperlink"`

    // Unknown element hoarding
    Unknown []CustomElement `xml:",any"` // anything not matched above
}

type CustomElement struct {
    XML       []byte // raw XML tokens for re-encoding
    Namespace string
}
```

## Data Flow

### Read Flow: .docx → In-Memory Model

```
.docx file (ZIP)
    │
    ▼
[opc.Package.Open()]
    │
    ├── Read ZIP entries → index of parts (URI → offset/size)
    ├── Parse [Content_Types].xml → content type map
    ├── Parse _rels/.rels → package-level relationships
    │
    ▼
[docx.Document.Open(path)] wraps the opc.Package
    │
    ├── Lazy: Parse word/_rels/document.xml.rels → part relationship graph
    ├── Lazy: Parse word/document.xml → Body (paragraphs, tables, sections)
    │         Unknown elements → hoarded in .Unknown fields
    ├── Lazy: Parse word/styles.xml → style definitions → StyleResolver
    ├── Lazy: Parse word/numbering.xml → numbering definitions
    ├── Lazy: Parse word/fontTable.xml → font table
    │
    ▼
In-memory model (docx.Document):
    ┌────────────────────────────────────────────┐
    │  Body: []ContentBlock (Paragraph | Table)   │
    │  Styles: map[string]*Style                  │
    │  Numbering: map[int]*NumberingDef           │
    │  FontTable: []*Font                         │
    │  Headers/Footers: map[string]*HeaderFooter  │
    │  Images, Themes: lazy-loaded parts          │
    └────────────────────────────────────────────┘
```

### Write Flow: In-Memory Model → .docx

```
User modifies document via docx API
    │
    ├── doc.AddParagraph() → creates CT_P, appends to Body
    ├── p.AddRun().SetText("...") → creates CT_R on CT_P
    ├── style changes update style resolver cache
    │
    ▼
[Document.Save() or Document.SaveTo(w)]
    │
    ├── 1. Sync: write in-memory parts back to XML
    │   ├── document.xml  ← marshal Body (all CT_P, CT_Tbl, etc.)
    │   ├── styles.xml    ← marshal Styles
    │   ├── numbering.xml ← marshal Numbering
    │   ├── fontTable.xml ← marshal FontTable
    │   └── other dirty parts (headers, footers, etc.)
    │
    ├── 2. Don't touch clean parts (lazy-loaded and unmodified)
    │
    ├── 3. Rebuild [Content_Types].xml from part index
    │
    ├── 4. Rebuild all _rels/.rels from relationship graph
    │
    ├── 5. Write ZIP archive
    │   ├── Re-order entries to match Word expectations:
    │   │   1. [Content_Types].xml (first entry)
    │   │   2. _rels/.rels
    │   │   3. word/document.xml
    │   │   4. word/_rels/document.xml.rels
    │   │   5. Remaining parts (styles, numbering, etc.)
    │   │   6. docProps/ (last)
    │   └── Deflate compress each entry
    │
    ▼
Valid .docx file
```

### Create Flow: Template-Based Document

```
docx.CreateFromTemplate(templatePath string) → *Document
    │
    ├── Open template.docx as Package (read-only copy)
    ├── Clone all parts into writable package
    ├── Preserve template styles.xml, numbering.xml unchanged
    ├── Clear body content (keep empty body with default section)
    │
    ▼
Ready for content addition:
    doc.AddParagraph().AddRun().SetText("New content")
```

### Style Resolution Data Flow

```
Given a paragraph with style="Heading1" and paragraphProperties:
    │
    ▼
[StyleResolver.Resolve(p)]
    │
    ├── 1. Lookup style "Heading1" in Styles map
    ├── 2. Follow basedOn chain: Heading1 → Heading → Normal
    │      (collect styles in order, detect cycles)
    ├── 3. Merge paragraph properties:
    │      Normal (base) ← Heading ← Heading1 ← direct (highest priority)
    ├── 4. Merge run properties:
    │      DefaultParagraphFont ← Normal ← Heading ← Heading1 ← run direct
    │
    ▼
Result: EffectiveCT_PPr (all resolved paragraph formatting)
        EffectiveCT_RPr (all resolved run formatting)
```

## DocumentFormat.OpenXml (C#) → Go Mapping

### How the .NET SDK Structures OOXML Access

Microsoft's `DocumentFormat.OpenXml` SDK defines the canonical part architecture. Key patterns and their Go equivalents:

| C# Pattern | Go Equivalent | Notes |
|-----------|---------------|-------|
| `WordprocessingDocument.Open(path, true)` | `opc.Package.Open(path)` + `docx.Document.Open(pkg)` | Go: two-step open separates OPC from docx semantics |
| `WordprocessingDocument.Create(path, type)` | `opc.Package.Create(path)` + `docx.Document.Create(pkg)` | Same two-step for creation |
| `MainDocumentPart.Document.Body.Append(child)` | `doc.Body().AppendParagraph(p)` | Go: method-based, not property-based |
| `MainDocumentPart.AddNewPart<StylePart>()` | `pkg.CreatePart("/word/styles.xml", ct)` | Go: explicit URI + content type |
| `StylePart.Style = new Styles(...)` | `stylesPart := schema.NewStyles(); ...` | Schema types are constructable directly |
| `style.BasedOn.Val = "Normal"` | `style.SetBasedOn("Normal")` | Go wrapper method vs direct property set |
| `Paragraph p = new Paragraph(new Run(new Text("hi")))` | `p := doc.AddParagraph(); p.AddRun().SetText("hi")` | Go: builder pattern, not constructor nesting |
| `using (WordprocessingDocument doc = ...)` | `defer doc.Close()` | Go: defer idiom instead of IDisposable |
| Part relationships via `AddPart(sourcePart, targetPart, rId)` | `pkg.AddRelationship(sourceURI, targetURI, relType, rId)` | Go: explicit URI-based, not type-based |

### Key Architectural Differences

1. **C#: Strongly-typed part classes.** `MainDocumentPart`, `StylePart`, `NumberingPart` are concrete types with typed properties. Go: OPC layer is generic; word-specific types live in docx/ and schema/. More flexible for extension, harder to discover without docs.

2. **C#: Constructor-nesting API.** `new Paragraph(new Run(new Text("hi")))`. Go: Builder pattern with method chaining. More lines but clearer intent.

3. **C#: Explicit part creation.** `MainDocumentPart.AddNewPart<StylePart>()` auto-generates URIs and rIds. Go: Manual URI management. Trade: more control vs more boilerplate. Mitigate: auto-increment helpers.

4. **C#: XML-backed model.** DocumentFormat.OpenXml's types update XML tree in-place. Go: Schema types marshal/unmarshal. Save is explicit: serialize to XML, write to ZIP.

5. **C#: Full schema coverage.** `DocumentFormat.OpenXml` has types for every OOXML element. Go: Start with common subset; extend as needed. Unmapped elements survive via round-trip hoarding.

### Capabilities to Adopt (Behavior, Not Code)

| Capability | Go Implementation Strategy |
|------------------|---------------------------|
| L1 semantic views (text, outline, issues) | docx/ methods: `Paragraph.Text()`, `Document.Text()`, `Document.Outline()` |
| L2 DOM operations (query, set, add) | Fluent builder API + path-like element addressing |
| L3 raw XML fallback | `X()` escape hatch on every wrapper |
| Template-based creation | `docx.CreateFromTemplate(path)` — clone all template parts |
| Style preservation | Full styles.xml parser with basedOn resolution chain |
| Round-trip dump | JSON serialization of document model |
| Watch mode | Future: `fsnotify` + `Document.Save()` loop |

## Build Order

The build must respect dependency hierarchy. Each layer depends only on layers below it.

### Phase 1: Foundation — opc package (no dependencies)
```
opc/package.go        # Open, Create, Save, Close
opc/part.go           # Part interface, InMemoryPart, FilePart
opc/content_types.go  # Content type map, defaults, overrides
opc/relationships.go  # Relationship graph, rId generation
opc/namespaces.go     # Constants
```
**Test:** Create a ZIP, add a part, verify the ZIP is valid (unzip -t).
**Gate:** Can create a minimal (empty) .docx that Word opens without error.

### Phase 2: Schema types (no dependencies on other project packages)
```
schema/wordprocessingml/types.go   # CT_P, CT_R, CT_Text, CT_PPr, CT_RPr, CT_Body, CT_Styles, etc.
schema/wordprocessingml/simple.go  # Simple types (ST_String, ST_HexColor, etc.)
```
**Approach:** Hand-write the essential subset (~50 types). Generate the rest from XSD if time allows.
**Gate:** Can marshal/unmarshal a minimal document.xml to/from Go structs.

### Phase 3: Style & Numbering model and resolution
```
style/style.go          # Style struct
style/resolver.go       # Resolve: basedOn chain, latent styles, theme colors
style/numbering.go      # NumberingDefinition, Level
style/font_table.go     # FontTable
```
**Gate:** Given a styles.xml, correctly resolve effective paragraph and run properties.

### Phase 4: Document model (depends on schema + style + opc)
```
docx/document.go        # Document struct, Open(), Create(), Save()
docx/paragraph.go       # Paragraph wrapper
docx/run.go             # Run wrapper
docx/table.go           # Table wrapper (basic)
docx/section.go         # Section properties
docx/header_footer.go   # Header/Footer wrappers
```
**Gate:** Can open an existing .docx, read text and styles, then save back with identical content (binary comparison of extracted XML parts).

### Phase 5: Editing and creation API
```
docx/ — Add methods on wrappers
  - Paragraph.AddRun(), Run.SetText(), Run.SetBold(), etc.
  - Document.AddParagraph(), Document.AddTable()
  - Style assignment: Paragraph.SetStyle(name)
  - Template-based creation: Document.CreateFromTemplate(path)
```
**Gate:** Can create a .docx from scratch with styled paragraphs and tables, and open it in Word with correct formatting.

### Phase 6: Advanced features
```
docx/image.go           # Image embedding
docx/ — nested tables, cell merging
docx/ — header/footer content editing
style/theme.go          # Theme color resolution
```
**Gate:** Can create a document with images, complex tables, and custom headers/footers.

## Anti-Patterns

### Anti-Pattern 1: Monolithic Single-Type XML Parse

**What people do:** Parse the entire .docx into one giant struct tree (one struct per XML element in document.xml).

**Why it's wrong:** Go's `encoding/xml` struggles with deeply nested, polymorphic XML (paragraphs contain runs contain text, but also bookmarks, comments, hyperlinks, etc.). The resulting struct is fragile, slow to compile, and brittle against schema changes.

**Do this instead:** Parse document.xml as a flat sequence of paragraphs and tables. Within each paragraph, parse runs but hoist unknown children into `[]CustomElement`. Accept that you're modelling the OOXML you use, not the full spec.

### Anti-Pattern 2: Heavy XML Tree Manipulation

**What people do:** Build an in-memory XML tree (like C#'s `XmlDocument`) and manipulate it via DOM methods (CreateElement, AppendChild etc.).

**Why it's wrong:** Go doesn't have a standard DOM-like tree with namespace support that matches OOXML's needs. `encoding/xml` is stream-oriented. Third-party DOM libs are not widely adopted.

**Do this instead:** Schema structs with XML tags + marshal/unmarshal. For round-trip safety, hoist unknown elements. This is Go-idiomatic and avoids DOM complexity.

### Anti-Pattern 3: Direct ZIP Entry Order Ignorance

**What people do:** Write ZIP entries in any order, expecting Word to handle it.

**Why it's wrong:** Word and other consumers expect a specific entry order: [Content_Types].xml must be first, then _rels/.rels, then word/document.xml, then remaining parts, then docProps/. Writing out of order can cause "Office cannot open this file" errors from certain validators or older Word versions.

**Do this instead:** Enforce entry ordering on save. Write entries in the documented canonical order. The ZIP format supports ordered entries even though most ZIP libraries don't enforce it — write them in the right order.

### Anti-Pattern 4: Copying DocumentFormat.OpenXml's Type Hierarchy

**What people do:** Build a Go type hierarchy that mirrors C#'s `OpenXmlElement` → `OpenXmlPart` → `MainDocumentPart` etc.

**Why it's wrong:** Go doesn't do inheritance well. Deep type hierarchies fight Go's composition-over-inheritance philosophy. The result is confusing: interface casts everywhere, hard-to-follow method resolution.

**Do this instead:** Flat Go packages with concrete types. opc.Package (not OPC Package with abstract base). docx.Document (not Document inheriting from OpenXmlPart). Use interfaces for behavior, not taxonomy.

## Integration Points

### External Dependencies

| Dependency | Purpose | Risk |
|------------|---------|------|
| `archive/zip` (stdlib) | Read/write ZIP | None. Stdlib, no CGO. |
| `encoding/xml` (stdlib) | XML parse/serialize | Must handle namespaces carefully. OOXML uses multiple namespaces on every element. Go's encoder doesn't handle multiple default namespaces well; must use explicit prefixes. |
| None | Pure Go standard library only | Intentional. No third-party XML or ZIP libraries needed. |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| `opc.Package` ↔ `docx.Document` | opc.Package is a constructor arg; Document wraps it | Strict one-directional: opc knows nothing about docx. |
| `docx.Document` ↔ `style.Resolver` | Document holds a Resolver instance | Resolver is dependency-injected or created during Open(). |
| `docx.*` wrappers ↔ `schema.*` | Wrappers hold pointer to CT_* struct | Wrappers mutate schema types directly. On save, Document marshals the schema types. |
| `opc.Package` ↔ `archive/zip` | Package wraps `*zip.Writer` and `*zip.Reader` | Direct 1:1 mapping. Package adds content type and relationship management on top. |

## Scaling Considerations

This is a library, not a service. "Scaling" is about document complexity, not user count:

| Concern | Small doc (<100KB, <10 parts) | Large doc (>10MB, >100 parts) | Huge doc (>100MB, >500 parts) |
|---------|------|------|------|
| **Open time** | <10ms (parse all) | ~50ms (lazy parse) | ~200ms (lazy parse, skip images) |
| **Memory** | ~200KB | ~10MB (parts in memory) | ~100MB+ (stream to disk for images) |
| **Save time** | <5ms | ~100ms | ~1s+ |
| **Style resolution cache** | Trivial | 100-500 styles → cache hit dominates | Same. Styles don't scale with doc size. |

**Key insight:** Document complexity is bounded by what Word handles. Word struggles with documents over ~100MB. The library should handle anything Word can. The main scaling concern is: keep binary size reasonable. The 33MB unioffice binary is a warning — code generation of all OOXML types is expensive. Limit schema to wordprocessingML only (not xlsx/pptx types). Use code generation selectively.

## Sources

- **unioffice (Go):** Reference for Go-idiomatic OOXML library structure. Layer separation (schema/ → document/ → internal/). Escape hatch pattern. Confirmed via source structure on GitHub.
- **DocumentFormat.OpenXml (C#):** Reference for part architecture and style inheritance behavior. Confirmed via the .NET Open XML SDK source structure.
- **OOXML / ISO 29500:** OPC packaging standard (ZIP + XML parts + relationships). Documented on Wikipedia and ECMA/ISO specifications.
- **DocumentFormat.OpenXml:** Microsoft's official .NET SDK. Source of truth for part hierarchy and lifecycle patterns.
- **gooxml (former unioffice predecessor):** Early Go OOXML library. Influenced unioffice's architecture. No longer maintained.

---

*Architecture research for: wordingo (pure Go .docx library)*
*Researched: 2026-07-25*
