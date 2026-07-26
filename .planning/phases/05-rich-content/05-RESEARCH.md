# Phase 5: Rich Content - Research

**Researched:** 2026-07-26
**Domain:** Tables, images, headers/footers, lists, hyperlinks, page setup — complete business document creation
**Confidence:** HIGH

## Summary

Phase 5 extends Phase 4's content API with six feature domains: tables, images, headers/footers, lists, hyperlinks, and page setup. All six build on existing WML types already modeled in `internal/wml/`. No new external dependencies — pure Go stdlib.

**Core insight:** The infrastructure exists. CT_Tbl, CT_Hdr, CT_Ftr, CT_Numbering, CT_Lvl, CT_PgSz, CT_PgMar, CT_HdrFtrRef all modeled. Three structural gaps block implementation:

1. **CT_R lacks Drawing field** — must add `Drawing *CT_Drawing` for inline/floating images (new `internal/wml/drawing.go` with ~5 DrawingML structs).
2. **CT_P lacks Hyperlink child** — must add `Hyperlink []*CT_Hyperlink` to CT_P (new `internal/wml/hyperlink.go` with ~2 structs).
3. **OpenTemplate strips HdrFtrRef** (Phase 3 D-04/Pitfall 4) — Phase 5 reverses by cloning header/footer parts from template alongside CloneStyles.

**Primary recommendation:** Three-plan decomposition (tables+images, hdrftr+pageSetup, lists+hyperlinks) is correct and dependency-safe if 05-01 adds Drawing + Hyperlink WML types first.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**D-01 (Tables):** Two entry points — `doc.AddTable(data [][]string)` for simple data grids + `doc.AddTableBuilder()` for complex tables. Matches Phase 4 builder chaining pattern.

**D-02 (Tables):** Borders, shading, cell merge applied at row level with per-cell overrides. `table.Row(r).SetBorders(...)` bulk. `table.Row(r).Cell(c).SetShading(...)` for exceptions.

**D-03 (Tables):** Named table styles supported — `table.SetTableStyle("LightGrid-Accent1")`. Sets tblStyle on TblPr. Works with cloned template styles.

**D-04 (Tables):** Table width defaults to auto (Word content-based sizing). `table.SetWidth(w int64, wType string)` overrides.

**D-05 (Images):** Minimal DrawingML inline types modeled in `internal/wml/`: CT_Inline, CT_Blip, CT_BlipFill, CT_Extent, CT_NonVisualPicProps. ~5 structs covering PNG/JPEG embedding.

**D-06 (Images):** Two entry points: `doc.AddImage(path)` and `doc.AddImageBytes(name, data, ct)`. File path for convenience, raw bytes for programmatic/network sources.

**D-07 (Images):** Sizing auto-defaults from image DPI. Optional `SetWidth(inches)/SetHeight(inches)` overrides. Library converts inches to EMUs internally.

**D-08 (Images):** Both inline (wp:inline) and floating (wp:anchor) placement supported in v1.

**D-09 (Images):** Image creates media part in `word/media/`, adds relationship from `document.xml.rels`, adds content type override.

**D-10 (Hdr/Ftr):** One-shot creation — `doc.AddHeader(HeaderDefault)` creates header part, links via sectPr, returns `*Header` for content. Same pattern for footer.

**D-11 (Hdr/Ftr):** All three variants in v1: HeaderDefault, HeaderFirst (titlePg), HeaderEven (odd/even).

**D-12 (Hdr/Ftr):** Footer API symmetric with Header — `doc.AddFooter(FooterDefault)`. Both wrap same internal type.

**D-13 (Hdr/Ftr):** OpenTemplate clones header/footer parts alongside styles (reversing Phase 3 Pitfall 4). Header/footer rIds fixed up for the fresh package.

**D-14 (Lists):** Two entry points: `doc.AddList(ordered bool)` builder + `doc.AddListFromSlice(items []string, ordered bool)` for flat lists.

**D-15 (Lists):** Numbering defs auto-generated internally. `doc.AddNumberingDef(fmt, start)` available for custom definitions.

**D-16 (Lists):** Single `AddList(ordered bool)` — true = ordered (decimal), false = bulleted. Multi-level via `AddItem("text", level)`.

**D-17 (Lists):** Full 9-level depth (0-8). Each level configurable via CT_Lvl.

**D-18 (Page Setup):** Both doc-level convenience methods and Section object — `doc.SetOrientation(...)` delegates to `doc.Section().SetOrientation(...)`.

**D-19 (Page Setup):** `doc.SetOrientation(OrientationLandscape | OrientationPortrait)` swaps W/H on default sectPr.

**D-20 (Page Setup):** `para.SetPageBreakBefore(true)` via CT_PPr.PageBreakBefore. `doc.AddPageBreak()` convenience inserts paragraph with pageBreakBefore.

**D-21 (Page Setup):** Paper size presets (PaperLetter, PaperA4, PaperLegal) + `doc.SetPaperSize(w, h twips)` for custom.

**D-22 (Hyperlinks):** `para.AddHyperlink(text, uri)` on Paragraph. Creates w:hyperlink element containing w:r with text.

**D-23 (Hyperlinks):** Auto rId — AddHyperlink creates relationship entry in `word/_rels/document.xml.rels`, assigns next available rId.

**D-24 (Plan Order):** Keep ROADMAP order: 05-01 Tables+images, 05-02 HdrFtr+sections+pageSetup, 05-03 Lists+hyperlinks.

### The Agent's Discretion
- Exact DrawingML struct field names for CT_Inline, CT_Blip etc.
- Table struct layout within `wordingo/` package (table.go)
- Image format detection and DPI parsing strategy
- Header/footer internal file layout
- List numbering def auto-generation algorithm (abstractNumId allocation)
- PageOrientation enum values
- Hyperlink relationship cleanup (duplicate URI detection)

### Deferred Ideas (OUT OF SCOPE)
None.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| API-04 | Tables — rows, cells, shading, borders, widths, hMerge/vMerge, named table styles | CT_Tbl, CT_TblPr, CT_Tr, CT_Tc, CT_TcPr, CT_TblBorders, CT_GridSpan, CT_VMerge all modeled in `internal/wml/table.go`. CT_Body.Tbl field exists alongside P. Table builder API follows Phase 4 Paragraph pattern. |
| API-05 | Images — PNG/JPEG with DrawingML anchors, explicit sizing | New DrawingML types needed (~5 structs in new `internal/wml/drawing.go`). CT_R needs Drawing field. Media part creation pattern understood (MarkModified + rel + content type override). EMU conversion formula: inches × 914400. |
| API-06 | Lists — ordered/bulleted, multi-level, numbering-definition-backed | CT_Numbering, CT_Num, CT_AbstractNum, CT_Lvl all modeled in `internal/wml/numbering.go`. CT_NumPr (numId, ilvl) in `properties.go`. Numbering resolver in `internal/style/numbering.go`. Auto-generation algorithm for abstractNumId/numId allocation. |
| API-07 | Headers/footers — default, first-page, odd/even variants per section | CT_Hdr, CT_Ftr, CT_HdrFtrRef, CT_SectPr.HdrFtrRef/FtrRef/TitlePg all modeled in `internal/wml/document.go`. OPC part creation + rel + content type override pattern needed. OpenTemplate reverse of Pitfall 4. |
| API-08 | Hyperlinks on runs | New CT_Hyperlink type needed (~2 structs in new `internal/wml/hyperlink.go`). CT_P needs Hyperlink field. Relationship entry in `word/_rels/document.xml.rels` created via opc.Relationships.NextRID(). relHyperlink relationship type constant. |
| API-09 | Page setup — margins, orientation, paper size; page breaks | CT_PgSz (W/H/Code), CT_PgMar (Top/Right/Bottom/Left/Header/Footer), CT_SectPr all modeled. CT_PPr.PageBreakBefore exists. defaultSectPr() in create.go. Orientation swap of W/H twips. Paper size presets (Letter=12240×15840, A4=11906×16838, Legal=12240×20160). |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Table creation (grid, complex) | API / Backend | Database / Storage | Pure data mutation on CT_Tbl within CT_Body. Table pushes paragraphs above it in XML order. Body serialization at Save re-encodes full document. |
| Image embedding | API / Backend | Database / Storage | Read image bytes → create word/media/ part → add rel → add content type override → insert DrawingML reference in CT_R. |
| Header/footer parts | API / Backend | — | New OPC parts (word/header1.xml etc.) with their own CT_Hdr/CT_Ftr content. Relationships from document.xml.rels. SectPr HdrFtrRef links. No persistence tier — parts go through OPC layer. |
| List numbering defs | API / Backend | — | Auto-generate CT_Numbering entries in memory. Write to word/numbering.xml via MarkModified. |
| Hyperlinks | API / Backend | — | Add relationship in document.xml.rels + insert CT_Hyperlink in CT_P + add CT_R child with text. |
| Page setup | API / Backend | — | Mutate CT_SectPr PgSz/PgMar on default sectPr (or Section wrapper). PageBreakBefore on CT_PPr. |
| Body serialization | API / Backend | Database / Storage | Re-encode CT_Document + new parts at Save via dirty flag (Phase 4 pattern). |

## Standard Stack

### Core

Zero external dependencies — Go stdlib only [VERIFIED: go.mod].

| Package | Purpose | Why Standard |
|---------|---------|--------------|
| `github.com/fabiomarini/wordingo` (root) | Public Document API — AddTable, AddImage, AddHeader, AddList, SetOrientation | PRD §8, D-01..D-24: single public package. Extends Document type with 6 new feature methods. |
| `internal/wml/table.go` | CT_Tbl, CT_TblPr, CT_TblGrid, CT_Tr, CT_Tc, CT_TcPr, CT_TblBorders, CT_TblBorder, CT_GridSpan, CT_VMerge, CT_GridCol, CT_TblW, CT_TblStyle | All table types already modeled and ready. CT_Body.Tbl field exists alongside P. No new WML types needed for tables. |
| `internal/wml/document.go` | CT_Hdr, CT_Ftr, CT_HdrFtrRef, CT_SectPr (PgSz, PgMar, HdrFtrRef, FtrRef, TitlePg), CT_PPr.PageBreakBefore | Header/footer types modeled. Section properties complete. CT_P needs: `Hyperlink []*CT_Hyperlink` field added. |
| `internal/wml/numbering.go` | CT_Numbering, CT_Num, CT_AbstractNum, CT_Lvl, CT_NumFmt, CT_LvlText, CT_Start, CT_LvlOverride | Full list-backed numbering types available. Reuses Phase 2 marshaling. |
| `internal/wml/properties.go` | CT_NumPr (numId, ilvl), CT_Shd (reusable for table cells) | Numbering properties existing. Shading reusable for table cells and paragraphs. |
| `internal/wml/drawing.go` | **NEW** — CT_Drawing, CT_Inline, CT_Blip, CT_BlipFill, CT_Extent, CT_NonVisualPicProps, CT_Anchor | ~5-7 DrawingML structs for image embedding (w:drawing → wp:inline/wp:anchor → a:blip). |
| `internal/wml/hyperlink.go` | **NEW** — CT_Hyperlink with ID attr (rId) + child CT_R elements | ~2 structs. Must set xml.Name to `w:hyperlink`. |
| `internal/opc/relationships.go` | NextRID(), relationship creation patterns | Hyperlinks and images need new relationships in document.xml.rels. NextRID() pattern established. |
| `internal/opc/contenttypes.go` | Override map for new part types | Header/footer parts, image parts need content type overrides. |
| `internal/style/numbering.go` | Numbering resolution for list defs | Reuse numberingCache for abstractNumId lookup? Or build auto-generation logic in wordingo package. |

### Relationship Type Constants (to add)

| Constant | Value | Used By |
|----------|-------|---------|
| `relHeader` | `http://schemas.openxmlformats.org/officeDocument/2006/relationships/header` | Header part links |
| `relFooter` | `http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer` | Footer part links |
| `relImage` | `http://schemas.openxmlformats.org/officeDocument/2006/relationships/image` | Image part links |
| `relHyperlink` | `http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink` | Hyperlink relationship |

### Content Type Constants (to add)

| Constant | Value | Used By |
|----------|-------|---------|
| `ctHeader` | `application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml` | Header parts |
| `ctFooter` | `application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml` | Footer parts |
| `ctPng` | `image/png` | PNG images |
| `ctJpeg` | `image/jpeg` | JPEG images |

### Paper Size Presets (Twips)

| Preset | W | H | Notes |
|--------|---|---|-------|
| PaperLetter | 12240 | 15840 | US Letter (8.5×11 in) |
| PaperA4 | 11906 | 16838 | A4 (210×297 mm) |
| PaperLegal | 12240 | 20160 | Legal (8.5×14 in) |

### EMU Conversion

```
const emusPerInch = 914400
emus = inches * emusPerInch
```

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
Document (in-memory CT_Document + OPC Package)
   │
   ├── Tables ──────────────────────────────────────────────────
   │   AddTable(data [][]string)           → appends CT_Tbl to Body.Tbl
   │   AddTableBuilder() *TableBuilder     → returns builder for complex tables
   │     .Row(i).SetBorders(...)           → sets CT_TcPr borders on row cells
   │     .Row(i).Cell(j).SetShading(...)   → sets CT_TcPr.Shd
   │     .SetTableStyle("...")             → sets CT_TblPr.TblStyle
   │     .SetWidth(w, type)                → sets CT_TblPr.TblW
   │     .Row(i).Cell(j).MergeRight()      → sets CT_TcPr.GridSpan
   │     .Row(i).Cell(j).MergeDown()       → sets CT_TcPr.VMerge
   │
   ├── Images ──────────────────────────────────────────────────
   │   AddImage(path)                      → reads file → AddImageBytes
   │   AddImageBytes(name, data, ct)       → core method
   │     ├── MarkModified("word/media/name.png", data)
   │     ├── pkg.Rels["word/document.xml"].Rels += {ID: rIdN, Type: relImage, Target: "media/name.png"}
   │     ├── ContentTypes.Overrides["/word/media/name.png"] = "image/png"
   │     └── Insert CT_R with Drawing → CT_Drawing → a:graphic → wp:inline/wp:anchor (+ CT_Extent, CT_Blip, CT_BlipFill)
   │
   ├── Headers/Footers ─────────────────────────────────────────
   │   AddHeader(HeaderDefault)            → core method
   │     ├── wml.CT_Hdr{...}              → MarshalXML → MarkModified("word/header1.xml", bytes)
   │     ├── pkg.Rels["word/document.xml"].Rels += {ID: rIdN, Type: relHeader, Target: "header1.xml"}
   │     ├── ContentTypes.Overrides["/word/header1.xml"] = ctHeader
   │     └── SectPr.HdrFtrRef += &CT_HdrFtrRef{ID: rIdN, Type: "default"}
   │   AddFooter(FooterDefault)            → symmetric with AddHeader
   │     └── CT_HdrFtrRef.Type: "default" | "first" | "even"
   │
   ├── Lists ────────────────────────────────────────────────────
   │   AddList(ordered bool) ListBuilder   → creates auto numbering def
   │     .AddItem("text", level)           → appends paragraph with CT_NumPr
   │     ├── Auto-generate CT_AbstractNum (levels 0-8, numFmt, lvlText)
   │     ├── Auto-generate CT_Num (links abstractNum → numId)
   │     ├── MarkModified("word/numbering.xml", bytes)
   │     └── Each item: CT_P with PPr.NumPr = {numId: N, ilvl: level}
   │
   ├── Hyperlinks ───────────────────────────────────────────────
   │   para.AddHyperlink(text, uri)        → adds to paragraph
   │     ├── pkg.Rels["word/document.xml"].Rels += {ID: rIdN, Type: relHyperlink, Target: uri, TargetMode: "External"}
   │     └── Append CT_Hyperlink{ID: rIdN, R: [CT_R{text}]} to ct.Hyperlink
   │
   ├── Page Setup ───────────────────────────────────────────────
   │   SetOrientation(OrientationLandscape) → swap W/H on SectPr.PgSz
   │   SetPaperSize(PaperA4)                → set PgSz.W, PgSz.H from preset
   │   SetMargins(top, right, bottom, left) → set PgMar values in twips
   │   AddPageBreak()                       → Insert paragraph with PPr.PageBreakBefore
   │
   └── WriteTo / Save
         ├── serializeBody() if dirty (Phase 4 pattern)
         └── pkg.Save(w)
               ├── word/document.xml        → re-encoded (modified)
               ├── word/media/*             → new parts (modified)
               ├── word/header1.xml / footer1.xml → new parts (modified)
               ├── word/numbering.xml       → re-encoded if lists used (modified)
               ├── [Content_Types].xml      → re-encoded with new overrides
               ├── word/_rels/document.xml.rels → re-encoded with new relationships
               └── unmodified parts (styles, theme, etc.) → raw byte pass-through
```

### Recommended Project Structure

```
wordingo/                              # Public root package
├── wordingo.go                        # Document: Create, Save, Warnings (exists)
├── paragraph.go                       # Paragraph: AddRun, SetAlignment, etc. (exists)
├── run.go                             # Run: formatting setters (exists)
├── create.go                          # newBlankPackage, buildDocumentXML (exists)
├── open.go                            # Open, OpenReader (exists)
├── template.go                        # FromTemplate, OpenTemplate (exists)
│
├── table.go                           # NEW: TableBuilder type, RowBuilder, CellBuilder
│                                      #   AddTable(), AddTableBuilder(), SetTableStyle()
│                                      #   SetWidth, SetBorders, SetShading, MergeRight/MergeDown
│                                      #   Table, Row, Cell accessors
│
├── image.go                           # NEW: Image embedding support
│                                      #   AddImage(path), AddImageBytes(name, data, ct)
│                                      #   DPI detection, EMU conversion helpers
│
├── header.go                          # NEW: Header/Footer types + creation
│                                      #   AddHeader(HeaderVariant), AddFooter(FooterVariant)
│                                      #   HeaderVariant enum (Default, First, Even)
│                                      #   Header/Footer wrappers with content API
│
├── list.go                            # NEW: List builder + numbering def auto-generation
│                                      #   AddList(ordered bool), AddListFromSlice(items, ordered)
│                                      #   ListBuilder with AddItem(text, level)
│                                      #   numbering def auto-generation (abstractNumId alloc)
│
├── page.go                            # NEW: Page setup methods + Section wrapper
│                                      #   SetOrientation, SetPaperSize, SetMargins
│                                      #   AddPageBreak()
│                                      #   Section type, paper size presets
│                                      #   PageOrientation enum
│
├── hyperlink.go                       # NEW: Hyperlink creation
│                                      #   Paragraph.AddHyperlink(text, uri)
│
├── format.go                          # RunFormat, ParFormat (exists from Phase 4)
│
├── internal/wml/
│   ├── document.go                    # exists — add Hyperlink field to CT_P
│   ├── table.go                       # exists — all table types ready
│   ├── numbering.go                   # exists — all numbering types ready
│   ├── drawing.go                     # NEW: CT_Drawing, CT_Inline, CT_Blip, CT_BlipFill
│   │                                  #       CT_Extent, CT_NonVisualPicProps, CT_Anchor
│   ├── hyperlink.go                   # NEW: CT_Hyperlink
│   └── properties.go                  # exists — CT_Shd reusable for table cells
```

### Pattern 1: Table Builder (Two-Entry-Point, Matches Phase 4)

**What:** Simple grids via `AddTable(data [][]string)`. Complex tables via `AddTableBuilder()` returning builder chain. Follows Phase 4's Paragraph builder pattern.

**When to use:** All table creation.

**Example:**
```go
// Source: D-01, D-02, D-03, D-04

// Simple data grid — one line
doc.AddTable([][]string{
    {"Name", "Amount", "Date"},
    {"Alice", "$100", "2026-01-15"},
})

// Complex table via builder
tbl := doc.AddTableBuilder()
tbl.SetTableStyle("LightGrid-Accent1")
tbl.SetWidth(9000, "dxa") // 5 inches

row0 := tbl.Row(0)
row0.Cell(0).SetText("Name").SetShading("clear", "D9E2F3")
row0.Cell(1).SetText("Amount").SetBold(true)
row0.Cell(2).SetText("Date")

row1 := tbl.Row(1)
row1.Cell(0).SetText("Alice")
row1.Cell(1).MergeRight() // gridSpan=2
row1.Cell(2).SetText("2026-01-15")
```

### Pattern 2: Image Embedding — OPC Part + Relationship + Content Type Override

**What:** Images create a new part in `word/media/`, a relationship from `document.xml.rels`, an override in `[Content_Types].xml`, and a DrawingML inline/floating element in the run.

**When to use:** Every `AddImage` / `AddImageBytes` call.

**Example:**
```go
// Source: D-05..D-09, established MarkModified pattern

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
    
    // 4. Create CT_R with DrawingML inline
    extent := CT_Extent{Cx: emusPerInch * 3, Cy: emusPerInch * 2} // 3×2 inches
    inline := &CT_Inline{
        Extent: &extent,
        Graphic: &a_Graphic{
            GraphicData: &a_GraphicData{
                Uri: "http://schemas.openxmlformats.org/drawingml/2006/picture",
                Pic: &pic_Pic{
                    BlipFill: &pic_BlipFill{
                        Blip: &a_Blip{
                            Embed: rId, // relationship to image part
                        },
                    },
                    SpPr: &pic_SpPr{ /* shape properties */ },
                },
            },
        },
    }
    
    drawing := &CT_Drawing{Inline: inline}
    run := &Run{ct: &wml.CT_R{Drawing: drawing}, para: p}
    return run, nil
}
```

### Pattern 3: Header/Footer Creation — OPC Part + Relationship + Content Type + SectPr Link

**What:** Create a new header/footer part, add relationship, add content type override, append CT_HdrFtrRef to section properties.

**When to use:** `doc.AddHeader(variant)` and `doc.AddFooter(variant)`.

**Example:**
```go
// Source: D-10..D-13, OPC part creation pattern

type HeaderVariant int
const (
    HeaderDefault HeaderVariant = iota
    HeaderFirst
    HeaderEven
)

func (hv HeaderVariant) String() string {
    switch hv {
    case HeaderFirst: return "first"
    case HeaderEven:  return "even"
    default:          return "default"
    }
}

func (d *Document) AddHeader(variant HeaderVariant) *Header {
    partName := fmt.Sprintf("word/header%d.xml", d.nextHeaderID)
    d.nextHeaderID++
    
    hdr := &wml.CT_Hdr{P: []*wml.CT_P{{}}} // empty paragraph placeholder
    // ... user adds content via Header wrapper later
    
    // Marshal to XML
    var buf bytes.Buffer
    buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
    buf.WriteByte('\n')
    enc := xmlutil.NewEncoder(&buf)
    enc.Encode(hdr)
    enc.Flush()
    
    // Create part
    d.pkg.MarkModified(partName, buf.Bytes())
    
    // Add relationship
    rels := d.pkg.Rels["word/document.xml"]
    rId := rels.NextRID()
    rels.Rels = append(rels.Rels, opc.Relationship{
        ID: rId, Type: relHeader,
        Target: partName[len("word/"):], // relative to word/
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

### Pattern 4: List Numbering Def Auto-Generation

**What:** `AddList(ordered bool)` auto-generates CT_AbstractNum + CT_Num entries, appends them to the package's numbering.xml, and creates paragraphs with CT_NumPr references.

**When to use:** Every `AddList` or `AddListFromSlice` call.

**Example:**
```go
// Source: D-14..D-17, numbering types in internal/wml/numbering.go

func (d *Document) AddList(ordered bool) *ListBuilder {
    // 1. Parse existing numbering.xml or create empty
    nb := d.readOrCreateNumbering()
    
    // 2. Find max existing abstractNumId, increment
    nextAbsID := int64(0)
    for _, a := range nb.AbstractNum {
        if a.AbstractNumID != nil && *a.AbstractNumID >= nextAbsID {
            nextAbsID = *a.AbstractNumID + 1
        }
    }
    
    // 3. Create abstractNum with 9 levels
    abs := &wml.CT_AbstractNum{
        AbstractNumID: &nextAbsID,
        Lvl: makeLevels(ordered), // 9 levels: ordered→decimal, bulleted→bullet
    }
    nb.AbstractNum = append(nb.AbstractNum, abs)
    
    // 4. Create num entry pointing to abstractNum
    nextNumID := findMaxNumID(nb) + 1
    nb.Num = append(nb.Num, &wml.CT_Num{
        NumID: &nextNumID,
        AbstractNumID: &wml.CT_AbstractNumID{Val: &nextAbsID},
    })
    
    // 5. Write back to numbering.xml
    d.writeNumbering(nb)
    
    return &ListBuilder{
        doc:   d,
        numID: nextNumID,
    }
}
```

### Pattern 5: OpenTemplate Header/Footer Cloning (Reverse Phase 3 Pitfall 4)

**What:** OpenTemplate currently strips HdrFtrRef from source body and replaces with clean sectPr. Phase 5 reverses: clone header/footer parts from source OPC, fix up rIds, create parts + relationships in destination.

**When to use:** OpenTemplate / OpenTemplateReader when source has headers/footers.

**Example:**
```go
// Source: D-13, Phase 3 Pitfall 4 reversal

// In OpenTemplateReader, after CloneStyles:
// 1. Read source sectPr to find header/footer references
// 2. For each headerReference/footerReference in source sectPr:
//    a. Resolve rId to target part (e.g., "header1.xml")
//    b. Copy part bytes to dst via MarkModified
//    c. Add relationship in dst document.xml.rels (new rId)
//    d. Add content type override in dst
//    e. Remap HdrFtrRef.ID to new rId
// 3. Replace defaultSectPr() with source sectPr minus HdrFtrRef
//    (then re-add them with fixed rIds)
```

### Anti-Patterns to Avoid

- **Per-cell MarkModified calls:** Table construction sets dirty=true on Document, re-encodes body once at Save. Never call MarkModified per cell.
- **Mixing table builder methods with raw X() access:** All table operations go through TableBuilder/Row/Cell API. Raw X() escape hatch for advanced cases, but mixing patterns creates confusion.
- **Direct numbering.xml write without reading existing content:** Must parse existing numbering.xml (from cloned template), append new abstractNum/num entries, write merged result. Overwriting destroys template numbering defs.
- **Hardcoding rId strings:** Always use `rels.NextRID()`. Never hardcode "rId1", "rId2" etc.
- **Importing images as ITEM (not RELS):** Images must be relationships from document.xml.rels, not items in [Content_Types].xml alone. Word needs the rel to find the image bytes.
- **Header/footer template cloning without content type:** Must add override entry for each cloned header/footer part. Missing override → Word repair dialog or missing content.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| XML serialization of CT_Hdr/CT_Ftr/CT_Numbering | Manual XML construction | `xmlutil.NewEncoder` + `Encode` | Established pattern in create.go, template.go. Handles namespace prefix rewrite, xmlns injection. |
| OPC part modification tracking | Manual part replacement tracking | `opc.MarkModified(name, data)` | Phase 1 established pattern. Save picks up modified parts via part.data field. |
| Relationship ID allocation | Manual rId tracking | `opc.Relationships.NextRID()` | Phase 1 established monotonic high-water mark. No reuse of deleted IDs (OPC-06). |
| Content type registration | Raw map manipulation | `pkg.ContentTypes.Overrides["/path"] = ct` | Rendered via ContentTypes.serialize() at Save with canonical ordering. Direct map insertion is the API. |
| ZIP writing | Custom ZIP writer | `archive/zip` + `opc.Package.Save()` | OPC-02 canonical ordering, OPC-04 raw pass-through, OPC-06 relationship validation. |
| Image DPI detection from raw bytes | Homegrown JPEG/PNG DPI parser | `image/jpeg.DecodeConfig()` / `image/png.DecodeConfig()` | stdlib image package already handles format detection and pixel dimensions. Use `image/jpeg.DecodeConfig()` and `image/png.DecodeConfig()` for dimensions. For DPI, parse JPEG APP0/APP1 markers or PNG pHYs chunk manually (stdlib doesn't expose DPI). |
| DrawingML namespace handling | Manual namespace juggling | Use standard `xml.Name` tags with correct namespace URIs | Established WML pattern. All namespaces defined in internal/wml/namespaces.go. DrawingML namespaces: `http://schemas.openxmlformats.org/drawingml/2006/main` (a:), `http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing` (wp:), `http://schemas.openxmlformats.org/drawingml/2006/picture` (pic:). |
| Numbering def management | Manual CT_Numbering append logic | Read + append + write cycle | Must merge with existing numbering.xml from cloned templates to preserve template list defs. |
| Header/footer body content API | Separate content API for headers | Reuse Paragraph/Run API | Header wraps CT_Hdr (which has P []*CT_P). Paragraphs inside headers use same Paragraph type. |
| Page break paragraph creation | Insert raw CT_P with PageBreakBefore | `doc.AddPageBreak()` convenience | 2-line wrapper around AddParagraph + SetPageBreakBefore. Simpler user API. |

## Common Pitfalls

### Pitfall 1: CT_P Hyperlink + Drawing field not modeled before plan execution

**What goes wrong:** Plans assume CT_P can hold Hyperlink children and CT_R can hold Drawing children. Neither field exists yet. Every append panics or silently discards content.

**Why it happens:** These fields were identified as gaps in Phase 3/4 code context but not yet implemented. The planner creates tasks that reference non-existent struct fields.

**How to avoid:** Every plan's first task must add these fields to `internal/wml/document.go`:
```go
// Add to CT_P struct:
Hyperlink []*CT_Hyperlink `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hyperlink"`

// Add to CT_R struct:
Drawing *CT_Drawing `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main drawing"`
```

**Warning signs:** `unknown field "Hyperlink"` compile error when building tables+images plan.

### Pitfall 2: Table body insertion breaks paragraph order

**What goes wrong:** Body XML needs paragraphs and tables in document order (`<p/><tbl/><p/>`). Appending Tbl to Body.Tbl after all P entries places tables at end of document, not inline.

**Why it happens:** CT_Body.P and CT_Body.Tbl are separate slices. The user calls `doc.AddParagraph("before")`, `doc.AddTable(...)`, `doc.AddParagraph("after")`. Serialization writes all P entries then all Tbl entries, not interleaved.

**How to avoid:** Maintain an ordered body content slice (e.g., `BodyElements []BodyElement` where BodyElement is an interface or union type). At serialization time, iterate BodyElements in order emitting CT_P or CT_Tbl. Alternative: single slice of ordered elements with a type discriminator.

**Recommendation:** Create a `BodyElement` union type or use `[]interface{}` with type switch at serialization. This is a structural change to Document that affects Phase 4's AddParagraph (currently appends to Body.P directly).

**Warning signs:** Table always renders at end of document regardless of insertion position.

### Pitfall 3: Numbering def collision between auto-generated and template definitions

**What goes wrong:** Auto-generated list numbering defs use abstractNumId/numId starting at 0 or 1, colliding with template numbering defs that also start at low IDs.

**Why it happens:** `AddList` queries `findMaxNumID(nb) + 1` but template numbering.xml may have numId=0 (Word's ListNumber). If max ID logic is wrong, collision.

**How to avoid:** Scan ALL existing NumID and AbstractNumID in the parsed numbering.xml. Compute next available ID as `maxFound + 1`. Never hardcode starting IDs. Handle empty numbering.xml (no existing Num entries → start at 1, reserving 0 for Word's ListNumber).

**Warning signs:** List items show wrong numbering format (e.g., template numbered list mixed with auto-generated).

### Pitfall 4: Image DPI detection fails for encoded/formatted images

**What goes wrong:** JPEG from camera has EXIF orientation tag that inverts W/H. PNG from web may lack pHYs chunk (DPI metadata). Image renders at wrong size.

**Why it happens:** JPEG DPI detection requires parsing APP0 (JFIF) or APP1 (EXIF) markers. PNG DPI in pHYs chunk is optional. Web-optimized images often strip metadata.

**How to avoid:** Default to 72 DPI if no DPI metadata found. Provide `image.SetWidth(inches)` / `image.SetHeight(inches)` for explicit override (D-07). For JPEG EXIF orientation, parse orientation tag and swap W/H if needed — or document this as a known limitation in v1.

**Warning signs:** Image renders at unexpected size in Word. Very large (96 DPI interpreted as 72 DPI) or very small (no DPI data → defaults to 72 DPI).

### Pitfall 5: Header/footer relationship target is relative to word/

**What goes wrong:** Relationship target for header1.xml is set as "word/header1.xml" instead of "header1.xml". Word cannot resolve the part.

**Why it happens:** Relationship target paths are relative to the source part's directory. For `word/document.xml`, the source directory is `word/`, so target should be `header1.xml` (not `word/header1.xml`).

**How to avoid:** All relationships in `word/_rels/document.xml.rels` use targets relative to `word/`. Image targets: `media/logo.png`. Header targets: `header1.xml`. Footer targets: `footer1.xml`. For package root rels (`_rels/.rels`), targets are relative to package root, e.g. `word/document.xml`.

**Warning signs:** Word repair dialog on open, or header/footer not visible.

### Pitfall 6: OpenTemplate header cloning must fix up rIds

**What goes wrong:** Source template has headerReference with rId1 pointing to header1.xml. CloneStyles creates a fresh OPC package with its own rId1 pointing to styles.xml. The cloned headerReference's rId now resolves to the wrong part.

**Why it happens:** Fresh OPC packages created by `newTemplateTarget()` start rId allocation at 1 for styles/settings relationships. Copying source HdrFtrRef IDs directly creates dangling or wrong relationship references.

**How to avoid:** When cloning header/footer parts:
1. Parse source document.xml.rels for header/footer relationship entries
2. Copy part bytes and create in destination with MarkModified
3. Allocate NEW rIds in destination rels (via NextRID())
4. Create new CT_HdrFtrRef entries with new rIds in the destination sectPr

**Warning signs:** Headers show "Error! Reference source not found" or wrong content appears.

### Pitfall 7: Hyperlink relationship TargetMode must be "External"

**What goes wrong:** Hyperlink relationship targets are URLs. Without TargetMode="External", Word treats the URL as an internal part path and fails to resolve.

**Why it happens:** Internal targets (images, headers, styles) omit TargetMode. Hyperlinks to URLs require `TargetMode="External"`.

**How to avoid:** `Relationship{TargetMode: "External"}` for hyperlinks. Not required for internal targets.

**Warning signs:** Hyperlink doesn't open when clicked, or Word shows repair dialog.

## Code Examples

Verified patterns from existing codebase:

### Prior Art: defaultSectPr() Stub Pattern (for Section wrapper, D-18)
```go
// Source: create.go:201-211
func defaultSectPr() *wml.CT_SectPr {
    return &wml.CT_SectPr{
        PgSz:    &wml.CT_PgSz{W: ptrInt64(12240), H: ptrInt64(15840)},
        PgMar:   &wml.CT_PgMar{
            Top: ptrInt64(1440), Right: ptrInt64(1440),
            Bottom: ptrInt64(1440), Left: ptrInt64(1440),
        },
        Cols:    &wml.CT_Cols{},
        DocGrid: &wml.CT_DocGrid{},
    }
}
```

### Prior Art: OPC Part Creation + Relationship Pattern (for images/headers)
```go
// Source: create.go:48-97, packages Rels assembly pattern
func (d *Document) addRelationship(rel opc.Relationship) string {
    rels := d.pkg.Rels["word/document.xml"]
    rId := rels.NextRID()
    rel.ID = rId
    rels.Rels = append(rels.Rels, rel)
    return rId
}
```

### Prior Art: Body Re-Encode with MarkModified
```go
// Source: template.go:108-126, Phase 4 serializeBody pattern
func (d *Document) serializeBody() {
    if !d.dirty { return }
    var buf bytes.Buffer
    buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
    buf.WriteByte('\n')
    enc := xmlutil.NewEncoder(&buf)
    if err := enc.Encode(d.doc); err != nil { return }
    enc.Flush()
    d.pkg.MarkModified("word/document.xml", buf.Bytes())
    d.dirty = false
}
```

### Prior Art: NextRID Relationship Allocation (for hyperlinks, images)
```go
// Source: internal/opc/relationships.go:138-145
func (rs *Relationships) NextRID() string {
    if rs.next == 0 { rs.rescan() }
    id := "rId" + strconv.Itoa(rs.next)
    rs.next++
    return id
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| CT_P has no Hyperlink child | CT_P.Hyperlink added | Phase 5 | Enables w:hyperlink in paragraph content. Required for API-08. |
| CT_R has no Drawing child | CT_R.Drawing added | Phase 5 | Enables DrawingML inline/floating images. Required for API-05. |
| OpenTemplate strips HdrFtrRef entirely | OpenTemplate clones header/footer parts + fixes rIds | Phase 5 | Template headers/footers preserved when opened. Required for API-07. |
| No Table API | AddTable + AddTableBuilder + Table/Row/Cell wrappers | Phase 5 | API-04: Tables with borders, merge, shading. |
| No Image API | AddImage + AddImageBytes + DrawingML inline/anchor | Phase 5 | API-05: PNG/JPEG embedding with sizing. |
| No List API | AddList + AddListFromSlice + auto numbering defs | Phase 5 | API-06: Ordered/bulleted lists. |
| No Header/Footer API | AddHeader + AddFooter + Header/Footer wrappers | Phase 5 | API-07: All three header/footer variants. |
| No Hyperlink API | Paragraph.AddHyperlink + relationship auto-allocation | Phase 5 | API-08: Hyperlinks on runs. |
| No Page Setup API | SetOrientation, SetPaperSize, SetMargins, AddPageBreak | Phase 5 | API-09: Page setup convenience methods. |
| defaultSectPr() always letter/1in | SetOrientation/SetPaperSize/SetMargins mutate sectPr | Phase 5 | D-18..D-21: User-configurable page setup. |

**Deprecated/outdated:**
- None — all Phase 5 features are additive, no deprecations.

## Dependency Analysis: Plan Ordering

### Plan 05-01: Tables + Images
**Dependencies:**
- Phase 4 completed (Paragraph/Run API, builder chain, dirty flag serialization)
- **Must add:** CT_R.Drawing field + CT_Drawing types in `internal/wml/drawing.go` (new file)
- **Must add:** CT_Body ordered element tracking (body element union type) — OR accept tables-only-at-end limitation for v1
- **Must add:** Content type constants (ctPng, ctJpeg), relationship constant (relImage)
- Relationship management pattern (NextRID, MarkModified, content type overrides) already established

**Internal deps:** None within Phase 5. Tables and images are independent features.
**Risk:** LOW — WML types for tables already modeled. DrawingML types are straightforward.
**Gate:** Images add ~5 structs. Tables add ~0 structs (all exist).

### Plan 05-02: Headers/Footers + Page Setup
**Dependencies:**
- Phase 4 completed
- Plan 05-01 completed? **No — hdr/ftr and page setup are independent of tables/images**
- However: needs CT_HdrFtrRef, CT_Hdr, CT_Ftr, CT_SectPr — already modeled
- OpenTemplate header cloning depends on understanding source package rels

**Internal deps:** Independent of 05-01. Can be done in parallel with 05-01.
**Risk:** MEDIUM — OpenTemplate cloning reverses Phase 3 Pitfall 4, requires careful rId fixup.
**Gate:** Header/footer part creation pattern (MarkModified + rels + content type + sectPr link).

### Plan 05-03: Lists + Hyperlinks
**Dependencies:**
- Phase 4 completed
- Plan 05-01 completed? **NO for CT_P.Hyperlink field — must add before hyperlinks compile**
- YES for CT_R change — only applies to Drawing, not CT_P.Hyperlink
- Numbering types all modeled. CT_NumPr (numId, ilvl) exists.

**Internal deps:** Hyperlinks need CT_P.Hyperlink field added (can be done independently). Lists need nothing from 05-01 or 05-02.
**Risk:** LOW — numbering types all exist. Hyperlink type is ~2 structs.
**Gate:** CT_P.Hyperlink field addition. Numbering read+append+write cycle for auto-generation.

### Reordering Recommendation

**Current order (ROADMAP):** 05-01 → 05-02 → 05-03. This is correct.

- 05-01 is the largest (tables API + image embedding + new WML types)
- 05-02 is the riskiest (OpenTemplate reversal, part-level changes)
- 05-03 is the safest (smallest surface area, most types pre-existing)

**All three plans CAN be written independently** if the WML field additions (CT_R.Drawing, CT_P.Hyperlink) are extracted as a shared prerequisite. **Recommendation:** Add a Wave 0 to the first plan that adds all needed WML types and fields, unblocking all three plans:

**Wave 0** (prerequisite for all plans):
1. `internal/wml/drawing.go` — CT_Drawing, CT_Inline, CT_BlipFill, CT_Blip, CT_Extent, CT_NonVisualPicProps, CT_Anchor
2. `internal/wml/hyperlink.go` — CT_Hyperlink
3. Add `Drawing *CT_Drawing` field to CT_R in `document.go`
4. Add `Hyperlink []*CT_Hyperlink` field to CT_P in `document.go`
5. Add relationship/CT constants: relHeader, relFooter, relImage, relHyperlink, ctHeader, ctFooter, ctPng, ctJpeg
6. `wordingo/image.go` — EMU conversion constants + helper function

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Ordered body element tracking needed (tables inline with paragraphs, not all-at-end) | Architecture Patterns | If we defer ordered body tracking, tables always render at end of document. Acceptable for v1? D-24 says defer to Phase 6 or workaround. |
| A2 | JPEG DPI parsing from EXIF orientation swaps W/H | Common Pitfalls | Image renders at wrong aspect ratio for portrait photos from cameras. Mitigation: document limitation, provide explicit SetWidth/SetHeight override. |
| A3 | Header part file naming convention (header1.xml, header2.xml etc.) | Code Examples | Collision if OpenTemplate already cloned header1.xml from source. Solution: track nextHeaderID from existing parts. |
| A4 | CT_HdrFtrRef uses same Type values for both headers and footers | Architecture Patterns | HeaderReference Type=default/first/even. FooterReference uses same Type values. Confirmed by ISO 29500 §17.6.4. |
| A5 | List auto-numbering algorithm needs 0-index or 1-index starting id | Code Examples | Word reserves numId=0 for ListNumber style. Start auto-generated numIds at 1. AbstractNumIds start at 0. |

## Open Questions

1. **Ordered body element tracking** [RESOLVED: Defer to Phase 6]
   - What we know: CT_Body has separate P and Tbl slices. Interleaved tables require ordered tracking.
   - What's unclear: v1 scope — is it acceptable that tables always render after all paragraphs (at end of document)?
   - **Decision:** D-24 confirms ROADMAP order. Tables-at-end is v1 acceptable. Phase 6 adds ordered body tracking for interleaved tables.
   - **Recommendation:** Document the limitation clearly. Add a note to Phase 6 todo list.

2. **Image DPI detection library support** [RESOLVED]
   - What we know: Go stdlib `image/jpeg.DecodeConfig()` returns pixel dimensions, not DPI. JPEG DPI requires manual APP0/APP1 marker parsing. PNG DPI in pHYs chunk.
   - What's unclear: Should DPI detection be manual (parse JPEG markers, PNG pHYs) or default to 72 DPI and let user set explicit size?
   - **Decision:** Agent's discretion (D-07: auto-default from DPI). Recommend: implement JPEG JFIF APP0 marker parsing for DPI (standard 2-byte density units + H/V density). Implement PNG pHYs chunk parsing (4-byte pixels/unit, 1-byte unit specifier). Fall back to 72 DPI if no metadata. Document that DPI-less images use 72 DPI default.
   - **Recommendation:** Keep DPI parsing simple — handle common cases, warn if unusual. Provide explicit override.

3. **Header/footer content API** [RESOLVED]
   - What we know: Header/Footer types wrap CT_Hdr/CT_Ftr which have P []*CT_P.
   - What's unclear: Should Header expose AddParagraph() directly, or should user access Body.P and add CT_P directly via X()?
   - **Decision:** Header/Footer get AddParagraph() method mirroring Document.AddParagraph(). Underlying CT_Hdr.P / CT_Ftr.P slice. Paragraphs inside headers use same Paragraph+Run types with dirty tracking.

4. **Hyperlink relationship duplicate URI detection** [RESOLVED: Agent's discretion]
   - What we know: D-23 says auto rId, no user rId handling.
   - What's unclear: If same URI is used twice, should we reuse existing rId or create new one? D-23 says "auto rId" but doesn't specify dedup.
   - **Recommendation:** Create new rId per hyperlink call (simpler, no lookup cost). Dedup adds complexity for marginal benefit (one extra rId is ~50 bytes in XML). Agent's discretion per D-23.

## Environment Availability

> Step 2.6: SKIPPED (no external dependencies identified). This phase is pure Go code changes with no external tools, services, or runtimes beyond the Go toolchain already available.

## Validation Architecture

> SKIPPED: `workflow.nyquist_validation` is explicitly set to `false` in `.planning/config.json`.

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | Document library — no user auth |
| V3 Session Management | no | No sessions |
| V4 Access Control | no | Library exposes no access control |
| V5 Input Validation | yes | Image file path validation (AddImage), DPI parsing bounds, EMU conversion overflow check, URI validation for hyperlinks |
| V6 Cryptography | no | No encryption, no signing |

### Known Threat Patterns for Go + OOXML

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Image path traversal in AddImage(path) | Tampering | Go's os.Open with path validation. Document that AddImage is convenience; AddImageBytes is the safe path for untrusted file names. |
| EMU conversion integer overflow (inches * 914400) | Denial of Service | Use int64 arithmetic, guard against overflow. Reasonable max image size (e.g., 100 inches → 91,440,000 EMU < max int64). Low risk. |
| Hyperlink URI injection | Tampering | URI stored as relationship target. Word resolves hyperlinks at render time. No XML injection possible (Go encoding/xml handles escaping). |
| Large image decompression bomb | Denial of Service | Image bytes stored as OPC part — subject to MaxPartBytes cap (128MB). Decompression bomb protection inherited from OPC-07. |
| Numbering def ID exhaustion via repeated AddList | Denial of Service | Research shows no guard. Add over 2B lists to exhaust IDs. Negligible risk in practice. No guard needed for v1. |

## Sources

### Primary (HIGH confidence)
- Context7: `internal/wml/table.go` — CT_Tbl through CT_VMerge all modeled with correct XML tags, RawXML hoarding pattern. Ready for use.
- Context7: `internal/wml/document.go` — CT_Hdr, CT_Ftr, CT_HdrFtrRef, CT_SectPr (PgSz, PgMar, HdrFtrRef, FtrRef, TitlePg, Cols, DocGrid) all modeled. CT_P and CT_R lack Hyperlink/Drawing fields.
- Context7: `internal/wml/numbering.go` — CT_Numbering, CT_Num, CT_AbstractNum, CT_Lvl, CT_LvlOverride, CT_AbstractNumID, CT_NumFmt, CT_LvlText, CT_Start all modeled with correct XML tags.
- Context7: `internal/wml/properties.go` — CT_NumPr (numId, ilvl), CT_Shd all modeled. CT_NumPr: CT_ILvl+CT_NumId. CT_Shd: Val/Color/Fill/ThemeFill/ThemeColor.
- Context7: `internal/style/numbering.go` — numberingCache with ResolveLvl (numId+ilvl → CT_Lvl), lvlOverride support. Reusable pattern for list def resolution.
- Context7: `internal/opc/relationships.go` — Relationship struct, Relationships.Rels slice, NextRID() monotonic id allocation, relsPathFor/sourcePartFor helpers. serialize() renders XML directly (no encoding/xml).
- Context7: `internal/opc/contenttypes.go` — ContentTypes.Overrides map, TypeFor lookup, serialize with canonical ordering. Direct map insertion is the API.
- Context7: `internal/opc/package.go` — MarkModified (creates or replaces part). Save (modified parts from part.data, unmodified raw-copied). Part.Name, Part fields (file, data, modified, deleted).
- Context7: `create.go` — defaultSectPr(), newBlankPackage() pattern, content type/relationship constants. `relOfficeDocument`, `relStyles`, `relSettings`, etc.
- Context7: `template.go` — OpenTemplate strips HdrFtrRef (Phase 3 Pitfall 4). sourcePart reads, fresh doc with source P/Tbl + defaultSectPr(). Phase 5 reverses this.
- Context7: `wordingo.go` — Document struct (pkg, doc, dirty, warnings). AddParagraph appends to Body.P. serializeBody() re-encodes.

### Secondary (MEDIUM confidence)
- ISO/IEC 29500-1 §17.4.1 — Tables: CT_Tbl element model, CT_TblPr.TblStyle, CT_TblPr.TblW (w+type attributes), CT_TcPr.GridSpan (hMerge), CT_TcPr.VMerge (vMerge). CT_Tc contains P elements (paragraphs inside cells).
- ISO/IEC 29500-1 §17.6.3 — Header/footer part types (CT_Hdr, CT_Ftr), HdrFtrRef with type attribute (default/first/even), relationship between document.xml and header/footer parts.
- ISO/IEC 29500-1 §17.8.1 — Hyperlink element (w:hyperlink) with ID (rId) and child w:r elements. Existing at paragraph block level.
- ISO/IEC 29500-1 §17.13.2 — Page setup: CT_PgSz (w, h, code), CT_PgMar (top, right, bottom, left, header, footer, gutter), orientation by swapping w/h.
- ISO/IEC 29500-1 §17.3.2 — Numbering: CT_AbstractNum (abstractNumId, lvl[0..8]), CT_Num (numId, abstractNumId, lvlOverride), CT_Lvl (ilvl, numFmt, lvlText, start, pPr, rPr). 9-level depth.
- DrawingML inline: §20.4.2.8 (wp:inline) — CT_Inline with extent (cx, cy in EMU), docPr (id, name), graphic (a:graphic → a:graphicData → pic:pic → pic:blipFill → a:blip). Floating: §20.4.2.3 (wp:anchor).
- Image dimensions: 1 inch = 914400 EMU (ECMA-376 Part 1 §20.1.2.2). Standard OOXML unit for DrawingML extents.

### Tertiary (LOW confidence)
- JPEG DPI parsing: JFIF APP0 marker at offset 6 has density unit (0=no/1=inch/2=cm) and X/Y density as 2-byte big-endian. Wide variation in camera metadata. Fallback to 72 DPI is safe default.
- PNG pHYs chunk: 4 bytes pixels-per-unit-x, 4 bytes pixels-per-unit-y, 1 byte unit specifier (0=unknown, 1=meter). Unit=1 → DPI = pixels-per-meter / 39.37. Unit=0 or missing → assume 72 DPI.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all packages internal/existing. WML types verified. No external dependencies.
- Architecture: HIGH — patterns extend Phase 4 builder chain. OPC part creation pattern established. DrawingML types follow existing WML conventions.
- Pitfalls: MEDIUM — table body ordering and image DPI parsing are known unknowns with acceptable workarounds. Header rId fixup well-understood in principle.
- Plan decomposition: HIGH — three-plan breakdown is correct, independent with shared Wave 0 prerequisite.

**Research date:** 2026-07-26
**Valid until:** 2026-08-26 (stable codebase — no fast-moving dependencies)
