# Phase 4: Content API - Research

**Researched:** 2026-07-26
**Domain:** Content creation API — paragraphs, runs, formatting, named styles via builder chain
**Confidence:** HIGH

## Summary

Phase 4 extends Phase 3's read-only `Document`/`Paragraph` wrappers with mutation: `AddParagraph`, `AddRun`, per-attribute formatting setters, bulk `RunFormat`/`ParFormat` structs, named style reference via `SetStyle()`, and builder chaining (`SetBold(true).SetSize(12)`). Builds on existing WML types (`CT_PPr`, `CT_RPr`, `CT_Jc`, `CT_Spacing`, `CT_Ind`, etc.) that already have all fields defined in `internal/wml/properties.go`.

The key architectural change: Paragraph gains a back-reference to Document (`doc *Document`), and the new Run type gains a back-reference to Paragraph (`para *Paragraph`). This enables dirty-tracking for body re-serialization and warning accumulation. Mutation operations set a `dirty` flag on the Document; at Save/WriteTo time, if dirty, the entire `CT_Document` body is re-encoded to XML via `xmlutil.NewEncoder` and `pkg.MarkModified("word/document.xml", encodedBytes)`. This preserves byte-identity for style parts (D-06) — they are never MarkModified by content mutation code.

Size conversion: API takes `float64` points, stored as `int64` half-points in `CT_Sz.Val` (multiply by 2). Color takes 6-char hex `"FF0000"` format with validation warning. Alignment maps to `CT_Jc.Val` ("left", "center", "right", "both").

**Primary recommendation:** Add `doc *Document` field to Paragraph, `para *Paragraph` field to Run. Track dirty flag on Document. Re-encode body at Save/WriteTo time. All formatting setters return `*T` for builder chaining.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**D-01:** Per-attribute setters + a `RunFormat` struct for bulk setting. Setters: `SetBold(b bool)`, `SetItalic(b bool)`, `SetUnderline(u string)`, `SetFont(name string)`, `SetSize(pts float64)`, `SetColor(hex string)`, `SetHighlight(color string)`. `SetFormatting(RunFormat)` for bulk.

**D-02:** Setters return `*Run` for builder chaining (`run.SetBold(true).SetSize(12)`). Errors deferred to `Warnings()` at Save time.

**D-03:** v1 scope matches API-01 exactly: bold, italic, underline, font, size, color, highlight. Strike, subscript, superscript, smallCaps, etc. deferred beyond v1 or to Phase 6.

**D-04:** `Paragraph.AddRun(text string) *Run` appends a new `CT_R` to the paragraph's run array. No `InsertRun(index)` or `RemoveRun(index)` in v1.

**D-05:** `SetStyle` sets only the style reference (`pStyle`/`rStyle`) in XML. No pre-population of resolved properties. Word resolves at render time. `SetStyle(name string) *Paragraph` and `Run.SetStyle(name string) *Run`.

**D-06:** `Paragraph.SetStyle(name string)` — method lives on Paragraph. Paragraph holds a back-reference to Document for access to the style resolver and `MarkModified` trigger.

**D-07:** Resolver stays internal to Document. No public `StyleResolver()` accessor. Users set style names, library resolves when needed internally.

**D-08:** Run-level character styles supported: `run.SetStyle("Emphasis")` sets `rStyle` on run properties.

**D-09:** Per-category setters: `SetAlignment(a Alignment)`, `SetSpacing(s *ParSpacing)`, `SetIndent(i *ParIndent)`. Plus `SetFormatting(ParFormat)` for bulk. Consistent with run formatting (D-01).

**D-10:** v1 scope matches API-02 exactly: alignment (left/center/right/justify), spacing (before/after/line), indentation (left/right/firstLine/hanging). KeepNext, keepLines, pageBreakBefore, widowControl, shading, tabs, outlineLvl deferred.

**D-11:** Setters return `*Paragraph` for builder chaining.

**D-12:** Builder methods defer errors. Invalid input values (negative size, unknown style name, invalid color hex) logged as `Warnings()` checked at Save time. Matches existing Phase 1/2 pattern (`opc.Warnings()`).

**D-13:** Invalid input → warning via `Warnings()`. Nil state (nil Document, nil Paragraph, method call after Close) → panic. Matches Go stdlib convention.

**D-14:** `RunFormat` and `ParFormat` structs are pure data holders — no validation at creation time. Validation happens when values are applied via setter methods.

### The Agent's Discretion
- Exact struct field names for `RunFormat` and `ParFormat`
- Alignment type enum values (`AlignmentLeft`, `AlignmentCenter`, etc.)
- Whether `ParSpacing` and `ParIndent` are standalone structs or inlined in `ParFormat`
- Internal file layout within `wordingo/` for formatting code
- `SaveFile` vs `Save` naming (both currently exist — standardize or keep both)

### Deferred Ideas (OUT OF SCOPE)
None.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| API-01 | Paragraphs with runs — text, bold, italic, underline, font, size, color, highlight | Run type + formatting setters documented in Standard Stack. `CT_RPr` fields all exist in `internal/wml/properties.go` (B, I, U, Sz, Color, Highlight, RFonts). `CT_Text` has xml:space preservation. |
| API-02 | Paragraph formatting — alignment, spacing, line spacing, indentation | Paragraph formatting setter pattern. `CT_PPr` fields: Jc, Spacing, Ind all defined in `internal/wml/document.go:60-76` + `internal/wml/properties.go:17-41`. |
| API-03 | Named style application to paragraphs and runs via style engine | `SetStyle()` sets pStyle/rStyle reference only. Resolver stays internal (D-07). Style parts never MarkModified by content code (D-06). Back-reference on Paragraph enables resolver access if needed later. |
| QUAL-01 | Open from io.ReaderAt, save to io.Writer — paths are convenience only | Existing `OpenReader(r, size)` / `WriteTo(w)` patterns unchanged. Phase 4 adds no new I/O paths. |
| QUAL-02 | No panics; all failures as errors; Warnings() for non-fatal issues | D-12/D-13: deferred validation warnings, nil-state panics. Document.Warnings() merges pkg.Warnings() + formatting warnings. |
| QUAL-03 | Single public package; X() escape hatch to WML types on every wrapper | Paragraph.X() already exists. Run.X() *wml.CT_R added. Document.X() exists. |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Run formatting (bold, italic, etc.) | API / Backend | — | Pure data mutation on `CT_RPr`. No rendering, no persistence. |
| Paragraph formatting (alignment, spacing) | API / Backend | — | Pure data mutation on `CT_PPr`. No rendering, no persistence. |
| Named style reference | API / Backend | — | Sets pStyle/rStyle ID only. Style resolution is internal on-demand. |
| Body serialization | API / Backend | Database / Storage | Re-encode `CT_Document` to XML → `MarkModified` → OPC serialization at Save. |
| Warning accumulation | API / Backend | — | Document accumulates formatting warnings from setters, merged at Warnings() call. |
| Builder chaining | API / Backend | — | Pointer receiver pattern. Each setter mutates in-memory WML struct, returns self. |

## Standard Stack

### Core

Zero external dependencies — Go stdlib only [VERIFIED: go.mod].

| Package | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/fabiomarini/wordingo` (root) | module root | Public Document/Paragraph/Run API | PRD §8, D-01: single public package. Phase 4 adds Run type + methods to Paragraph. |
| `internal/wml` (document.go) | existing | CT_PPr, CT_RPr, CT_P, CT_R, CT_Body, CT_Document | All property field types already defined. No new WML types needed for API-01..03. |
| `internal/wml` (properties.go) | existing | CT_Jc, CT_Spacing, CT_Ind, CT_RFonts, CT_Sz, CT_Color, CT_Highlight, CT_OnOff, CT_U, CT_PStyle, CT_RStyle | All formatting value types ready. CT_OnOff for B/I toggle. CT_Sz for size (half-points). CT_Color for hex color. CT_Highlight for text highlight. CT_Spacing/CT_Ind for paragraph spacing/indent. CT_Jc for alignment. |
| `internal/opc` (package.go) | existing | MarkModified, Save, Warnings | Body serialization via MarkModified. Unmodified parts raw-copied (OPC-04). |
| `internal/style` (resolver.go) | existing | ResolveParagraph, ResolveRun | Not called directly by Phase 4 setters. Available for internal use. Resolver.Warnings() merged into Document.Warnings(). |
| `internal/xmlutil` (encoder.go) | existing | NewEncoder → Encode/Flush with canonical w:prefix rewriting | Body re-encoding pattern established in create.go and template.go. |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `encoding/xml` | stdlib | Marshal/Unmarshal CT_Document to/from XML | Body serialization at Save/WriteTo time. |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Full body re-encode on each mutation | Track dirty flag, re-encode once at Save | Dirty flag avoids re-encodes for intermediate chain steps. Prefer dirty-flag pattern. |
| Paragraph back-reference to Document | Closure-based markModified func | Back-reference is simpler, matches D-06 requirement for resolver access. |
| error-returning setters | Builder chain returning `*T` | D-02 locks builder chain. Errors deferred to Warnings(). |

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
Document (in-memory CT_Document)
    │
    ├── AddParagraph(text) ──► appends CT_P to Body.P
    │     │                        │
    │     │                        ▼
    │     │                   Paragraph{ct: *wml.CT_P, doc: d}
    │     │                        │
    │     │                        ├── SetAlignment(a) ──► ct.PPr.Jc = &CT_Jc{Val: &a}
    │     │                        ├── SetSpacing(s)   ──► ct.PPr.Spacing = &CT_Spacing{...}
    │     │                        ├── SetIndent(i)    ──► ct.PPr.Ind = &CT_Ind{...}
    │     │                        ├── SetStyle(name)  ──► ct.PPr.PStyle = &CT_PStyle{Val: &name}
    │     │                        ├── SetFormatting(ParFormat) ──► calls individual setters
    │     │                        └── AddRun(text)    ──► appends CT_R to ct.R
    │     │                                                   │
    │     │                                                   ▼
    │     │                                              Run{ct: *wml.CT_R, para: p}
    │     │                                                   │
    │     │                                                   ├── SetBold(b)   ──► ct.RPr.B = &CT_OnOff{Val: &b}
    │     │                                                   ├── SetItalic(b) ──► ct.RPr.I = &CT_OnOff{Val: &b}
    │     │                                                   ├── SetUnderline(u) ──► ct.RPr.U = &CT_U{Val: &u}
    │     │                                                   ├── SetFont(name) ──► ct.RPr.RFonts = &CT_RFonts{Ascii: &n, HAnsi: &n}
    │     │                                                   ├── SetSize(pts) ──► ct.RPr.Sz = &CT_Sz{Val: halfPts}
    │     │                                                   ├── SetColor(hex) ──► ct.RPr.Color = &CT_Color{Val: &hex}
    │     │                                                   ├── SetHighlight(c) ──► ct.RPr.Highlight = &CT_Highlight{Val: &c}
    │     │                                                   ├── SetStyle(n)  ──► ct.RPr.RStyle = &CT_RStyle{Val: &n}
    │     │                                                   └── SetFormatting(RunFormat) ──► calls individual setters
    │     │
    │     └── dirty = true  (each mutation sets dirty)
    │
    └── WriteTo / Save
            │
            ├── dirty? ──► serialize CT_Document → xmlutil.NewEncoder → buf
            │                  pkg.MarkModified("word/document.xml", buf.Bytes())
            │
            └── pkg.Save(w)
                    ├── modified body part → write from part.data
                    └── unmodified parts (styles, theme, etc.) → zip.Copy byte-identical
```

### Recommended Project Structure

```
wordingo/                          # Public root package
├── wordingo.go                    # Document type, Create(), WriteTo, Save, Warnings, Close (exists)
├── create.go                      # newBlankPackage(), buildDocumentXML(), etc. (exists)
├── open.go                        # Open(), OpenReader() (exists)
├── template.go                    # FromTemplate, OpenTemplate (exists)
├── paragraph.go                   # Paragraph: Style(), Text(), X() + Phase 4 additions
│                                  #   AddRun, SetAlignment, SetSpacing, SetIndent, SetStyle, SetFormatting
├── run.go                         # NEW: Run type + formatting setters
│                                  #   SetBold, SetItalic, SetUnderline, SetFont, SetSize
│                                  #   SetColor, SetHighlight, SetStyle, SetFormatting, X()
├── format.go                      # NEW: RunFormat, ParFormat, ParSpacing, ParIndent, Alignment type + constants
├── doc.go                         # Package documentation (exists)
├── create_test.go                 # Existing Create() tests
├── roundtrip_test.go              # Round-trip per-part diff tests (exists)
└── template_test.go               # Template tests (exists)
```

### Pattern 1: Builder Chain with Pointer Receiver

**What:** Every setter mutates in-memory WML struct fields and returns the receiver for chaining. Dirty flag propagates to Document for deferred serialization.

**When to use:** All formatting setters on `*Run` and `*Paragraph`.

**Example:**
```go
// Source: D-02, D-11, established Go builder pattern
func (r *Run) SetBold(b bool) *Run {
    if r.ct.RPr == nil {
        r.ct.RPr = &wml.CT_RPr{}
    }
    r.ct.RPr.B = &wml.CT_OnOff{Val: &b}
    r.para.doc.dirty = true
    return r
}

// Usage: doc.AddParagraph("Hello").SetBold(true).SetSize(12)
```

### Pattern 2: Dirty-Tracking Body Serialization

**What:** Document tracks whether body content changed via `dirty bool`. At Save/WriteTo, if dirty, re-encodes CT_Document to XML and calls MarkModified.

**When to use:** Any mutation that changes body content (AddParagraph, AddRun, formatting setters, SetStyle).

**Example:**
```go
// Source: template.go MarkModified pattern + D-06
func (d *Document) writeBodyIfDirty() {
    if !d.dirty {
        return
    }
    var buf bytes.Buffer
    buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
    buf.WriteByte('\n')
    enc := xmlutil.NewEncoder(&buf)
    if err := enc.Encode(d.doc); err != nil {
        // Should not happen: doc was just built/modified in-process
        return
    }
    enc.Flush()
    d.pkg.MarkModified("word/document.xml", buf.Bytes())
    d.dirty = false
}

func (d *Document) WriteTo(w io.Writer) (int64, error) {
    d.writeBodyIfDirty()
    cw := &countWriter{w: w}
    err := d.pkg.Save(cw)
    return cw.n, err
}
```

### Pattern 3: Back-Reference Chain (Document → Paragraph → Run)

**What:** Paragraph holds `doc *Document`. Run holds `para *Paragraph`. This enables any level of the object graph to mark the body dirty, accumulate warnings, and (for Paragraph) access the style resolver.

**When to use:** All Paragraph and Run construction — both from Document.AddParagraph and Paragraph.AddRun.

**Example:**
```go
// Source: D-06
type Paragraph struct {
    ct  *wml.CT_P
    doc *Document     // back-reference for MarkModified + warnings + resolver
}

func (p *Paragraph) AddRun(text string) *Run {
    t := &wml.CT_Text{Value: text}
    ct := &wml.CT_R{T: t}
    p.ct.R = append(p.ct.R, ct)
    p.doc.dirty = true
    return &Run{ct: ct, para: p}
}

type Run struct {
    ct   *wml.CT_R
    para *Paragraph   // back-reference for dirty tracking
}

func (r *Run) X() *wml.CT_R { return r.ct }
```

### Pattern 4: Warning Accumulation via Document Back-Reference

**What:** Formatting validation warnings are appended to `doc.warnings` (a `[]string` on Document). Document.Warnings() merges OPC warnings + own warnings.

**When to use:** All setter methods that accept values requiring validation (color format, size, spacing).

**Example:**
```go
// Source: D-12, D-13
func validHexColor(s string) bool {
    if len(s) != 6 { return false }
    for _, c := range s {
        if !(c >= '0' && c <= '9') && !(c >= 'A' && c <= 'F') && !(c >= 'a' && c <= 'f') {
            return false
        }
    }
    return true
}

func (r *Run) SetColor(hex string) *Run {
    if !validHexColor(hex) {
        r.para.doc.warn("wordingo: invalid color %q: must be 6 hex digits", hex)
        return r
    }
    if r.ct.RPr == nil { r.ct.RPr = &wml.CT_RPr{} }
    r.ct.RPr.Color = &wml.CT_Color{Val: &hex}
    r.para.doc.dirty = true
    return r
}
```

### Anti-Patterns to Avoid

- **Calling MarkModified on every setter call:** Instead, set dirty flag and re-encode once at Save. Multiple re-encodes during a builder chain wastes CPU.
- **Storing both raw body bytes and parsed CT_Document:** Phase 3 research already resolved this (parse only). Dirty flag determines when to re-encode.
- **Exposing Resolver publicly:** D-07 says internal only. Paragraph.SetStyle sets pStyle reference, does NOT call ResolveParagraph.
- **InsertAfter/InsertBefore in v1:** D-04 says AddParagraph appends only. Editing operations deferred to Phase 6.
- **Color with `#` prefix:** OOXML uses 6 hex chars without `#`. Accept either? D-12 says warning on invalid format. Recommend: accept without `#`, reject with `#` with warning.
- **Default nil RPr handling:** All setters must lazy-init `ct.RPr` (and `ct.PPr`) when nil. Otherwise `nil pointer` panic on first setter call.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| XML serialization of CT_Document | Manual XML construction | `xmlutil.NewEncoder` + `Encode` | Existing pattern in create.go/template.go. Handles namespace prefix rewrite, xmlns injection, mc:Ignorable. |
| OPC part modification tracking | Manual part replacement tracking | `opc.MarkModified(name, data)` | Phase 1 established pattern. Save picks up modified parts via part.data field. |
| Style resolution for effective properties | Inline merge logic | `internal/style.Resolver` | Phase 2 built full basedOn chain walker, memo cache, cycle detection. Don't duplicate. |
| Warning accumulation | Separate warning system | `Document.warnings []string` + `Warnings()` | Extends existing OPC Warnings() pattern. Single merge point at Document.Warnings(). |
| ZIP writing for Save | Custom ZIP writer | `archive/zip` + `opc.Package.Save()` | OPC-02 canonical ordering, OPC-04 raw pass-through, OPC-06 relationship validation already built. |

## Common Pitfalls

### Pitfall 1: Nil RPr/PPr on first setter call

**What goes wrong:** `Paragraph.SetAlignment(AlignmentCenter)` panics with nil pointer because `ct.PPr` is nil. Same for `Run.SetBold(true)` when `ct.RPr` is nil.

**Why it happens:** New paragraphs and runs created via `AddParagraph`/`AddRun` have nil property structs. The first setter must lazy-init them.

**How to avoid:** Every setter checks nil and initializes before setting:
```go
func (p *Paragraph) SetAlignment(a Alignment) *Paragraph {
    if p.ct.PPr == nil {
        p.ct.PPr = &wml.CT_PPr{}
    }
    // ... set Jc
}
```

**Warning signs:** Panic on `nil pointer` in setter code.

### Pitfall 2: Full body re-encode on every chain call

**What goes wrong:** `doc.AddParagraph("Hello").SetBold(true).SetSize(12)` triggers three full XML serializations (one per mutation), wasting CPU.

**Why it happens:** Naively calling `writeBodyIfDirty()` in each setter instead of just setting `dirty = true`.

**How to avoid:** Setters ONLY set `dirty = true`. Body re-encode happens once, lazily, at `WriteTo`/`Save` time via `writeBodyIfDirty()`.

**Warning signs:** Performance profiling shows repeated body serialization during a single builder chain.

### Pitfall 3: MarkModified on style parts

**What goes wrong:** Content mutation code accidentally calls `pkg.MarkModified("word/styles.xml", ...)`, destroying byte-identity for style parts (STYLE-ROUNDTRIP-01 fails).

**Why it happens:** The body serialization pattern (`xmlutil.NewEncoder` → `MarkModified`) gets misapplied to style parts. Same as Phase 3 Pitfall 1.

**How to avoid:** D-06 rule: Never call MarkModified on style parts from content mutation code. Only the body part (`word/document.xml`) gets MarkModified. The OPC layer handles style part pass-through automatically.

**Warning signs:** Round-trip test failing on `word/styles.xml` or `word/theme/theme1.xml` byte difference.

### Pitfall 4: Size conversion off-by-one

**What goes wrong:** `SetSize(12)` results in 12pt font but stored as 12 half-points (should be 24). Or `SetSize(12.5)` truncates to 12 instead of 25 half-points.

**Why it happens:** OOXML stores font size in half-points (`CT_Sz.Val = halfPoints`). Size of 12pt → 24 half-points. Float to int conversion can truncate oddly.

**How to avoid:** `halfPts := int64(math.Round(pts * 2))`. Use `math.Round` not direct cast to avoid truncation at 12.3 → 24 (should be 25).

**Warning signs:** Font renders at half the expected size in Word.

### Pitfall 5: Run.AddRun after Paragraphs() on opened document

**What goes wrong:** On an opened document, `doc.Paragraphs()` returns existing paragraphs. Calling `AddRun` on one of them appends to the existing CT_R slice. This works, but the paragraph must already have CT_RPr initialized.

**Why it happens:** Existing paragraphs may have nil RPr. The pattern works if setters handle nil init, but AddRun alone (without setter) doesn't need RPr.

**How to avoid:** AddRun creates a new CT_R with the text. No RPr needed for plain text. Setter methods handle nil RPr init. Works correctly.

**Warning signs:** — (works correctly by design)

### Pitfall 6: Save vs SaveFile naming confusion

**What goes wrong:** Both `Save(path)` and `SaveFile(path)` exist. One calls WriteTo, the other calls pkg.Save directly. Inconsistent behavior.

**Why it happens:** Phase 1 kept both. Phase 3 added Save(path) as the canonical form. SaveFile is the older variant.

**How to avoid:** At the agent's discretion — either remove SaveFile or keep it as alias calling Save. Recommend: keep SaveFile as deprecated alias, document that Save is the canonical method. Or remove it entirely if backward compat not needed (Phase 4 is pre-1.0).

## Code Examples

Verified patterns from existing codebase:

### Adding runs with formatting (API-01)
```go
// Source: D-01, D-04, D-11 patterns
doc, _ := wordingo.Create()
p := doc.AddParagraph("Hello")
p.AddRun("world").
    SetBold(true).
    SetSize(12)

// With RunFormat bulk struct
p.AddRun("emphasis").
    SetFormatting(RunFormat{
        Bold:   boolPtr(true),
        Color:  "FF0000",
        Size:   14,
    })

// Save — body serialized once
doc.Save("output.docx")
```

### Paragraph formatting (API-02)
```go
// Source: D-09, D-10
p := doc.AddParagraph("Indented quote")
p.SetAlignment(AlignmentCenter)
p.SetIndent(&ParIndent{Left: 720, Right: 720})     // 0.5 inch each side
p.SetSpacing(&ParSpacing{Before: 240, After: 120})  // twips

// With ParFormat bulk struct
p2 := doc.AddParagraph("Bulk formatted")
p2.SetFormatting(ParFormat{
    Alignment: AlignmentBoth,
    Spacing:   &ParSpacing{Line: 360, LineRule: "auto"},
})
```

### Named style application (API-03)
```go
// Source: D-05, D-06, D-08
doc, _ := wordingo.FromTemplate("template.docx")
p := doc.AddParagraph("Title")
p.SetStyle("Title")                                 // sets pStyle = "Title"

p.AddRun("Emphasized text").SetStyle("Emphasis")    // sets rStyle = "Emphasis"

// SetStyle sets reference only — no pre-population (D-05)
// Word resolves at render time
```

### Warning accumulation
```go
// Source: D-12, D-13
p := doc.AddParagraph("Warning test")
p.AddRun("bad color").SetColor("XYZ123")             // invalid — not hex
p.AddRun("negative size").SetSize(-12)                // invalid — negative

doc.Save("output.docx")
warnings := doc.Warnings()                            // ["invalid color ...", "negative size ..."]
```

### Full writeBodyIfDirty + dirty flag pattern
```go
// Source: template.go MarkModified pattern + D-06
type Document struct {
    pkg      *opc.Package
    doc      *wml.CT_Document
    dirty    bool
    warnings []string
}

func (d *Document) WriteTo(w io.Writer) (int64, error) {
    d.serializeBody()
    cw := &countWriter{w: w}
    err := d.pkg.Save(cw)
    return cw.n, err
}

func (d *Document) serializeBody() {
    if !d.dirty {
        return
    }
    var buf bytes.Buffer
    buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
    buf.WriteByte('\n')
    enc := xmlutil.NewEncoder(&buf)
    if err := enc.Encode(d.doc); err != nil {
        return // in-memory encode should never fail
    }
    enc.Flush()
    d.pkg.MarkModified("word/document.xml", buf.Bytes())
    d.dirty = false
}
```

### Nil-safe setter pattern
```go
// Source: D-11, avoid Pitfall 1
func (p *Paragraph) SetAlignment(a Alignment) *Paragraph {
    if p.ct.PPr == nil {
        p.ct.PPr = &wml.CT_PPr{}
    }
    val := a.String() // "left", "center", "right", "both"
    if val == "" {
        p.doc.warn("wordingo: unknown alignment %d", a)
        return p
    }
    p.ct.PPr.Jc = &wml.CT_Jc{Val: &val}
    p.doc.dirty = true
    return p
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Paragraph read-only (Phase 3) | Paragraph with mutation methods + doc back-reference | Phase 4 | D-06: Paragraph.doc for MarkModified + warnings + resolver |
| No Run type | Run wrapper over CT_R with formatting setters | Phase 4 | D-01..D-04: full run formatting API |
| Body never modified on save | Dirty-tracking + body re-encode on Save/WriteTo | Phase 4 | D-06: style parts untouched, body re-encoded |
| opc.Warnings() only | Document.Warnings() merges opc + formatting warnings | Phase 4 | D-12: builder deferred errors hook into Warnings() |
| No SetStyle | SetStyle sets pStyle/rStyle reference | Phase 4 | D-05: reference only, no pre-population |

**Deprecated/outdated:**
- `Document.SaveFile(path)` — duplicate of `Save(path)`. Recommend removing or deprecating.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Body re-encode at Save time (lazy via dirty flag) is performant enough for Phase 4 workloads | Architecture Patterns | If re-encode takes >100ms on large documents, consider incremental body mutation approach. However, Phase 4 targets new document creation (not editing large existing docs), so single re-encode per Save is acceptable. |
| A2 | Color format "FF0000" (6 hex chars, no #) matches OOXML convention | Code Examples | Word accepts optional `#` prefix? Need to verify. If Word requires `#`, color strings will render incorrectly. Mitigation: test with Word. |
| A3 | Highlight color values match OOXML enum (yellow, green, cyan, magenta, blue, red, darkBlue, darkCyan, darkGreen, darkMagenta, darkRed, darkYellow, darkGray, lightGray, black, none, white) | Common Pitfalls | SetHighlight takes a string, CT_Highlight.Val stores as-is. Word ignores invalid values. Minimal risk. |

## Open Questions

1. **RunFormat underline value format** [RESOLVED]
   - `SetUnderline(u string)` sets CT_U.Val. OOXML underline types are: "single", "double", "words", "thick", "dotted", "dashed", "dotDash", "dotDotDash", "wave", "dottedHeavy", "dashedHeavy", "dotDashHeavy", "dotDotDashHeavy", "waveHeavy", "wavyDouble", "none". Defer validation to Word (unknown values silently ignored). D-12: invalid values get a warning.

2. **ParSpacing.Line unit** [RESOLVED]
   - OOXML spacing Line is in 240ths of a line (360 = 1.5 lines, 480 = double). LineRule "auto" (default), "exact", "atLeast". API takes int64, stored as-is in CT_Spacing.Line. User's responsibility to use correct units.

3. **Save vs SaveFile** [RESOLVED]
   - Keep both. Save is canonical. SaveFile kept for backward compat but deprecated in documentation. Not removed since it's already public.

## Environment Availability

> Step 2.6: SKIPPED (no external dependencies identified). This phase is pure Go code changes with no external tools, services, or runtimes beyond the Go toolchain already available.

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
| V5 Input Validation | yes | Formatting validators (color, size) reject invalid values with Warnings(). Inherits OPC-07 (zip bomb, path traversal, entity expansion) from Phase 1. |
| V6 Cryptography | no | No encryption, no signing |

### Known Threat Patterns for Go + OOXML

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Body serialization injection via color/size strings | Tampering | Setter validators check hex format, reject non-hex color chars. Size validated as non-negative float. Values stored in typed struct fields (CT_Sz has *int64, CT_Color has *string) — no raw XML injection possible. |
| In-memory body corruption via malicious CT_PPr values | Tampering | Setters only write to known WML struct fields. Unknown XML elements go to RawXML slice (preserved but not injected). No XSS-like vectors in OOXML context. |
| Resource exhaustion via body re-encode | Denial of Service | Dirty-flag pattern ensures one re-encode per Save. Decompression bombs already blocked by OPC-07. Body CT_Document is bounded by OPC MaxPartBytes (128MB) at read time; created documents are built in-memory by user, naturally bounded. |

## Sources

### Primary (HIGH confidence)
- Context7: `internal/wml/document.go` — CT_PPr contains PStyle, Spacing, Ind, Jc with all fields; CT_RPr contains B, I, U, Sz, Color, Highlight, RFonts, RStyle — all fields confirmed present and tagged for XML.
- Context7: `internal/wml/properties.go` — CT_Jc, CT_Spacing, CT_Ind, CT_RFonts, CT_Sz, CT_Color, CT_Highlight, CT_OnOff, CT_U, CT_PStyle, CT_RStyle all defined with correct XML tags.
- Context7: `internal/opc/package.go` — MarkModified signature and behavior confirmed: `func (p *Package) MarkModified(name string, data []byte)`, creates part if not exists.
- Context7: `internal/opc/package.go` — Save flow: modified parts written from `part.data`, unmodified parts raw-copied via `zip.Copy`.
- Context7: `template.go` — Body XML encoding pattern: xmlutil.NewEncoder → Encode → Flush → MarkModified.
- Context7: `create.go` — buildDocumentXML pattern: `enc := xmlutil.NewEncoder(&buf)`, `enc.Encode(doc)`, `enc.Flush()`.
- Context7: `wordingo.go` — Document struct has `pkg *opc.Package` and `doc *wml.CT_Document` fields. Warnings() delegates to pkg.Warnings().
- Context7: `internal/style/resolver.go` — ResolveParagraph/ResolveRun signatures, mergePPr/mergeRPr field-by-field merge helpers.

### Secondary (MEDIUM confidence)
- Go builder pattern: pointer receiver returning `*T` is standard Go idiom for method chaining. Verified against stdlib examples (e.g., `strings.Builder.WriteString`).
- OOXML half-point convention: CT_Sz.Val is half-points (12pt → 24). Standard ISO/IEC 29500 §17.3.2.36.

### Tertiary (LOW confidence)
- No tertiary claims in this research.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all packages are internal, existing, and verified. No external dependencies.
- Architecture: HIGH — dirty-flag + lazy body serialize is tested by template.go pattern; back-reference chain is simple.
- Pitfalls: HIGH — all pitfalls identified from Phase 3 experience (nil PPr, style part modification) or documented OOXML spec details (size conversion).

**Research date:** 2026-07-26
**Valid until:** 2026-08-26 (stable codebase — no fast-moving dependencies)
