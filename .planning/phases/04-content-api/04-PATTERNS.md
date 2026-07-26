# Phase 4: Content API - Pattern Map

**Mapped:** 2026-07-26
**Files analyzed:** 4 (2 new, 2 modified)
**Analogs found:** 4 / 4

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `run.go` (NEW) | model | CRUD | `paragraph.go` | exact (same role + data flow) |
| `format.go` (NEW) | utility | transform | `internal/wml/properties.go` | role-match (struct types) |
| `wordingo.go` (MODIFY) | model | CRUD | `template.go` | partial (body serialization pattern) |
| `paragraph.go` (MODIFY) | model | CRUD | `paragraph.go` (existing) | exact (same file, adding methods) |

## Pattern Assignments

### `run.go` (model, CRUD) — NEW

**Analog:** `paragraph.go` (35 lines)

**Imports pattern** (paragraph.go lines 1-7):
```go
package wordingo

import (
    "strings"

    "github.com/fabiomarini/wordingo/internal/wml"
)
```

**Wrapper-over-WML-type pattern** (paragraph.go lines 11-13):
```go
// Paragraph wraps a WordprocessingML paragraph (w:p).
type Paragraph struct {
    ct *wml.CT_P
}
```
→ Run analog:
```go
// Run wraps a WordprocessingML run (w:r).
// Phase 4: formatting setters, builder chain.
type Run struct {
    ct   *wml.CT_R
    para *Paragraph   // back-reference for dirty + warning chain
}
```

**X() escape hatch pattern** (paragraph.go lines 34-35):
```go
func (p *Paragraph) X() *wml.CT_P { return p.ct }
```
→ Run analog:
```go
func (r *Run) X() *wml.CT_R { return r.ct }
```

### `format.go` (utility, transform) — NEW

**Analog:** `internal/wml/properties.go` (173 lines)

**Value-type struct pattern** (properties.go lines 17-21):
```go
type CT_Jc struct {
    XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main jc"`
    Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}
```
→ Format structs (no xml tags — pure Go API types):
```go
// RunFormat holds bulk run formatting values.
// Zero value = no override per field. Use pointers to distinguish unset vs zero.
type RunFormat struct {
    Bold      *bool
    Italic    *bool
    Underline *string
    Font      *string
    Size      *float64
    Color     *string
    Highlight *string
}

// ParFormat holds bulk paragraph formatting values.
type ParFormat struct {
    Alignment *Alignment
    Spacing   *ParSpacing
    Indent    *ParIndent
}

type ParSpacing struct {
    Before   int64  // twips
    After    int64  // twips
    Line     int64  // 240ths of a line
    LineRule string // "auto", "exact", "atLeast"
}

type ParIndent struct {
    Left      int64 // twips
    Right     int64 // twips
    FirstLine int64 // twips
    Hanging   int64 // twips
}
```

**Enum pattern** (no direct analog in properties.go — new pattern):
```go
type Alignment int

const (
    AlignmentLeft   Alignment = iota
    AlignmentCenter
    AlignmentRight
    AlignmentBoth
)

func (a Alignment) String() string {
    switch a {
    case AlignmentLeft:   return "left"
    case AlignmentCenter: return "center"
    case AlignmentRight:  return "right"
    case AlignmentBoth:   return "both"
    default:              return ""
    }
}
```

### `wordingo.go` (model, CRUD) — MODIFY

**Analog:** `template.go` (body serialization) + `wordingo.go` (current Document struct)

**Dirty flag + warning field pattern** (new fields on Document, line 46-49 existing):
```go
type Document struct {
    pkg      *opc.Package
    doc      *wml.CT_Document
    dirty    bool          // NEW: body content changed since last serialize
    warnings []string      // NEW: formatting validation warnings
}
```

**AddParagraph method pattern** (new — no direct analog, uses Paragraph constructor pattern from Paragraphs() at wordingo.go lines 105-114):
```go
// AddParagraph appends a paragraph with optional text and returns it.
func (d *Document) AddParagraph(text string) *Paragraph {
    ct := &wml.CT_P{}
    if text != "" {
        ct.R = []*wml.CT_R{{T: &wml.CT_Text{Value: text}}}
    }
    d.doc.Body.P = append(d.doc.Body.P, ct)
    d.dirty = true
    return &Paragraph{ct: ct, doc: d}
}
```

**writeBodyIfDirty/serializeBody pattern** (from template.go lines 116-126):
```go
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

**WriteTo hook** (wordingo.go lines 64-68, modified):
```go
func (d *Document) WriteTo(w io.Writer) (int64, error) {
    d.serializeBody()               // NEW: flush dirty body before save
    cw := &countWriter{w: w}
    err := d.pkg.Save(cw)
    return cw.n, err
}
```

**Warnings merge pattern** (wordingo.go lines 92-94, modified):
```go
func (d *Document) Warnings() []string {
    var all []string
    all = append(all, d.pkg.Warnings()...)
    all = append(all, d.warnings...)
    return all
}
```

**warn helper** (new on Document):
```go
func (d *Document) warn(format string, args ...any) {
    d.warnings = append(d.warnings, fmt.Sprintf(format, args...))
}
```

### `paragraph.go` (model, CRUD) — MODIFY

**Analog:** self (existing paragraph.go) + create.go (CT_P construction pattern lines 155-161)

**doc back-reference field** (paragraph.go lines 11-13, modified):
```go
type Paragraph struct {
    ct  *wml.CT_P
    doc *Document   // NEW: back-reference for dirty + warnings + resolver
}
```

**AddRun method** (new on Paragraph):
```go
func (p *Paragraph) AddRun(text string) *Run {
    t := &wml.CT_Text{Value: text}
    ct := &wml.CT_R{T: t}
    p.ct.R = append(p.ct.R, ct)
    p.doc.dirty = true
    return &Run{ct: ct, para: p}
}
```

**Nil-safe setter pattern** (new — based on D-11, avoid Pitfall 1 from RESEARCH.md):
```go
func (p *Paragraph) SetAlignment(a Alignment) *Paragraph {
    if p.ct.PPr == nil {
        p.ct.PPr = &wml.CT_PPr{}
    }
    val := a.String()
    if val == "" {
        p.doc.warn("wordingo: unknown alignment %d", a)
        return p
    }
    p.ct.PPr.Jc = &wml.CT_Jc{Val: &val}
    p.doc.dirty = true
    return p
}
```

**Spacing setter** (new):
```go
func (p *Paragraph) SetSpacing(s *ParSpacing) *Paragraph {
    if s == nil {
        p.doc.warn("wordingo: ParSpacing is nil")
        return p
    }
    if p.ct.PPr == nil {
        p.ct.PPr = &wml.CT_PPr{}
    }
    sp := &wml.CT_Spacing{}
    if s.Before != 0 { sp.Before = &s.Before }
    if s.After != 0  { sp.After = &s.After }
    if s.Line != 0   { sp.Line = &s.Line }
    if s.LineRule != "" { sp.LineRule = &s.LineRule }
    p.ct.PPr.Spacing = sp
    p.doc.dirty = true
    return p
}
```

**Indent setter** (new):
```go
func (p *Paragraph) SetIndent(i *ParIndent) *Paragraph {
    if i == nil {
        p.doc.warn("wordingo: ParIndent is nil")
        return p
    }
    if p.ct.PPr == nil {
        p.ct.PPr = &wml.CT_PPr{}
    }
    ind := &wml.CT_Ind{}
    if i.Left != 0      { ind.Left = &i.Left }
    if i.Right != 0     { ind.Right = &i.Right }
    if i.FirstLine != 0 { ind.FirstLine = &i.FirstLine }
    if i.Hanging != 0   { ind.Hanging = &i.Hanging }
    p.ct.PPr.Ind = ind
    p.doc.dirty = true
    return p
}
```

**SetStyle pattern** (new — D-05, sets pStyle reference only):
```go
func (p *Paragraph) SetStyle(name string) *Paragraph {
    if p.ct.PPr == nil {
        p.ct.PPr = &wml.CT_PPr{}
    }
    p.ct.PPr.PStyle = &wml.CT_PStyle{Val: &name}
    p.doc.dirty = true
    return p
}
```

**Bulk SetFormatting pattern** (new — calls individual setters):
```go
func (p *Paragraph) SetFormatting(f ParFormat) *Paragraph {
    if f.Alignment != nil {
        p.SetAlignment(*f.Alignment)
    }
    if f.Spacing != nil {
        p.SetSpacing(f.Spacing)
    }
    if f.Indent != nil {
        p.SetIndent(f.Indent)
    }
    return p
}
```

### Run formatting setters (in `run.go`)

**Builder chain setter pattern** (D-02, D-11):
```go
func (r *Run) SetBold(b bool) *Run {
    if r.ct.RPr == nil {
        r.ct.RPr = &wml.CT_RPr{}
    }
    r.ct.RPr.B = &wml.CT_OnOff{Val: &b}
    r.para.doc.dirty = true
    return r
}

func (r *Run) SetItalic(b bool) *Run { ... /* same pattern, sets I */ }
```

**Size conversion pattern** (RESEARCH.md Avoid Pitfall 4):
```go
func (r *Run) SetSize(pts float64) *Run {
    if pts < 0 {
        r.para.doc.warn("wordingo: negative font size %f", pts)
        return r
    }
    halfPts := int64(math.Round(pts * 2))
    if r.ct.RPr == nil {
        r.ct.RPr = &wml.CT_RPr{}
    }
    r.ct.RPr.Sz = &wml.CT_Sz{Val: &halfPts}
    r.para.doc.dirty = true
    return r
}
```

**Color validation pattern** (D-12, RESEARCH.md Pattern 4):
```go
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

**Run.SetStyle pattern** (D-08):
```go
func (r *Run) SetStyle(name string) *Run {
    if r.ct.RPr == nil {
        r.ct.RPr = &wml.CT_RPr{}
    }
    r.ct.RPr.RStyle = &wml.CT_RStyle{Val: &name}
    r.para.doc.dirty = true
    return r
}
```

**Run.SetFormatting bulk pattern** (D-01):
```go
func (r *Run) SetFormatting(f RunFormat) *Run {
    if f.Bold != nil       { r.SetBold(*f.Bold) }
    if f.Italic != nil     { r.SetItalic(*f.Italic) }
    if f.Underline != nil  { r.SetUnderline(*f.Underline) }
    if f.Font != nil       { r.SetFont(*f.Font) }
    if f.Size != nil       { r.SetSize(*f.Size) }
    if f.Color != nil      { r.SetColor(*f.Color) }
    if f.Highlight != nil  { r.SetHighlight(*f.Highlight) }
    return r
}
```

---

## Shared Patterns

### Builder Chain (pointer receiver returning `*T`)
**Source:** `paragraph.go` (existing read-only) → all new setter methods
**Apply to:** All setters on `*Run` and `*Paragraph`
```go
// Pattern: return receiver for chaining
func (r *Run) SetBold(b bool) *Run {
    // mutate ...
    return r
}
```

### Dirty-Tracking Body Serialization
**Source:** `template.go:116-126`, `wordingo.go:64-68`
**Apply to:** `Document.WriteTo`, all mutation methods
```go
// Mutation methods set dirty flag; serializeBody flushes at Save/WriteTo
func (d *Document) serializeBody() {
    if !d.dirty { return }
    // encode CT_Document → MarkModified("word/document.xml", buf)
    d.dirty = false
}
```

### Warnings Accumulation
**Source:** `internal/opc/package.go:108-113` (opc.Warnings), `internal/style/resolver.go:52-53,395-397`
**Apply to:** `Document.Warnings()` merge, all setter validation
```go
func (d *Document) warn(format string, args ...any) {
    d.warnings = append(d.warnings, fmt.Sprintf(format, args...))
}
func (d *Document) Warnings() []string {
    var all []string
    all = append(all, d.pkg.Warnings()...)
    all = append(all, d.warnings...)
    return all
}
```

### Nil-Safe Property Init
**Source:** RESEARCH.md Pattern 1 (avoid Pitfall 1)
**Apply to:** All setters accessing `ct.PPr` or `ct.RPr`
```go
if p.ct.PPr == nil {
    p.ct.PPr = &wml.CT_PPr{}
}
```

### Back-Reference Chain (Document → Paragraph → Run)
**Source:** D-06, RESEARCH.md Pattern 3
**Apply to:** `Paragraph.doc`, `Run.para` — enables dirty flag propagation
```go
type Paragraph struct {
    ct  *wml.CT_P
    doc *Document
}
type Run struct {
    ct   *wml.CT_R
    para *Paragraph
}
```

### X() Escape Hatch
**Source:** `paragraph.go:34-35`, `wordingo.go:98-100`
**Apply to:** `Run` type (new), `Paragraph` (existing)
```go
func (r *Run) X() *wml.CT_R { return r.ct }
```

---

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| *(none)* | — | — | All files have close analogs |

---

## Metadata

**Analog search scope:** `wordingo/`, `internal/wml/`, `internal/opc/`, `internal/style/`, `internal/xmlutil/`
**Files scanned:** 12
**Pattern extraction date:** 2026-07-26
