# Phase 3: Document Model - Research

**Researched:** 2026-07-25
**Domain:** OPC document lifecycle — open/read/save round-trip + template cloning
**Confidence:** HIGH

## Summary

Phase 3 delivers the public `wordingo/` API package's `Document` wrapper that bridges Phase 1's OPC layer and Phase 2's style engine. The OPC layer already guarantees round-trip fidelity (OPC-04 raw pass-through of unmodified parts) and the style cloner (`CloneStyles`) already byte-copies the 5 style parts. Phase 3's job is orchestration: open an existing package, parse `word/document.xml` eagerly into `CT_Document`, expose `Paragraphs()` as read-only wrappers over `*wml.CT_P`, and ensure Save never triggers `MarkModified` on style parts.

FromTemplate clones the style dependency graph via Phase 2's `CloneStyles`, then clears or keeps the body depending on function called. OpenTemplate preserves existing body content. Both use the same clone mechanism — the difference is body policy only.

Round-trip verification reuses Phase 1's `opc.DiffParts()` pattern. Style parts protection is automatic via OPC Save's raw-copy behavior — Document never calls `MarkModified` on style parts during read-only or FromTemplate paths.

**Primary recommendation:** Implement `Document` as a thin orchestrator over `*opc.Package`. Parse `word/document.xml` eagerly in `Open`. Never touch style parts. Reuse `opc.DiffParts` for round-trip CI.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Create public `wordingo/` package in this phase with read-only `Document` wrapper. Phase 4 adds mutation methods.
- **D-02:** `Document.Paragraphs() []*Paragraph` returns typed paragraph wrappers over `*wml.CT_P`. Each `Paragraph` exposes read-only accessors: `Style()`, `Text()`, `X()` escape hatch. No Table support in Phase 3.
- **D-03:** `word/document.xml` parsed eagerly into `CT_Document`/`CT_Body` on `Open()`. Supporting parts (headers, footers, numbering, settings, styles, theme, fontTable) stay lazy — parsed on first access via the OPC layer's existing lazy `Part.Open()`. Matches PRD NFR (<50ms for 100-page doc).
- **D-04:** Two separate functions:
  - `FromTemplate(path)` — clones template style graph, clears body (CREATE-03)
  - `OpenTemplate(path)` — clones template style graph, keeps existing body (CREATE-04)
  Both available as reader variants: `FromTemplateReader(r, size)`, `OpenTemplateReader(r, size)`.
- **D-05:** Per-part byte diff (same as Phase 1 D-04). Open fixture → save to buffer → unzip both → diff each part byte-identical. Not whole-file ZIP diff (ZIP metadata varies).
- **D-06:** Automatic via OPC Save raw-copy behavior. Document never calls `opc.MarkModified` on style parts (styles, numbering, fontTable, theme, settings) during read-only or FromTemplate paths. Only explicit user style mutation (Phase 4+) triggers MarkModified. No extra guard layer needed.
- **D-07:** Full PRD §8 function set for Phase 3:
  - `Open(path string) (*Document, error)`
  - `OpenReader(r io.ReaderAt, size int64) (*Document, error)`
  - `Create() (*Document, error)` (already exists)
  - `FromTemplate(path string) (*Document, error)`
  - `FromTemplateReader(r io.ReaderAt, size int64) (*Document, error)`
  - `OpenTemplate(path string) (*Document, error)`
  - `OpenTemplateReader(r io.ReaderAt, size int64) (*Document, error)`
  - `(*Document).Paragraphs() []*Paragraph`
  - `(*Document).Save(path string) error` (already exists)
  - `(*Document).WriteTo(w io.Writer) (int64, error)`
  - `(*Document).Warnings() []string` (already exists)
  - `(*Document).Close() error`
- **D-08:** Minimal round-trip fixture set: 3-4 docx files under `testdata/roundtrip/` — blank, single-paragraph, multi-heading, header+footer. Used for per-part byte diff CI tests.

### the agent's Discretion
- Internal file/function layout within `wordingo/` package
- Exact read-only accessor shape on `Paragraph` (Style() string, Text() string, X() *wml.CT_P)
- Fixture filenames and location within `testdata/roundtrip/`
- Whether to store raw body bytes for eager parse + diff-check or parse to CT_Document only
- `OpenTemplate` vs `OpenTemplateReader` naming (confirm with user during plan)

### Deferred Ideas (OUT OF SCOPE)
- Document.Tables() — Phase 5
- Paragraph mutation (AddRun, SetStyle, formatting) — Phase 4
- Edit operations (insert/delete, replace text) — Phase 6
- Tables in body iteration — Phase 5
- Template merge ({{placeholder}}) — Phase 6
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| STYLE-ROUNDTRIP-01 | Editing a document never rewrites style parts unless user explicitly modifies styles | OPC-04 raw pass-through handles this. Document never calls MarkModified on style parts during read-only paths. Style parts start as unmodified (sourced from `zip.File`), Save raw-copies them. D-06 confirms no extra guard needed. |
| STYLE-ROUNDTRIP-02 | Unmodified content formatting is byte-identical after edit + save | Same mechanism — OPC Save raw-copies any Part where `modified==false`. Only body parts modified by content changes get serialized from replacement data. Per-part diff (D-05) proves this. |
| CREATE-03 | FromTemplate(path) clones template package, preserves style/theme/numbering parts, provides ready body | `CloneStyles` (Phase 2) byte-copies the 5 style parts. FromTemplate then clears body (builds new empty CT_Body with one sectPr). Existing `buildDocumentXML()` generates minimal body. |
| CREATE-04 | Pre-populated template support — open template, keep existing body content, insert at specified locations | `OpenTemplate` uses `CloneStyles` same as FromTemplate, but copies source's `word/document.xml` body content instead of clearing it. Source document.xml bytes read via `Part.Open()`, unmarshaled, body kept. |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Open existing document | API / Backend | — | `Document.Open()` orchestrates `opc.Open()` + eager body parse. Pure data-in/data-out. |
| Read body paragraphs | API / Backend | — | `Document.Paragraphs()` wraps parsed `CT_Body.P` — no rendering, no browser. |
| Round-trip save | API / Backend | Database / Storage | `Document.Save()` delegates to `opc.Package.Save()` — ZIP serialization with raw pass-through. |
| FromTemplate | API / Backend | — | Orchestrates `opc.Open(source)` + `CloneStyles` + body clear. No persistence tier. |
| OpenTemplate | API / Backend | — | Same as FromTemplate but copies body bytes. No persistence tier. |
| Lazy part loading | API / Backend | — | Supporting parts parsed on first `Part.Open()` call — triggered by consumer code. |

## Standard Stack

### Core

Zero external dependencies — Go stdlib only [VERIFIED: go.mod].

| Package | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/fabiomarini/wordingo` (root) | module root | Public `Document`/`Paragraph` API | PRD §8, D-01: single public package |
| `internal/opc` | existing | ZIP I/O, content types, relationships, raw pass-through | Phase 1 — round-trip foundation |
| `internal/wml` | existing | `CT_Document`, `CT_Body`, `CT_P`, `CT_R`, `CT_Text` | Phase 1 — body type definitions |
| `internal/style` | existing | `CloneStyles` byte pass-through | Phase 2 — template style cloning |
| `internal/xmlutil` | existing | `NewEncoder`, namespace prefix rewrite | Phase 1 — XML serialization for body re-encoding |

**Version verification:**
```bash
go version        # go1.26.5 (project requires 1.23+)
```
All `internal/` packages are local to the module — no version to verify.

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| OPC raw pass-through (OPC-04) | Re-serialize all parts on save | Would destroy byte-identity — defeats STYLE-ROUNDTRIP. OPC-04 is the correct approach. |
| Eager parse of document.xml | Keep document.xml lazy too | Would break `Paragraphs()` access — either parse on first call (complexifies API contract) or parse eagerly. Eager parse on Open is simpler, matches PRD NFR targets. |
| CloneStyles vs manual part-by-part copy | Copy each style part manually | CloneStyles already exists, tested, handles edge cases (absent source parts, content types, relationship wiring). Don't duplicate. |

## Package Legitimacy Audit

> This phase installs zero external packages. All code uses Go stdlib + internal packages.

| Package | Registry | Verdict | Disposition |
|---------|----------|---------|-------------|
| *(none)* | — | N/A | Zero external dependencies — Go stdlib only |

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Architecture Patterns

### System Architecture Diagram

```
Open(path) / OpenReader(r, size)
    │
    ▼
opc.Open(r, size)  ──►  ZIP parsed, parts enumerated,
    │                      [Content_Types].xml + .rels eager,
    │                      part payloads lazy
    ▼
Read word/document.xml via Part.Open()
    │
    ▼
xmlutil.NewSafeDecoder → xml.Unmarshal → *wml.CT_Document
    │                     (eager — D-03)
    ▼
Document{ pkg, doc }
    │
    ├── Paragraphs() ──► iterate doc.Body.P → []*Paragraph{ *wml.CT_P }
    │
    ├── Save / WriteTo ──► pkg.Save(w)
    │                          ├── unmodified parts → zip.Copy (OPC-04 byte-identical)
    │                          └── modified parts → MarkModified data → serialize
    │
    └── Close ──► nil pkg ref

FromTemplate / OpenTemplate(path)
    │
    ▼
opc.Open(source) → sourcePkg
    │
    ├── Create() newBlankPackage() → dstPkg  (or use Create to get a fresh target)
    │
    ├── CloneStyles(sourcePkg, dstPkg)  ──► byte-copies 5 style parts,
    │                                       adds content types + relationships
    │
    ├── FromTemplate: dstPkg.MarkModified("word/document.xml", newEmptyBodyXML)
    │   OpenTemplate:  read source document.xml → parse → keep body → re-encode → MarkModified
    │
    └── Document{ pkg: dstPkg, doc: parsedDocument }
```

### Recommended Project Structure

```
wordingo/                          # Public root package
├── wordingo.go                    # D-01: Document type, Create(), Save, SaveFile, Warnings, X (exists)
├── create.go                      # newBlankPackage(), buildDocumentXML(), constants (exists)
├── open.go                        # D-07: Open(), OpenReader() — eager body parse
├── template.go                    # D-04/D-07: FromTemplate(), FromTemplateReader(),
│                                  #           OpenTemplate(), OpenTemplateReader()
├── paragraph.go                   # D-02: Paragraph type — Style(), Text(), X()
├── doc.go                         # Package documentation
├── create_test.go                 # Existing Create() tests
├── roundtrip_test.go              # D-05: Per-part byte diff round-trip tests
├── template_test.go               # FromTemplate/OpenTemplate tests
└── defaults/                      # Existing default XML parts
```

### Pattern 1: Wrapper-over-schema with `X()` escape hatch

**What:** Every public API type wraps an internal WML struct. Methods provide safe read access. `X()` returns the underlying schema type for escape-hatch access.

**When to use:** All public API types in this project — Document wraps `*opc.Package`, Paragraph wraps `*wml.CT_P`, Run wraps `*wml.CT_R`.

**Example:**
```go
// Source: Phase 1 established pattern in wordingo.go, create.go
// Phase 3 extends to Paragraph.

type Document struct {
    pkg *opc.Package
    doc *wml.CT_Document  // eagerly parsed body (D-03)
}

func (d *Document) X() *opc.Package { return d.pkg }

type Paragraph struct {
    ct *wml.CT_P
}

func (p *Paragraph) Style() string {
    if p.ct.PPr != nil && p.ct.PPr.PStyle != nil && p.ct.PPr.PStyle.Val != nil {
        return *p.ct.PPr.PStyle.Val
    }
    return ""
}

func (p *Paragraph) Text() string {
    // Walk runs collecting text values, respecting run boundaries (WML-03)
    var b strings.Builder
    for _, r := range p.ct.R {
        if r.T != nil {
            b.WriteString(r.T.Value)
        }
        // Phase 4: also handle line breaks, tabs, etc.
    }
    return b.String()
}

func (p *Paragraph) X() *wml.CT_P { return p.ct }
```

### Pattern 2: Template cloning via CloneStyles byte pass-through

**What:** FromTemplate/OpenTemplate create a fresh package via `Create()`, then call `CloneStyles(source, dst)` to byte-copy the 5 style parts. The difference is body policy only.

**When to use:** All template-origin document creation.

**Example:**
```go
func FromTemplate(path string) (*Document, error) {
    srcPkg, err := opc.Open(...)   // open template
    if err != nil { return nil, err }

    dstPkg := newBlankPackage()    // fresh target (Create() internals)

    if err := style.CloneStyles(srcPkg, dstPkg); err != nil {
        return nil, err
    }

    // For FromTemplate: empty body (one section, no paragraphs)
    // For OpenTemplate: copy source body content
    emptyBodyXML := buildEmptyBodyXML()
    dstPkg.MarkModified("word/document.xml", emptyBodyXML)

    return &Document{pkg: dstPkg}, nil
}
```

### Anti-Patterns to Avoid

- **Calling MarkModified on style parts during read-only paths:** Destroys byte-identity. Style parts are unmodified. Only mark body modified.
- **Full-ZIP byte comparison for round-trip:** ZIP metadata (timestamps, compression) varies across saves. Use per-part diff (D-05).
- **Re-serializing all parts on save:** OPC-04 raw pass-through exists for a reason. Let zip.Copy handle unmodified parts.
- **Two-phase clone that modifies target on partial failure:** CloneStyles already implements atomic precondition scan (D-08). Don't build another.
- **Storing body as both raw bytes and parsed CT_Document:** Either parse to CT_Document (preferred for access) or keep raw bytes and parse on demand. D-03 says parse eagerly. If you keep raw bytes for diff-check too, that's two sources of truth — pick one.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| ZIP reading/writing | Custom ZIP reader/writer | `archive/zip` (stdlib) | Handles deflate, CRC, central directory, streaming. |
| XML namespace prefix normalization | Prefix-mapping logic | `internal/xmlutil` (exists) | Phase 1 built URI-based namespace registry with canonical prefix rewrite. Already battle-tested. |
| Style cloning (5-part dependency graph) | Manual part-by-part copy | `internal/style.CloneStyles()` | Exists, tested in Phase 2. Handles absent source parts, content types, relationship wiring, atomic precondition scan. |
| Per-part byte diff comparison | Custom ZIP comparison | `opc.DiffParts()` (exists) | Phase 1 built it for round-trip tests. Excludes manifests ([Content_Types].xml, .rels). Ready to import. |
| Safe XML decoding (entity expansion) | Entity expansion limits | `xmlutil.NewSafeDecoder()` (exists) | Phase 1 built it. Protects against billion laughs, quadratic blowup. |
| OPC raw pass-through | Re-serialize unmodified parts | `opc.Package.Save()` `zip.Copy` | Phase 1 OPC-04. Untouched parts byte-identical via zip.Writer.Copy. |

**Key insight:** Phase 1 and Phase 2 already built all the machinery Phase 3 needs. The Document type is an orchestrator — it wires existing pieces together without inventing new infrastructure. Every "Don't Hand-Roll" item here is an existing, tested component.

## Common Pitfalls

### Pitfall 1: Accidental MarkModified on style parts

**What goes wrong:** Document.Open reads and parses word/document.xml. The body parse is correct. But somewhere in the flow, a developer calls `pkg.MarkModified("word/styles.xml", ...)` on the re-encoded styles content. This destroys byte-identity for the styles part (STYLE-ROUNDTRIP-01 fails).

**Why it happens:** Temptation to "normalize" or "pretty-print" style parts on save. The pattern for body modification (re-encode body, MarkModified) gets misapplied to style parts.

**How to avoid:** D-06 rule: Never call MarkModified on style parts during read-only or FromTemplate paths. If you're not modifying styles, don't touch them. The OPC layer handles pass-through automatically for parts where `modified==false`.

**Warning signs:** Round-trip test failing on `word/styles.xml` or `word/theme/theme1.xml` byte difference. Grep for `MarkModified("word/styles"` — it should only appear in Phase 4 (style mutation) code.

### Pitfall 2: Eager parse of document.xml mutates body bytes

**What goes wrong:** Round-trip test: Open → save → diff. `word/document.xml` differs byte-for-byte even though no content was changed.

**Why it happens:** `encoding/xml` may reorder attributes, add/remove xmlns declarations, or change whitespace compared to the original file. The parsed CT_Document, when re-encoded, doesn't produce identical bytes.

**How to avoid:** The body part *will* be modified (OPC-03 covers namespace normalization). That's acceptable and expected — the round-trip contract (D-05) covers *unmodified* parts (style parts, numbering, etc.). body is always modified because we parsed it. However, the *body content semantics* must be identical. Round-trip tests should distinguish: modified parts (body) checked for semantic equality; unmodified parts (styles, theme, etc.) checked for byte identity.

**Warning signs:** Round-trip test failing only on `word/document.xml` bytes. Check whether the test correctly excludes body from the byte-identity assertion.

### Pitfall 3: FromTemplate body clearing leaves stale sectPr or residual data

**What goes wrong:** FromTemplate copies source document.xml body, then tries to "clear" it. If the clearing logic misses sectPr or section properties, the template's page setup leaks through, or remnant content remains.

**Why it happens:** Clearing body content but forgetting to reset sectPr to default values. Or using `doc.Body.P = nil` but leaving `doc.Body.Tbl` populated.

**How to avoid:** Build a fresh `CT_Body` with an empty paragraph and a minimal sectPr (page size Letter, 1-inch margins). Do not modify the cloned body at all — replace it entirely. Reuse existing `buildDocumentXML()` approach from `create.go` which produces a clean body.

**Warning signs:** FromTemplate output has template's page size or contains template boilerplate content.

### Pitfall 4: OpenTemplate body contains references to old rIds

**What goes wrong:** OpenTemplate copies source body content. That content has header/footer references (`w:headerReference`, `w:footerReference`) with rIds pointing to the source package's relationship set. But the cloned target has fresh relationships from CloneStyles (new rIds). The body's header/footer rIds are now dangling.

**Why it happens:** Body's `CT_SectPr.HdrFtrRef` entries contain rId values tied to the source part's relationship graph. CloneStyles creates new rIds for the style parts but does NOT copy header/footer parts or their relationships.

**How to avoid:** For OpenTemplate in Phase 3, the body is kept but header/footer parts are not cloned (deferred to Phase 5). Document the limitation: OpenTemplate keeps body content (paragraphs and text) but headers/footers from the template are not preserved. The sectPr from the source body should be stripped of header/footer references, or use the default sectPr from Create(). Phase 5 adds header/footer support.

**Warning signs:** Save validation fails with dangling relationship targets. `opc.Package.Save` validates all relationships before writing (OPC-06) and returns `ErrInvalidPackage` for dangling targets.

### Pitfall 5: `Save` vs `WriteTo` semantics — Close replaces WriteTo in public API

**What goes wrong:** PRD §8 lists both Save(path) and WriteTo(w io.Writer). The existing wordingo.go has Save(w io.Writer) which takes a writer. The PRD's Save takes a path. These conflict.

**Why it happens:** The existing API (`wordingo.go:34`) has `func (d *Document) Save(w io.Writer) error` which writes to a writer, not a path. SaveFile writes to a path. The Phase 3 API (D-07) lists both `Save(path string) error` and `WriteTo(w io.Writer) (int64, error)`.

**How to avoid:** Phase 3 should rename the existing `Save(w io.Writer)` to `WriteTo(w io.Writer)` (matching PRD §8) and add a new `Save(path string) error` that opens the file and calls WriteTo. This is backward-incompatible for anyone importing the module, but Phase 3 is the first public API release. Existing internal consumers (tests) update trivially.

**Warning signs:** Confusion between the existing `Save(w io.Writer)` and planned `Save(path)`. Check the PRD API sketch vs. existing implementation.

## Code Examples

### Open existing document and read paragraphs
```go
// Source: Derived from Phase 1 patterns + D-02/D-03
doc, err := Open("existing.docx")
if err != nil { return err }
defer doc.Close()

for _, p := range doc.Paragraphs() {
    fmt.Printf("Style: %s, Text: %s\n", p.Style(), p.Text())
}

// No-op save — round-trip fidelity
if err := doc.Save("existing-roundtrip.docx"); err != nil {
    return err
}
```

### FromTemplate — clone style graph, clear body
```go
// Source: D-04, using CloneStyles + buildEmptyBodyXML
func FromTemplate(path string) (*Document, error) {
    f, err := os.Open(path)
    if err != nil { return nil, err }
    defer f.Close()
    fi, err := f.Stat()
    if err != nil { return nil, err }
    return FromTemplateReader(f, fi.Size())
}

func FromTemplateReader(r io.ReaderAt, size int64) (*Document, error) {
    src, err := opc.Open(r, size)
    if err != nil { return nil, fmt.Errorf("wordingo: open template: %w", err) }

    dst := newBlankPackage()
    if err := style.CloneStyles(src, dst); err != nil {
        return nil, fmt.Errorf("wordingo: clone styles: %w", err)
    }

    emptyBody := buildEmptyBodyXML()
    dst.MarkModified("word/document.xml", emptyBody)

    doc, err := parseDocument(dst)
    if err != nil { return nil, err }
    return &Document{pkg: dst, doc: doc}, nil
}
```

### Per-part byte diff round-trip test
```go
// Source: opc_test.go:101-145 — DiffParts pattern
// Reused verbatim from Phase 1. Import as opc.DiffParts.

func TestRoundTrip_MultiHeading(t *testing.T) {
    data, err := os.ReadFile("testdata/roundtrip/multi-heading.docx")
    if err != nil { t.Fatal(err) }

    doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
    if err != nil { t.Fatal(err) }

    // Read paragraphs to exercise eager parse
    paragraphs := doc.Paragraphs()
    if len(paragraphs) == 0 {
        t.Error("expected paragraphs in fixture")
    }

    // Save and diff
    var buf bytes.Buffer
    if _, err := doc.WriteTo(&buf); err != nil {
        t.Fatal(err)
    }

    // Per-part byte diff (excludes manifests)
    opc.DiffParts(t, data, buf.Bytes())
}
```

### Lazy part loading verification test
```go
// Source: D-03 — test that supporting parts remain lazy
func TestLazyLoading(t *testing.T) {
    data, err := os.ReadFile("testdata/roundtrip/header-footer.docx")
    if err != nil { t.Fatal(err) }

    doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
    if err != nil { t.Fatal(err) }

    // Document body parsed eagerly — CT_Document non-nil
    if doc.doc == nil || doc.doc.Body == nil {
        t.Fatal("body should be eagerly parsed")
    }

    // header/footer parts should still be in raw ZIP state
    // (verified by checking that part's .data is nil and .file is non-nil)
    headerPart, ok := doc.pkg.Parts["word/header1.xml"]
    if !ok {
        t.Skip("no header part in fixture")
    }
    if headerPart.Modified() {
        t.Error("header part should not be marked modified")
    }
    // Part.Open() will defer to zip.File — no premature parse
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Document only wraps `*opc.Package` | Document also wraps `*wml.CT_Document` (eager body parse) | Phase 3 | D-03: body parsed eagerly, supporting parts lazy |
| No public Paragraph wrapper | `Paragraph` wrapping `*wml.CT_P` with read-only accessors | Phase 3 | D-02: API consumers iterate paragraphs without touching WML internals |
| Word/WriteTo combined as Save | Save(path)/WriteTo(writer) separated | Phase 3 | D-07: matches PRD §8 API sketch |
| No template creation | FromTemplate + OpenTemplate with CloneStyles | Phase 3 | D-04: template style graph cloned, body policy differs per function |

**Deprecated/outdated:**
- Existing `Document.Save(w io.Writer) error` — renamed to `WriteTo(w io.Writer) (int64, error)` in Phase 3 to match PRD §8. `Save(path string) error` becomes the file-level convenience.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | OpenTemplate body's header/footer rIds will dangle after CloneStyles (new relationship IDs allocated) | Common Pitfalls | Test failing with ErrInvalidPackage on Save. Fix: strip sectPr header/footer refs in OpenTemplate, or skip OpenTemplate header docs in Phase 3. |
| A2 | CloneStyles handles the case where source has no numbering.xml or fontTable.xml | Standard Stack | Verified in cloner_test.go (TestCloneStyles_AbsentSourcePartSkip). Low risk. |

**If this table is empty:** *(Not empty — see A1)*

## Open Questions (RESOLVED)

1. **OpenTemplate sectPr handling** [RESOLVED]
   - What we know: Source body's sectPr contains header/footer rIds that point to the source relationship graph. CloneStyles creates new rIds, so these go dangling.
   - Resolution: Strip header/footer references from sectPr in OpenTemplate for Phase 3 (D-04). Implemented via fresh defaultSectPr() in OpenTemplateReader (03-02-PLAN.md Task 1). Phase 5 adds proper header/footer cloning. Document this limitation in function doc comments.

2. **OpenTemplate naming** [RESOLVED]
   - What we know: The PRD calls it "pre-populated template support" — D-04 names it OpenTemplate.
   - Resolution: OpenTemplate and OpenTemplateReader are the final names per D-04 after user confirmation during /gsd-discuss-phase. No rename needed.

3. **Raw body bytes vs parsed CT_Document only** [RESOLVED]
   - What we know: Eager parse of document.xml on Open produces `*wml.CT_Document`. Re-encoding body changes byte content.
   - Resolution: Parse to `CT_Document` only (D-03). Raw bytes are on the part's `file` field (unmodified before MarkModified). No separate store needed. Round-trip diff compares original ZIP against re-saved ZIP — body expected to differ (re-encoded), style parts expected to match.

## Environment Availability

> Step 2.6: SKIPPED (no external dependencies identified). This phase is pure Go code changes with no external tools, services, or runtimes beyond the Go toolchain already available (go1.26.5).

## Validation Architecture

> SKIPPED: `workflow.nyquist_validation` is explicitly set to `false` in `.planning/config.json`. No validation architecture section required.

## Security Domain

> Required: `security_enforcement` is enabled in `.planning/config.json` (absent = enabled).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | Document library — no user auth |
| V3 Session Management | no | No sessions |
| V4 Access Control | no | Library exposes no access control |
| V5 Input Validation | yes | Zip bomb protection (OPC-07), path traversal rejection (OPC-07), entity expansion limits (xmlutil.NewSafeDecoder) — all inherited from Phase 1 |
| V6 Cryptography | no | No encryption, no signing |
| V12 File & Resources | yes | ZIP decompression size caps (OPC-07 MaxPartBytes/MaxTotalBytes), file path validation (OPC-07 validatePartName) |

### Known Threat Patterns

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Zip bomb (decompression amplification) | Denial of Service | OPC-07: MaxPartBytes (50MB), MaxTotalBytes (500MB), MaxCompressionRatio (200:1) — inherited from Phase 1 |
| Path traversal via part name | Tampering | OPC-07: validatePartName rejects parent traversal, absolute paths, drive letters, backslashes — inherited from Phase 1 |
| XML entity expansion (billion laughs) | Denial of Service | xmlutil.NewSafeDecoder with MaxEntityDepth (100) + MaxEntityExpansions (10,000) — inherited from Phase 1 |
| External relationship SSRF | Information Disclosure | OPC Skips external relationships on fetch; surfaced as Warnings() only — inherited from Phase 1 |

## Sources

### Primary (HIGH confidence)
- Context7: OPC package.go — confirmed MarkModified/Save/raw pass-through behavior
- Context7: wml/document.go — confirmed CT_Document/CT_Body/CT_P types
- Context7: style/cloner.go — confirmed CloneStyles interface and behavior
- Context7: opc_test.go — confirmed DiffParts round-trip test pattern
- Context7: create.go — confirmed newBlankPackage/buildDocumentXML patterns
- [CITED: .planning/PRD.md §8 — API sketch for Phase 3 function set]
- [CITED: .planning/CONTEXT.md — D-01..D-08 locked decisions]
- [CITED: .planning/REQUIREMENTS.md — STYLE-ROUNDTRIP-01..02, CREATE-03, CREATE-04]
- [CITED: .planning/ROADMAP.md — Phase 3 plan breakdown]

### Secondary (MEDIUM confidence)
- Verified existing test patterns in opc_test.go, cloner_test.go, create_test.go
- Verified Phase 2 CloneStyles atomic precondition scan (D-08) and byte pass-through (D-09)

### Tertiary (LOW confidence)
- ISO/IEC 29500 Part 1 §17.4 (Document Body), §17.6 (Sections) — standard OOXML spec references, not verified against local source

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all packages are internal, already implemented, and locally verified
- Architecture: HIGH — orchestrator pattern, CloneStyles reuse, OPC pass-through all validated by existing code
- Pitfalls: MEDIUM — A1 (OpenTemplate rId dangling) is an edge case discovered during analysis, not yet verified against code

**Research date:** 2026-07-25
**Valid until:** 2026-08-25 (stable codebase — no fast-moving dependencies)
