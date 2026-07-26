# Phase 5: Rich Content - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-07-26
**Phase:** 05-rich-content
**Areas discussed:** Table builder API, Image handling depth, Header/footer part API, List API + numbering, Page setup model, Hyperlink model, Plan breakdown order

---

## Table builder API

| Option | Description | Selected |
|--------|-------------|----------|
| Data-driven + builder | AddTable(data) for simple grids + AddTableBuilder() for complex | ✓ |
| Row-centric fluent | doc.AddTable(rows, cols).Row(0).Cell(0).AddParagraph() | |
| Cell-centric builder | table.Cell(r,c) returns *Cell with AddParagraph/SetShading | |

| Option | Description | Selected |
|--------|-------------|----------|
| Per-cell via Cell() | table.Row(r).Cell(c).SetShading().SetBorder() | |
| Row-level bulk | Row-level shorthands + per-cell overrides | ✓ |
| Format struct bulk | CellFormat struct like RunFormat/ParFormat | |

| Option | Description | Selected |
|--------|-------------|----------|
| Yes, SetTableStyle(name) | Sets tblStyle on TblPr | ✓ |
| Defer to v2 | Skip named table styles | |
| Auto from template | Inherit from template styles | |

| Option | Description | Selected |
|--------|-------------|----------|
| Explicit preferred | Default to 100% page width | |
| Auto by default | Word content-based sizing | |
| Both settable | Auto default, SetWidth(w, type) available | ✓ |

---

## Image handling depth

| Option | Description | Selected |
|--------|-------------|----------|
| Minimal inline types | CT_Inline, CT_Blip, CT_BlipFill, CT_Extent, CT_NonVisualPicProps | ✓ |
| Opaque RawXML | User provides DrawingML XML string | |
| Full DrawingML | All common DrawingML types | |

| Option | Description | Selected |
|--------|-------------|----------|
| Embed from file | doc.AddImage(path) | |
| Embed from bytes | doc.AddImageBytes(name, data, ct) | |
| Both | Both entry points | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| Explicit EMU sizes | User provides EMUs | |
| Float inches/cm | User provides inches, lib converts | |
| Auto + explicit | Auto from DPI, SetWidth/SetHeight override | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| Inline only (v1) | wp:inline only | |
| Inline + anchor | wp:inline + wp:anchor | ✓ |

---

## Header/footer part API

| Option | Description | Selected |
|--------|-------------|----------|
| doc.NewHeader() + doc.SetHeader() | Two-step creation + assignment | |
| doc.AddHeader() one-shot | Creates part, links via sectPr | ✓ |
| Section-based | doc.Section().SetHeader() | |

| Option | Description | Selected |
|--------|-------------|----------|
| All three in v1 | Default, First, Even | ✓ |
| Default only (v1) | HeaderDefault only | |
| Default + First | Most common combo | |

| Option | Description | Selected |
|--------|-------------|----------|
| Symmetric | doc.AddFooter() mirrors AddHeader | ✓ |
| Separate *Footer type | Distinct type for CT_Ftr | |

| Option | Description | Selected |
|--------|-------------|----------|
| Clone header parts | Clone header/footer parts with rId fixup | ✓ |
| Strip like Phase 3 | Keep Phase 3 behavior | |
| Preserve but fix rIds | Clone, fix rIds | |

---

## List API + numbering

| Option | Description | Selected |
|--------|-------------|----------|
| doc.AddList() builder | Builder with AddItem for multi-level | ✓ |
| doc.AddListItem(text, level) | Simple per-item add | |
| List from slice | doc.AddListFromSlice(items, ordered) | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| Auto-generate | Numbering defs created internally | |
| Manual + auto | Both auto and AddNumberingDef() | ✓ |
| Template-only | Reference existing template numbering | |

| Option | Description | Selected |
|--------|-------------|----------|
| AddOrderedList() / AddBulletList() | Separate methods | |
| AddList(ordered bool) | Single method with flag | ✓ |
| Style-based | Inherit from paragraph style | |

| Option | Description | Selected |
|--------|-------------|----------|
| Unlimited (9 levels) | Full OOXML support | ✓ |
| 3 levels (v1) | Limit to 3 | |
| 5 levels | Compromise | |

---

## Page setup model

| Option | Description | Selected |
|--------|-------------|----------|
| Document-level setters | doc.SetPageMargins().SetOrientation() | |
| Section object | doc.Section().SetMargins() | |
| Both | Doc-level convenience + Section object | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| SetOrientation(orient) | Enum-based, library swaps W/H | |
| SetPaperSize(w,h) only | No orientation enum | |
| SetOrientation + SetPaperSize | Both available | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| para.SetPageBreakBefore() | Existing CT_PPr field | |
| doc.AddPageBreak() | Convenience method | |
| Both | Both paths | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| Named constants | PaperLetter, PaperA4, PaperLegal | |
| Custom only | Always explicit WxH | |
| Presets + custom | Both paths | ✓ |

---

## Hyperlink model

| Option | Description | Selected |
|--------|-------------|----------|
| para.AddHyperlink(text, uri) | On Paragraph, w:hyperlink element | ✓ |
| run.SetHyperlink(uri) | On Run, wraps in hyperlink | |
| Both convenience | Both paths | |

| Option | Description | Selected |
|--------|-------------|----------|
| Auto rId | Creates relationship, assigns rId | ✓ |
| Manual rId | User provides rId | |

---

## Plan breakdown order

| Option | Description | Selected |
|--------|-------------|----------|
| Keep as-is | 05-01 Tables+images, 05-02 HdrFtr+pageSetup, 05-03 Lists+hyperlinks | ✓ |
| Lists first | Start with lists (reuse Phase 2 resolver) | |
| Headers first | Fix Pitfall 4 first | |

---

## the agent's Discretion

- Exact DrawingML struct field names for CT_Inline, CT_Blip etc.
- Table struct layout within `wordingo/` package
- Image format detection and DPI parsing strategy
- Header/footer internal file layout
- List numbering def auto-generation algorithm
- PageOrientation enum values
- Hyperlink relationship cleanup

## Deferred Ideas

None — discussion stayed within phase scope.
