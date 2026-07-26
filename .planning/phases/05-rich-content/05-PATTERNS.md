# Phase 5: Rich Content - Pattern Map

**Mapped:** 2026-07-26
**Files analyzed:** 17 (8 new, 6 modified, 3 WML gap-fill)
**Analogs found:** 17 / 17

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `wordingo/table.go` | public builder | builder-chain | `wordingo/paragraph.go` | exact |
| `wordingo/image.go` | public API + OPC | file-I/O + builder | `wordingo/paragraph.go` | role-match |
| `wordingo/header.go` | public API + OPC | CRUD + OPC part | `wordingo/create.go` (part creation) | role-match |
| `wordingo/list.go` | public builder | builder-chain + CRUD | `wordingo/paragraph.go` | role-match |
| `wordingo/page.go` | public API | request-response (mutate sectPr) | `wordingo/paragraph.go` (SetAlignment) | role-match |
| `wordingo/hyperlink.go` | public API | request-response (rel + element) | `wordingo/run.go` (Run type) | exact |
| `internal/wml/drawing.go` | WML schema type | serialization | `internal/wml/table.go` | exact |
| `internal/wml/hyperlink.go` | WML schema type | serialization | `internal/wml/table.go` | exact |
| `wordingo.go` (modify) | public API | CRUD | existing methods (AddParagraph) | exact |
| `wordingo/paragraph.go` (modify) | public API | request-response | existing methods (AddRun) | exact |
| `internal/wml/document.go` (modify) | WML schema | serialization | existing CT_P/CT_R fields | exact |
| `internal/wml/numbering.go` (modify) | WML schema | serialization | existing (likely no change) | exact |
| `template.go` (modify) | OPC transform | file-I/O + part cloning | existing OpenTemplateReader | exact |
| `create.go` (modify) | config | static constants | existing ctMain/relStyles | exact |

## Pattern Assignments

### `wordingo/table.go` — TableBuilder (public builder, builder-chain)

**Analog:** `wordingo/paragraph.go` lines 9-168 — Paragraph builder chaining

**Imports pattern** (paragraph.go:1-7):
```go
package wordingo

import (
    "strings"
    "github.com/fabiomarini/wordingo/internal/wml"
)
```

**Core builder pattern** (paragraph.go:9-13, 42-52):
```go
// TableBuilder wraps a CT_Tbl with builder chaining.
type TableBuilder struct {
    ct  *wml.CT_Tbl
    doc *Document
}

// Row returns a row builder for the given index.
// Appends new rows as needed to match idx.
func (t *TableBuilder) Row(idx int) *RowBuilder {
    // grow t.ct.Tr slice to include idx
    // ...
    return &RowBuilder{ct: t.ct.Tr[idx], doc: t.doc}
}
```

**SetStyle pattern** (paragraph.go:133-151):
```go
func (t *TableBuilder) SetTableStyle(name string) *TableBuilder {
    if t == nil {
        panic("wordingo: SetTableStyle called on nil TableBuilder")
    }
    if t.ct.TblPr == nil {
        t.ct.TblPr = &wml.CT_TblPr{}
    }
    t.ct.TblPr.TblStyle = &wml.CT_TblStyle{Val: &name}
    t.doc.dirty = true
    return t
}
```

**SetWidth pattern** (paragraph.go: setter with value + type):
```go
func (t *TableBuilder) SetWidth(w int64, wType string) *TableBuilder {
    if t == nil {
        panic("wordingo: SetWidth called on nil TableBuilder")
    }
    if t.ct.TblPr == nil {
        t.ct.TblPr = &wml.CT_TblPr{}
    }
    t.ct.TblPr.TblW = &wml.CT_TblW{W: &w, Type: &wType}
    t.doc.dirty = true
    return t
}
```

**Cell merge pattern** (uses existing CT_GridSpan, CT_VMerge in `internal/wml/table.go:110-120`):
```go
func (c *CellBuilder) MergeRight() *CellBuilder {
    if c.ct.TcPr == nil {
        c.ct.TcPr = &wml.CT_TcPr{}
    }
    span := int64(2) // or increment existing
    c.ct.TcPr.GridSpan = &wml.CT_GridSpan{Val: &span}
    c.doc.dirty = true
    return c
}
```

**X() escape hatch** (paragraph.go:34-40):
```go
func (t *TableBuilder) X() *wml.CT_Tbl {
    if t == nil {
        panic("wordingo: X called on nil TableBuilder")
    }
    return t.ct
}
```

---

### `wordingo/image.go` — Image embedding (public API + OPC part creation, file-I/O + builder)

**Analog:** `wordingo/paragraph.go` (builder return) + `wordingo/create.go:48-97` (OPC pattern)

**OPC part creation pattern** (create.go:118-126 + relationships.go:138-145):
```go
func (d *Document) AddImageBytes(name string, data []byte, contentType string) (*Run, error) {
    mediaPath := "word/media/" + name

    // 1. Create media part
    d.pkg.MarkModified(mediaPath, data)

    // 2. Add relationship from document.xml.rels
    rels := d.pkg.Rels["word/document.xml"]
    rId := rels.NextRID()
    rels.Rels = append(rels.Rels, opc.Relationship{
        ID:     rId,
        Type:   relImage,
        Target: "media/" + name,
    })

    // 3. Add content type override
    d.pkg.ContentTypes.Overrides["/"+mediaPath] = contentType

    // 4. Create run with DrawingML
    run := &Run{ct: &wml.CT_R{Drawing: drawing}, para: para}
    return run, nil
}
```

**EMU conversion constant pattern** (RESEARCH.md:157-160):
```go
const emusPerInch = 914400

func inchesToEMU(inches float64) int64 {
    return int64(inches * emusPerInch)
}
```

**DPI detection pattern** (RESEARCH.md:830-835 — stdlib image.DecodeConfig + manual APP0/pHYs):
```go
// Use image/jpeg.DecodeConfig / image/png.DecodeConfig for pixel dimensions
// Parse JPEG APP0 (JFIF) marker for DPI or PNG pHYs chunk
// Fallback: 72 DPI
```

---

### `wordingo/header.go` — Header/Footer API (public API + OPC part, CRUD + part creation)

**Analog:** `wordingo/create.go:48-97` (package assembly), `wordingo/paragraph.go` (Paragraph wrapper pattern)

**Part creation + sectPr link pattern** (create.go:152-195 + document.go:170-181 for CT_SectPr):
```go
type HeaderVariant int

const (
    HeaderDefault HeaderVariant = iota
    HeaderFirst
    HeaderEven
)

func (hv HeaderVariant) String() string {
    switch hv {
    case HeaderFirst:
        return "first"
    case HeaderEven:
        return "even"
    default:
        return "default"
    }
}

func (d *Document) AddHeader(variant HeaderVariant) *Header {
    partName := fmt.Sprintf("word/header%d.xml", d.nextHeaderID)
    d.nextHeaderID++

    hdr := &wml.CT_Hdr{P: []*wml.CT_P{{
        PPr: &wml.CT_PPr{PStyle: &wml.CT_PStyle{Val: strPtr("Normal")}},
    }}} // empty paragraph placeholder

    // Marshal
    var buf bytes.Buffer
    buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
    buf.WriteByte('\n')
    enc := xmlutil.NewEncoder(&buf)
    enc.Encode(hdr)
    enc.Flush()

    // Create part
    d.pkg.MarkModified(partName, buf.Bytes())

    // Add relationship (target relative to word/)
    rels := d.pkg.Rels["word/document.xml"]
    rId := rels.NextRID()
    rels.Rels = append(rels.Rels, opc.Relationship{
        ID: rId, Type: relHeader,
        Target: "header" + strconv.Itoa(d.nextHeaderID-1) + ".xml",
    })

    // Add content type
    d.pkg.ContentTypes.Overrides["/"+partName] = ctHeader

    // Link via sectPr
    ref := &wml.CT_HdrFtrRef{ID: rId, Type: variant.String()}
    d.doc.Body.SectPr.HdrFtrRef = append(d.doc.Body.SectPr.HdrFtrRef, ref)
    d.dirty = true

    return &Header{ct: hdr, doc: d}
}
```

**Header/Footer wrapper** (mirrors paragraph.go:9-13):
```go
type Header struct {
    ct  *wml.CT_Hdr
    doc *Document
}

func (h *Header) AddParagraph(text string) *Paragraph {
    ct := &wml.CT_P{}
    if text != "" {
        ct.R = []*wml.CT_R{{T: &wml.CT_Text{Value: text}}}
    }
    h.ct.P = append(h.ct.P, ct)
    h.doc.dirty = true
    return &Paragraph{ct: ct, doc: h.doc}
}
```

---

### `wordingo/list.go` — List builder (public builder, builder-chain + numbering CRUD)

**Analog:** `wordingo/paragraph.go` (builder chain) + `internal/style/numbering.go:53-81` (read numbering.xml)

**Numbering def auto-generation pattern** (internal/style/numbering.go:53-81 — existing read pattern):
```go
func (d *Document) AddList(ordered bool) *ListBuilder {
    // 1. Parse existing numbering.xml or create empty
    nb := d.readOrCreateNumbering()

    // 2. Find max existing abstractNumId
    nextAbsID := int64(0)
    for _, a := range nb.AbstractNum {
        if a.AbstractNumID != nil && *a.AbstractNumID >= nextAbsID {
            nextAbsID = *a.AbstractNumID + 1
        }
    }

    // 3. Create abstractNum with 9 levels
    abs := &wml.CT_AbstractNum{
        AbstractNumID: &nextAbsID,
        Lvl:   makeLevels(ordered), // 9 levels with numFmt + lvlText
    }
    nb.AbstractNum = append(nb.AbstractNum, abs)

    // 4. Create num entry
    nextNumID := findMaxNumID(nb) + 1
    nb.Num = append(nb.Num, &wml.CT_Num{
        NumID: &nextNumID,
        AbstractNumID: &wml.CT_AbstractNumID{Val: &nextAbsID},
    })

    // 5. Write merged numbering.xml
    d.writeNumbering(nb)

    return &ListBuilder{
        doc:   d,
        numID: nextNumID,
    }
}
```

**Level generation pattern** (ordered vs bulleted, RESEARCH.md:460-463):
```go
func makeLevels(ordered bool) []*wml.CT_Lvl {
    lvls := make([]*wml.CT_Lvl, 9)
    for i := int64(0); i < 9; i++ {
        fmt := "bullet"
        text := "\u2022"
        if ordered {
            fmt = "decimal"
            text = "%1."
        }
        start := int64(1)
        lvls[i] = &wml.CT_Lvl{
            ILvl:    &i,
            NumFmt:  &wml.CT_NumFmt{Val: &fmt},
            LvlText: &wml.CT_LvlText{Val: &text},
            Start:   &wml.CT_Start{Val: &start},
        }
    }
    return lvls
}
```

**CT_NumPr assignment on paragraphs** (internal/wml/properties.go:44-49):
```go
// On each list item paragraph:
p.ct.PPr = &wml.CT_PPr{}
p.ct.PPr.NumPr = &wml.CT_NumPr{
    ILvl:  &wml.CT_ILvl{Val: &level},
    NumId: &wml.CT_NumId{Val: &numID},
}
```

---

### `wordingo/page.go` — Page setup (public API, request-response/mutate)

**Analog:** `wordingo/paragraph.go` SetAlignment pattern (mutate-and-return-receiver)

**SetOrientation pattern** (setter on existing CT_SectPr, create.go:204-211 for defaultSectPr):
```go
const (
    OrientationPortrait  PageOrientation = iota
    OrientationLandscape
)

func (d *Document) SetOrientation(o PageOrientation) *Document {
    if d.doc.Body.SectPr == nil {
        d.doc.Body.SectPr = defaultSectPr()
    }
    sz := d.doc.Body.SectPr.PgSz
    if o == OrientationLandscape {
        // Swap: W > H indicates landscape
        if *sz.W < *sz.H {
            *sz.W, *sz.H = *sz.H, *sz.W
        }
    } else {
        if *sz.W > *sz.H {
            *sz.W, *sz.H = *sz.H, *sz.W
        }
    }
    d.dirty = true
    return d
}
```

**AddPageBreak pattern** (one-liner creating paragraph with PageBreakBefore):
```go
func (d *Document) AddPageBreak() *Paragraph {
    p := d.AddParagraph("")
    p.ct.PPr = &wml.CT_PPr{
        PageBreakBefore: &wml.CT_OnOff{Val: ptrBool(true)},
    }
    d.dirty = true
    return p
}
```

**Paper size presets pattern** (create.go:152-195 for default twips):
```go
const (
    PaperLetter int64 = 12240 // W
    PaperLetterH int64 = 15840 // H
    PaperA4    int64 = 11906
    PaperA4H   int64 = 16838
    PaperLegal int64 = 12240
    PaperLegalH int64 = 20160
)

func (d *Document) SetPaperSize(w, h int64) *Document {
    if d.doc.Body.SectPr == nil {
        d.doc.Body.SectPr = defaultSectPr()
    }
    d.doc.Body.SectPr.PgSz.W = &w
    d.doc.Body.SectPr.PgSz.H = &h
    d.dirty = true
    return d
}
```

---

### `wordingo/hyperlink.go` — Hyperlink on Paragraph (public API)

**Analog:** `wordingo/run.go:9-19` (Run type wrapper + X()), `wordingo/paragraph.go:42-52` (AddRun pattern)

**AddHyperlink pattern** (paragraph.go:42-52 — method on Paragraph, returns Run):
```go
func (p *Paragraph) AddHyperlink(text, uri string) *Run {
    if p == nil {
        panic("wordingo: AddHyperlink called on nil Paragraph")
    }

    // 1. Create relationship in document.xml.rels
    rels := p.doc.pkg.Rels["word/document.xml"]
    rId := rels.NextRID()
    rels.Rels = append(rels.Rels, opc.Relationship{
        ID:         rId,
        Type:       relHyperlink,
        Target:     uri,
        TargetMode: "External", // CRITICAL: hyperlinks need TargetMode
    })

    // 2. Create CT_R with text
    run := &wml.CT_R{T: &wml.CT_Text{Value: text}}

    // 3. Create CT_Hyperlink containing the run
    link := &wml.CT_Hyperlink{
        ID: rId,
        R:  []*wml.CT_R{run},
    }

    // 4. Append to paragraph
    // CT_P.Hyperlink field added in document.go
    p.ct.Hyperlink = append(p.ct.Hyperlink, link)
    p.doc.dirty = true

    return &Run{ct: run, para: p}
}
```

---

### `internal/wml/drawing.go` — DrawingML inline types (WML schema, serialization)

**Analog:** `internal/wml/table.go:9-16` (CT_Tbl pattern — XMLName + namespace + RawXML)

**Full CT_Drawing/CT_Inline pattern** (namespace constants in namespaces.go:16-20):
```go
package wml

import (
    "encoding/xml"
    "github.com/fabiomarini/wordingo/internal/xmlutil"
)

// CT_Drawing wraps a DrawingML element (w:drawing).
type CT_Drawing struct {
    XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main drawing"`
    Inline  *CT_Inline `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing inline"`
    Anchor  *CT_Anchor `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing anchor"`
    Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_Inline is an inline DrawingML object (wp:inline).
type CT_Inline struct {
    XMLName      xml.Name       `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing inline"`
    Extent       *CT_Extent     `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing extent"`
    DocPr        *CT_DocPr      `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing docPr"`
    Graphic      *CT_Graphic    `xml:"http://schemas.openxmlformats.org/drawingml/2006/main graphic"`
    Raw          []xmlutil.RawXML `xml:",any"`
}

// CT_Extent is the drawing extent in EMU.
type CT_Extent struct {
    XMLName xml.Name `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing extent"`
    Cx      int64    `xml:"cx,attr"`
    Cy      int64    `xml:"cy,attr"`
}

// CT_BlipFill is picture fill (pic:blipFill).
type CT_BlipFill struct {
    XMLName xml.Name `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture blipFill"`
    Blip    *CT_Blip `xml:"http://schemas.openxmlformats.org/drawingml/2006/main blip"`
    Stretch *CT_Stretch `xml:"http://schemas.openxmlformats.org/drawingml/2006/main stretch"`
    Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_Blip is a reference to an embedded image.
type CT_Blip struct {
    XMLName xml.Name `xml:"http://schemas.openxmlformats.org/drawingml/2006/main blip"`
    Embed   string   `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships embed,attr"`
    Raw     []xmlutil.RawXML `xml:",any"`
}
```

---

### `internal/wml/hyperlink.go` — CT_Hyperlink (WML schema, serialization)

**Analog:** `internal/wml/table.go` (CT_Tr pattern — XMLName + children)

**CT_Hyperlink pattern**:
```go
package wml

import (
    "encoding/xml"
    "github.com/fabiomarini/wordingo/internal/xmlutil"
)

// CT_Hyperlink is a hyperlink at paragraph level (w:hyperlink).
type CT_Hyperlink struct {
    XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hyperlink"`
    ID      string   `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
    R       []*CT_R  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
    Raw     []xmlutil.RawXML `xml:",any"`
}
```

---

### `wordingo.go` — Add methods (modify, public API entry points)

**Analog:** Existing `AddParagraph` (wordingo.go:222-237)

**AddTable entry points** (wordingo.go:222-237):
```go
// AddTable creates a simple table from string data.
func (d *Document) AddTable(data [][]string) (*TableBuilder, error) {
    if d == nil {
        panic("wordingo: AddTable called on nil Document")
    }
    // Create CT_Tbl with grid defined by max column count
    // Create CT_Tr for each row, CT_Tc for each cell with CT_P + CT_R{T}
    // Append to Body.Tbl
    // set d.dirty = true
}

// AddTableBuilder returns a new table builder for complex tables.
func (d *Document) AddTableBuilder() *TableBuilder {
    // Return builder for user to chain
}
```

**Tables() accessor** (wordingo.go:211-220 — Paragraphs() pattern):
```go
func (d *Document) Tables() []*TableBuilder {
    if d.doc == nil || d.doc.Body == nil {
        return nil
    }
    tbls := make([]*TableBuilder, len(d.doc.Body.Tbl))
    for i, t := range d.doc.Body.Tbl {
        tbls[i] = &TableBuilder{ct: t, doc: d}
    }
    return tbls
}
```

---

### `internal/wml/document.go` — CT_P Hyperlink + CT_R Drawing fields (modify, WML schema)

**CT_P field addition** (document.go:52-57):
```go
// Add to CT_P struct after R field:
Hyperlink []*CT_Hyperlink `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hyperlink"`
// CT_P currently at line 55: R []*CT_R — add Hyperlink after R
```

**CT_R field addition** (document.go:79-87):
```go
// Add to CT_R struct after Cr field:
Drawing *CT_Drawing `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main drawing"`
// CT_R currently at line 86: Raw — add Drawing before Raw
```

---

### `template.go` — Header/footer cloning (modify, OPC transform)

**Analog:** Existing OpenTemplateReader lines 76-133, CloneStyles pattern

**Header cloning pattern** (reverses Pitfall 4, inserts between CloneStyles and body copy):
```go
// After CloneStyles(src, dst) in OpenTemplateReader:

// 1. Parse source sectPr for header/footer refs
srcRels := src.Rels["word/document.xml"]
dstRels := dst.Rels["word/document.xml"]

// 2. For each HdrFtrRef in source sectPr:
//    a. Find the relationship in src rels by rId
//    b. Resolve target path relative to word/
//    c. Copy part bytes from src to dst via d.pkg.MarkModified
//    d. Create new rId via dstRels.NextRID()
//    e. Add relationship entry with new rId, same Type, same Target
//    f. Add content type override
//    g. Remap src HdrFtrRef.ID to new rId in destination sectPr

// 3. Do NOT use defaultSectPr() — use src sectPr with fixed HdrFtrRef
```

---

### `create.go` — New constants (modify, config)

**Content type constants** (create.go:28-37):
```go
const (
    ctHeader = "application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"
    ctFooter = "application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"
    ctPng    = "image/png"
    ctJpeg   = "image/jpeg"
)
```

**Relationship type constants** (create.go:39-46):
```go
const (
    relHeader    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/header"
    relFooter    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer"
    relImage     = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"
    relHyperlink = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink"
)
```

---

## Shared Patterns

### 1. Wrapper-over-WML with X() Escape Hatch
**Source:** `wordingo/paragraph.go:9-13, 34-40`, `wordingo/run.go:9-19`
**Apply to:** All new public types (TableBuilder, Header, Footer, ListBuilder, Section)

```go
// Every public wrapper holds a pointer to the WML CT_ type
type Paragraph struct {
    ct  *wml.CT_P
    doc *Document
}

// Every wrapper exposes X() for escape-hatch access
func (p *Paragraph) X() *wml.CT_P {
    if p == nil {
        panic("wordingo: X called on nil Paragraph")
    }
    return p.ct
}
```

### 2. Builder Chaining (SetXxx returns receiver)
**Source:** `wordingo/paragraph.go:55-70` (SetAlignment), `wordingo/run.go:21-31` (SetBold)
**Apply to:** TableBuilder, RowBuilder, CellBuilder, Header, ListBuilder

```go
// Pattern: nil-check → lazy-init props → set field → doc.dirty = true → return self
func (p *Paragraph) SetAlignment(a Alignment) *Paragraph {
    if p == nil {
        panic("wordingo: SetAlignment called on nil Paragraph")
    }
    if p.ct.PPr == nil {
        p.ct.PPr = &wml.CT_PPr{}
    }
    val := a.String()
    p.ct.PPr.Jc = &wml.CT_Jc{Val: &val}
    p.doc.dirty = true
    return p
}
```

### 3. Deferred Error via Warnings()
**Source:** `wordingo.go:100-115`, `wordingo/run.go:73-75`
**Apply to:** All new files — non-fatal issues call `doc.warn()` instead of error return

```go
func (r *Run) SetSize(pts float64) *Run {
    if pts < 0 {
        r.para.doc.warn("wordingo: negative font size %f", pts)
        return r
    }
    // ... proceed with valid value
}
```

### 4. Nil Receiver Guard
**Source:** Every public method in `paragraph.go`, `run.go`, `wordingo.go`
**Apply to:** Every public method

```go
func (p *Paragraph) AddRun(text string) *Run {
    if p == nil {
        panic("wordingo: AddRun called on nil Paragraph")
    }
    // ...
}
```

### 5. OPC Part Creation (MarkModified + Rels + ContentTypes)
**Source:** `create.go:48-97` (newBlankPackage), `internal/opc/relationships.go:138-145` (NextRID)
**Apply to:** Image embedding, header/footer creation, hyperlink creation

```go
// Three-step OPC part registration:
// 1. Part payload
d.pkg.MarkModified(partName, data)

// 2. Relationship (always use NextRID, never hardcode)
rels := d.pkg.Rels["word/document.xml"]
rId := rels.NextRID()
rels.Rels = append(rels.Rels, opc.Relationship{ID: rId, Type: relType, Target: target})

// 3. Content type override
d.pkg.ContentTypes.Overrides["/"+partName] = ctType
```

### 6. Body Serialization at Save
**Source:** `wordingo.go:178-200` (serializeBody)
**Apply to:** Any new feature that modifies CT_Document contents

```go
func (d *Document) serializeBody() {
    if !d.dirty { return }
    var buf bytes.Buffer
    buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
    buf.WriteByte('\n')
    enc := xmlutil.NewEncoder(&buf)
    if err := enc.Encode(d.doc); err != nil {
        d.warn("wordingo: serialize body: %v", err)
        return
    }
    enc.Flush()
    d.pkg.MarkModified("word/document.xml", buf.Bytes())
    d.dirty = false
}
```

### 7. WML CT_ Struct Convention (XMLName + Namespace + RawXML)
**Source:** `internal/wml/table.go:9-16`, `internal/wml/document.go:170-181`
**Apply to:** New types in `internal/wml/drawing.go`, `internal/wml/hyperlink.go`

```go
type CT_Tbl struct {
    XMLName xml.Name        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tbl"`
    TblPr   *CT_TblPr       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblPr"`
    // children with full namespace URI
    Raw     []xmlutil.RawXML `xml:",any"`  // always hoard unknown children
}
```

### 8. Test Pattern (Create → Call → Assert)
**Source:** `style_test.go:13-42`
**Apply to:** All new public API surface

```go
func TestTableBuilder(t *testing.T) {
    doc, err := Create()
    if err != nil {
        t.Fatal(err)
    }
    tbl := doc.AddTableBuilder()
    tbl.SetTableStyle("LightGrid-Accent1")
    if tbl.ct.TblPr == nil || tbl.ct.TblPr.TblStyle == nil || *tbl.ct.TblPr.TblStyle.Val != "LightGrid-Accent1" {
        t.Fatal("TblStyle not set")
    }
}
```

---

## Deviation Notes

| File | Pattern | Deviation | Why |
|---|---|---|---|
| `wordingo/table.go` | Builder chaining | Table has 3 levels (TableBuilder → RowBuilder → CellBuilder) vs Paragraph's single level | Complex table API needs row/cell nesting. Each sub-builder holds doc ref for dirty flag. |
| `wordingo/image.go` | Builder chaining | Does NOT return self — returns `*Run` | Image adds a run to the last paragraph or a specified paragraph. Different from paragraph stutter-chaining. |
| `wordingo/header.go` | OPC part pattern | Must serialize CT_Hdr/CT_Ftr to XML BEFORE MarkModified (not lazy like styles) | Header content is user-provided, not pre-baked. Must encode then store in OPC part. |
| `wordingo/list.go` | Builder chain + CRUD | Must read existing numbering.xml, parse, append, and re-serialize | Lists need to merge auto-generated defs with template numbering defs. Pure CRUD. |
| `template.go` | OpenTemplate pattern | Reverses Phase 3 Pitfall 4 — adds header/footer cloning instead of stripping | HdrFtrRef stripping was always a bug. Phase 5 corrects it. Must handle rId fixup. |
| `wordingo/hyperlink.go` | Relationship pattern | TargetMode="External" required | All other relationships are internal. Hyperlinks to URLs need External mode. |
| `internal/wml/drawing.go` | CT_ struct pattern | Uses 3 different namespaces (wp:, a:, pic:) | DrawingML spans multiple XML namespaces. Each struct must use correct ns URI. |

---

## Pitfall Warnings

### P1: CT_P + CT_R fields not modeled before plan execution
**File(s):** `internal/wml/document.go`
**Action:** First task of **every sub-plan** must add `CT_P.Hyperlink` and `CT_R.Drawing` fields. Compile error otherwise.
**Pattern:** See `internal/wml/document.go:52-57` modified section above.

### P2: Table body insertion breaks paragraph order
**File(s):** `wordingo.go` (serializeBody), `internal/wml/document.go` (CT_Body)
**Action:** V1 accepts tables-at-end limitation. Use `Body.P + Body.Tbl` separate slices. Phase 6 adds ordered body tracking.
**This is explicitly deferred** per D-24.

### P3: Numbering def collision with template
**File(s):** `wordingo/list.go`
**Action:** Always scan ALL existing NumID/AbstractNumID in parsed numbering.xml. Compute `maxFound + 1`. Start auto-generated numIds at 1 (reserve 0 for Word's ListNumber).

### P4: Image DPI detection edge cases
**File(s):** `wordingo/image.go`
**Action:** Default 72 DPI when no metadata. Provide explicit `SetWidth`/`SetHeight` overrides. JPEG EXIF orientation may swap W/H — document as v1 limitation.

### P5: Header/footer relationship target relative to word/
**File(s):** `wordingo/header.go`
**Action:** Relationship targets in `word/_rels/document.xml.rels` must be relative to `word/`: `header1.xml`, NOT `word/header1.xml`.

### P6: OpenTemplate header rId fixup required
**File(s):** `template.go`
**Action:** Source rIds (e.g., "rId2") may clash with destination rIds. Must allocate fresh rIds via `dstRels.NextRID()` for every cloned header/footer part.

### P7: Hyperlink TargetMode="External" required
**File(s):** `wordingo/hyperlink.go`
**Action:** `Relationship{TargetMode: "External"}` for hyperlinks. Omit for internal targets (images, headers, styles).

### P8: Per-cell MarkModified calls forbidden
**File(s):** `wordingo/table.go`
**Action:** Table construction sets `dirty = true` on Document. SerializeBody re-encodes once at Save. Never call MarkModified per cell/row.

### P9: Header/footer cloning must add content type override
**File(s):** `template.go`
**Action:** Each cloned part needs `ContentTypes.Overrides["/word/header1.xml"] = ctHeader`. Missing override → Word repair dialog.

---

## No Analog Found

All files have close analogs in the existing codebase.

---

## Metadata

**Analog search scope:** `wordingo/`, `wordingo/internal/wml/`, `wordingo/internal/opc/`, `wordingo/internal/style/`
**Files scanned:** ~20 Go source files
**Pattern extraction date:** 2026-07-26
