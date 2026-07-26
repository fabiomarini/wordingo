# Phase 06: Merge & Edit (v0.1.0) - Pattern Map

**Mapped:** 2026-07-26
**Files analyzed:** 10 (4 new, 4 modified, 2 unchanged refs)
**Analogs found:** 10 / 10

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `merge.go` (new) | service | transform | `list.go` (service, CRUD) | role-match |
| `body.go` (new) | model | CRUD | `paragraph.go` (model, CRUD) | role-match |
| `wordingo.go` (modify) | controller/service | CRUD | existing `wordingo.go` self | exact |
| `run.go` (modify) | model | CRUD | existing `run.go` self | exact |
| `table.go` (modify) | model | CRUD | existing `table.go` self | exact |
| `header.go` (modify) | model | CRUD | existing `header.go` self | exact |
| `merge_test.go` (new) | test | transform | `table_test.go` (test, CRUD) | role-match |
| `body_test.go` (new) | test | CRUD | `roundtrip_test.go` (test, CRUD) | role-match |
| `edit_test.go` (new) | test | CRUD | `table_test.go` (test, CRUD) | role-match |
| `uc_test.go` (new) | test | CRUD | `roundtrip_test.go` (test, CRUD) | role-match |

## Pattern Assignments

### `merge.go` (service, transform)

**Analog:** `list.go` lines 64-127 (readOrCreateNumbering + writeNumbering — part read/write pattern) + `template.go` lines 121-192 (header/footer part enumeration)

**Imports pattern** (list.go lines 3-10):
```go
import (
    "bytes"
    "fmt"

    "github.com/fabiomarini/wordingo/internal/opc"
    "github.com/fabiomarini/wordingo/internal/wml"
    "github.com/fabiomarini/wordingo/internal/xmlutil"
)
```

**Core merge walk — analog from template.go lines 126-158 (header part enumeration):**
```go
// Source: template.go lines 126-158
for _, ref := range srcSectPr.HdrFtrRef {
    if srcRels == nil { continue }
    srcRel := findRelByID(srcRels, ref.ID)
    if srcRel == nil { continue }
    target := path.Join("word", srcRel.Target)
    srcPart, ok := src.Parts[target]
    if !ok { continue }
    partBytes, err := readPartBytes(srcPart)
    if err != nil { continue }
    // ... modify part data ...
    dst.MarkModified(target, partBytes)
}
```

**Warning pattern — analog from list.go lines 90-97:**
```go
// Source: list.go lines 90-97
if err := enc.Encode(nb); err != nil {
    d.warn("wordingo: encode numbering.xml: %v", err)
    return
}
```

**Core replacePlaceholders function — RESEARCH.md code example (derived from D-04/D-05):**
- Use `strings.Index` for `{{key}}` detection (no regex per anti-patterns)
- Build joined text + runSpans with offset map
- Map char offsets back to run indices
- First fragment gets value, others deleted (set to nil)

**Header/Footer walk — template.go lines 121-192 (clone header/footer pattern):**
- Parse `sectPr.HdrFtrRef`/`FtrRef` to enumerate parts
- Read each part, decode CT_Hdr/CT_Ftr, walk `P[i].R` runs
- Call `h.sync()` after modification (per header.go L42-57)

**dirty flag — wordingo.go lines 187-209:**
```go
// Source: wordingo.go lines 191-209
if !d.dirty { return }
d.checkStyleNames()
var buf bytes.Buffer
buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
buf.WriteByte('\n')
enc := xmlutil.NewEncoder(&buf)
if err := enc.Encode(d.doc); err != nil {
    d.warn("wordingo: serialize body: %v", err)
    return
}
if err := enc.Flush(); err != nil {
    d.warn("wordingo: flush body: %v", err)
    return
}
d.pkg.MarkModified("word/document.xml", buf.Bytes())
d.dirty = false
```

---

### `body.go` (model, CRUD)

**Analog:** `paragraph.go` lines 9-40 (Paragraph wrapper pattern) + `table.go` lines 246-325 (TableBuilder wrapper pattern)

**Imports pattern** (run.go lines 1-7):
```go
package wordingo

import (
    "github.com/fabiomarini/wordingo/internal/wml"
)
```

**Wrapper-with-X pattern — paragraph.go lines 9-40:**
```go
// Source: paragraph.go lines 9-40
type Paragraph struct {
    ct  *wml.CT_P
    doc *Document
}

func (p *Paragraph) X() *wml.CT_P {
    if p == nil {
        panic("wordingo: X called on nil Paragraph")
    }
    return p.ct
}
```

**Core BodyElement pattern — RESEARCH.md code example (D-09):**
```go
// BodyElement is one item in the document body.
// Struct union (not interface) to avoid heap allocation per agent discretion.
type BodyElementType int
const (
    ElementParagraph BodyElementType = iota
    ElementTable
)

type BodyElement struct {
    Type  BodyElementType
    Para  *Paragraph
    Table *TableBuilder
}
```

**Body() accessor pattern — wordingo.go lines 220-229 (Paragraphs()):**
```go
// Source: wordingo.go lines 220-229
func (d *Document) Paragraphs() []*Paragraph {
    if d.doc == nil || d.doc.Body == nil { return nil }
    paras := make([]*Paragraph, len(d.doc.Body.P))
    for i, p := range d.doc.Body.P {
        paras[i] = &Paragraph{ct: p, doc: d}
    }
    return paras
}
```

---

### `wordingo.go` — Merge(), Body(), InsertBefore/After, DeleteParagraph (controller/service, CRUD)

**Analog:** Existing methods: AddParagraph lines 232-246, Paragraphs lines 220-229, AddList lines 195-219

**Merge() pattern — follow AddList() pattern for top-level method:**
```go
// Source: list.go lines 195-219
func (d *Document) AddList(ordered bool) *ListBuilder {
    if d == nil {
        panic("wordingo: AddList called on nil Document")
    }
    // ... operations ...
    d.dirty = true
    return &ListBuilder{...}
}
```

**MergeOpts nil-default pattern (D-03):**
```go
if opts == nil {
    opts = &MergeOpts{ScopedParts: ScopedParts{Body: true, Tables: true, Headers: true, Footers: true}}
}
```

**Unused key warning pattern — wordingo.go lines 122-124:**
```go
// Source: wordingo.go lines 122-124
func (d *Document) warn(format string, args ...any) {
    d.warnings = append(d.warnings, fmt.Sprintf(format, args...))
}
```

**InsertBefore/After by pointer — splice CT_Body.P slice:**
```go
// Find by pointer identity (p.ct == ct), then:
d.doc.Body.P = append(d.doc.Body.P[:i+1], d.doc.Body.P[i:]...)
d.doc.Body.P[i] = ct
d.dirty = true
```

**DeleteParagraph — splice CT_Body.P slice:**
```go
// Find by pointer identity, then:
d.doc.Body.P = append(d.doc.Body.P[:i], d.doc.Body.P[i+1:]...)
d.dirty = true
```

**Error/warning pattern — list.go lines 90-97:**
```go
// For non-fatal issues, warn and continue
d.warn("wordingo: merge key %q not found in document", k)
```

---

### `run.go` — SetText(), ReplaceText() (model, CRUD)

**Analog:** Existing SetBold() lines 21-31, SetColor() lines 98-112

**SetText pattern — follow SetBold pattern (lines 21-31):**
```go
// Source: run.go lines 21-31
func (r *Run) SetBold(b bool) *Run {
    if r == nil {
        panic("wordingo: SetBold called on nil Run")
    }
    if r.ct.RPr == nil { r.ct.RPr = &wml.CT_RPr{} }
    r.ct.RPr.B = &wml.CT_OnOff{Val: &b}
    r.para.doc.dirty = true
    return r
}
```

**SetText adapts to:**
```go
func (r *Run) SetText(s string) *Run {
    if r == nil {
        panic("wordingo: SetText called on nil Run")
    }
    if r.ct.T == nil {
        r.ct.T = &wml.CT_Text{}
    }
    r.ct.T.Value = s
    r.para.doc.dirty = true
    return r
}
```

**ReplaceText uses strings.ReplaceAll:**
```go
func (r *Run) ReplaceText(old, new string) *Run {
    if r == nil {
        panic("wordingo: ReplaceText called on nil Run")
    }
    if r.ct.T == nil { return r }
    r.ct.T.Value = strings.ReplaceAll(r.ct.T.Value, old, new)
    r.para.doc.dirty = true
    return r
}
```

---

### `table.go` — DeleteRow() (model, CRUD)

**Analog:** AddTable() lines 247-294 (error-returning pattern), Row() lines 94-102 (index access)

**DeleteRow error pattern — add after AddTable (lines 247-294) or Row (lines 94-102):**
```go
// Source: table.go lines 247-253, 276 (AddTable error pattern)
func (d *Document) AddTable(data [][]string) (*TableBuilder, error) {
    if d == nil { panic(...) }
    if len(data) == 0 { return nil, fmt.Errorf("wordingo: AddTable: empty data") }
    // ...
}

// DeleteRow pattern (per research — returns error, not panic):
func (tb *TableBuilder) DeleteRow(idx int) error {
    if tb == nil { panic("wordingo: DeleteRow called on nil TableBuilder") }
    if idx < 0 || idx >= len(tb.ct.Tr) {
        return fmt.Errorf("wordingo: DeleteRow: index %d out of range (rows: %d)", idx, len(tb.ct.Tr))
    }
    tb.ct.Tr = append(tb.ct.Tr[:idx], tb.ct.Tr[idx+1:]...)
    tb.doc.dirty = true
    return nil
}
```

---

### `header.go` — ParagraphContainer interface (model, CRUD)

**Analog:** Existing Header/Footer struct patterns (lines 12-57, 62-73, 75-106)

**ParagraphContainer interface — follows existing Header pattern:**
```go
// Added to header.go — ParagraphContainer interface
type ParagraphContainer interface {
    Paragraphs() []*Paragraph
    InsertParagraphAt(idx int, ct *wml.CT_P) *Paragraph
    DeleteParagraphAt(idx int)
}

// Header.Paragraphs() — follows doc.Paragraphs() pattern:
func (h *Header) Paragraphs() []*Paragraph {
    paras := make([]*Paragraph, len(h.ct.P))
    for i, p := range h.ct.P {
        paras[i] = &Paragraph{ct: p, doc: h.doc}
    }
    return paras
}

// Header.DeleteParagraphAt — splice + sync():
func (h *Header) DeleteParagraphAt(idx int) {
    h.ct.P = append(h.ct.P[:idx], h.ct.P[idx+1:]...)
    h.sync()
}
```

**Sync pattern — header.go lines 42-57:**
```go
// Source: header.go lines 42-57
func (h *Header) sync() {
    var buf bytes.Buffer
    buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
    buf.WriteByte('\n')
    enc := xmlutil.NewEncoder(&buf)
    if err := enc.Encode(h.ct); err != nil {
        h.doc.warn("wordingo: encode header: %v", err)
        return
    }
    if err := enc.Flush(); err != nil {
        h.doc.warn("wordingo: flush header: %v", err)
        return
    }
    h.doc.pkg.MarkModified(h.partName, buf.Bytes())
    h.doc.dirty = true
}
```

---

### `merge_test.go` (test, transform)

**Analog:** `table_test.go` lines 9-55 (create, operate, save, verify zip content)

**Core test pattern — table_test.go lines 9-55:**
```go
// Source: table_test.go lines 9-55
func TestTable_AddTable(t *testing.T) {
    doc, err := Create()
    if err != nil { t.Fatal(err) }

    // ... operate ...

    var buf bytes.Buffer
    if _, err := doc.WriteTo(&buf); err != nil { t.Fatal(err) }

    content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
    if !strings.Contains(content, "<w:tbl") { t.Error(...) }
}
```

**Split-run test — synthetic fixture approach from roundtrip_test.go lines 118-137:**
```go
// Use buildSyntheticFixture (roundtrip_test.go lines 118-137) to create
// hostile split-run templates with known CT_P/CT_R structure, then
// OpenReader + doc.Merge() + verify output XML
```

**Table cell merge test — create table with placeholders, verify output:**
```go
doc, err := Create()
tb, _ := doc.AddTable([][]string{{"{{name}}", "{{value}}"}})
doc.Merge(map[string]string{"name": "Alice", "value": "42"}, nil)
// verify "Alice" and "42" in output XML
```

**Warning test — wordingo.go Warnings() pattern (lines 111-118):**
```go
func TestMerge_MissingKey(t *testing.T) {
    doc, _ := Create()
    doc.AddParagraph("Hello {{name}}")
    doc.Merge(map[string]string{"name": "Alice", "missing": "never"}, nil)
    warnings := doc.Warnings()
    // verify warnings contain "merge key" and "missing"
}
```

---

### `body_test.go` (test, CRUD)

**Analog:** `roundtrip_test.go` lines 477-500 (synthetic inline tests)

**BodyElement test pattern:**
```go
func TestBody_BodyAccessor(t *testing.T) {
    doc, err := Create()
    if err != nil { t.Fatal(err) }
    doc.AddParagraph("P1")
    tb, _ := doc.AddTable([][]string{{"A1"}})

    elems := doc.Body()
    // verify len, types, order
}
```

---

### `edit_test.go` (test, CRUD)

**Analog:** `table_test.go` lines 57-125 (operation → save → verify)

**Edit test pattern — follows table_test.go:**
```go
func TestEdit_InsertBefore(t *testing.T) {
    doc, _ := Create()
    p := doc.AddParagraph("original")
    doc.InsertBefore(p, "before")
    // verify paragraph order in output
}
```

---

### `uc_test.go` (test, CRUD)

**Analog:** `roundtrip_test.go` lines 311-339 (fixture-based round-trip test)

**UC end-to-end pattern — roundtrip_test.go lines 311-339:**
```go
// Source: roundtrip_test.go lines 311-339
func TestRoundTrip_Blank(t *testing.T) {
    generateFixtures(t)
    data, err := os.ReadFile(filepath.Join(fixtureDir, "blank.docx"))
    if err != nil { t.Fatal(err) }

    doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
    if err != nil { t.Fatal(err) }

    // ... operate ...

    var buf bytes.Buffer
    if _, err := doc.WriteTo(&buf); err != nil { t.Fatal(err) }
    diffParts(t, data, buf.Bytes())
}
```

**UC1–UC4 pattern — OpenTemplate + Merge + edit + Save + verify output opens in Word:**
```go
func TestUC_TemplateMergeAndEdit(t *testing.T) {
    doc, err := OpenTemplate("testdata/template-with-placeholders.docx")
    if err != nil { t.Fatal(err) }

    doc.Merge(map[string]string{"name": "Alice", "title": "Report"})

    // Verify no unused key warnings
    // Verify output docx can be re-opened (no corruption)
}
```

---

## Shared Patterns

### Nil Receiver Panic Guard
**Source:** Every method in `paragraph.go`, `run.go`, `table.go`, `header.go`, `wordingo.go`
**Apply to:** All public methods on Document, Paragraph, Run, TableBuilder, Header, Footer
```go
if x == nil {
    panic("wordingo: MethodName called on nil ReceiverType")
}
```

### Dirty Flag + Serialization
**Source:** `wordingo.go` lines 187-209
**Apply to:** Merge engine (merge.go), InsertBefore/After/DeleteParagraph (wordingo.go), SetText/ReplaceText (run.go), DeleteRow (table.go)
```go
// Set after each mutation:
r.para.doc.dirty = true  // from run.go

// At save time, serialized via:
func (d *Document) serializeBody() {
    if !d.dirty { return }
    // encode CT_Document → bytes → MarkModified → dirty=false
}
```

### Warnings Accumulation
**Source:** `wordingo.go` lines 111-124
**Apply to:** Merge engine (merge.go) for unused keys, InsertBefore/After for paragraph-not-found
```go
func (d *Document) Warnings() []string { ... d.warnings ... }
func (d *Document) warn(format string, args ...any) {
    d.warnings = append(d.warnings, fmt.Sprintf(format, args...))
}
```

### Error-Bearing Methods vs Panic
**Source:** `table.go` lines 247-253 vs `list.go` lines 196-198
**Apply to:** DeleteRow (error return), InsertBefore/After (warn if not found)
- **Panic:** Nil receiver only
- **Error return:** Input validation (out of bounds, empty data)
- **Warn:** Non-fatal operational issues (missing key, paragraph not found)

### Wrapper with X() Escape Hatch
**Source:** All wrapper types: Paragraph, Run, TableBuilder, RowBuilder, CellBuilder, ListBuilder, Header, Footer
**Apply to:** BodyElement types (body.go), ParagraphContainer implementations (header.go)
```go
type BodyParagraph struct{ P *Paragraph }
func (bp BodyParagraph) X() *wml.CT_P { return bp.P.X() }
```

### Header/Footer Part Sync
**Source:** `header.go` lines 42-57, 75-90
**Apply to:** Merge engine when modifying header/footer paragraphs
```go
func (h *Header) sync() {
    // encode CT_Hdr → bytes → d.pkg.MarkModified(partName, buf.Bytes())
}
```

## No Analog Found

All files have close analogs. No unmapped files.

## Metadata

**Analog search scope:** root wordingo package + internal/wml
**Files scanned:** wordingo.go, paragraph.go, run.go, table.go, header.go, list.go, template.go, open.go, internal/wml/document.go, internal/wml/table.go, internal/wml/hyperlink.go, create_test.go, roundtrip_test.go, table_test.go
**Pattern extraction date:** 2026-07-26
