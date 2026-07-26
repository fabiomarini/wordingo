# Phase 5: Rich Content - Context

**Gathered:** 2026-07-26
**Status:** Ready for planning

<domain>
## Phase Boundary

Users create complete business documents programmatically: tables, images, headers/footers, lists, hyperlinks, page setup. Building on Phase 4's content API (paragraphs, runs, formatting, named styles). Six content domains, each needing its own public API surface.

Requirements: API-04..09. Template merge and edit operations deferred to Phase 6.

</domain>

<decisions>
## Implementation Decisions

### Table Builder API (API-04)
- **D-01:** Two entry points — `doc.AddTable(data [][]string)` for simple data grids + `doc.AddTableBuilder()` for complex tables (borders, merge, styles, cell formatting). Matches Phase 4 builder chaining pattern.
- **D-02:** Borders, shading, cell merge applied at row level with per-cell overrides. `table.Row(r).SetBorders(...)` bulk. `table.Row(r).Cell(c).SetShading(...)` for exceptions.
- **D-03:** Named table styles supported — `table.SetTableStyle("LightGrid-Accent1")`. Sets `tblStyle` on `TblPr`. Works with cloned template styles.
- **D-04:** Table width defaults to auto (Word content-based sizing). `table.SetWidth(w int64, wType string)` overrides. CT_TblW already modeled.

### Image Handling (API-05)
- **D-05:** Minimal DrawingML inline types modeled in `internal/wml/`: CT_Inline, CT_Blip, CT_BlipFill, CT_Extent, CT_NonVisualPicProps. ~5 structs covering PNG/JPEG embedding.
- **D-06:** Two entry points: `doc.AddImage(path)` and `doc.AddImageBytes(name, data, ct)`. File path for convenience, raw bytes for programmatic/network sources.
- **D-07:** Sizing auto-defaults from image DPI. Optional `SetWidth(inches)` / `SetHeight(inches)` overrides. Library converts inches to EMUs internally.
- **D-08:** Both inline (wp:inline) and floating (wp:anchor) placement supported in v1.
- **D-09:** Image creates media part in `word/media/`, adds relationship from `document.xml.rels`, adds content type override.

### Headers/Footers (API-07)
- **D-10:** One-shot creation — `doc.AddHeader(HeaderDefault)` creates header part, links via sectPr, returns `*Header` for content. Same pattern for footer.
- **D-11:** All three variants in v1: HeaderDefault, HeaderFirst (titlePg), HeaderEven (odd/even). CT_SectPr.TitlePg already modeled.
- **D-12:** Footer API symmetric with Header — `doc.AddFooter(FooterDefault)`. Both wrap same internal type (CT_Hdr and CT_Ftr have identical schema).
- **D-13:** OpenTemplate clones header/footer parts alongside styles (reversing Phase 3 Pitfall 4). Header/footer rIds fixed up for the fresh package.

### Lists (API-06)
- **D-14:** Two entry points: `doc.AddList(ordered bool)` builder + `doc.AddListFromSlice(items []string, ordered bool)` for flat lists.
- **D-15:** Numbering defs auto-generated internally. `doc.AddNumberingDef(fmt, start)` available for custom definitions (e.g., "Article I", "(a)").
- **D-16:** Single `AddList(ordered bool)` — `true` = ordered (decimal), `false` = bulleted. Multi-level via `AddItem("text", level)`.
- **D-17:** Full 9-level depth (0-8). Each level configurable via CT_Lvl (reuses Phase 2 existing type).

### Page Setup (API-09)
- **D-18:** Both doc-level convenience methods and Section object — `doc.SetOrientation(...)` delegates to `doc.Section().SetOrientation(...)`. Forward-compatible with multi-section v2.
- **D-19:** `doc.SetOrientation(OrientationLandscape | OrientationPortrait)` swaps W/H on default sectPr. `doc.SetPaperSize(w, h twips)` for explicit dimensions.
- **D-20:** `para.SetPageBreakBefore(true)` via CT_PPr.PageBreakBefore (already modeled). `doc.AddPageBreak()` convenience inserts paragraph with pageBreakBefore.
- **D-21:** Paper size presets (PaperLetter, PaperA4, PaperLegal) + `doc.SetPaperSize(w, h twips)` for custom. CT_PgSz.Code already modeled.

### Hyperlinks (API-08)
- **D-22:** `para.AddHyperlink(text, uri)` on Paragraph. Creates w:hyperlink element containing w:r with text. Matches OOXML structure (hyperlink at paragraph level, not run level).
- **D-23:** Auto rId — AddHyperlink creates relationship entry in `word/_rels/document.xml.rels`, assigns next available rId. User never touches rIds.

### Plan Breakdown
- Keep ROADMAP order: 05-01 Tables+images, 05-02 HdrFtr+sections+pageSetup, 05-03 Lists+hyperlinks

### the agent's Discretion
- Exact DrawingML struct field names for CT_Inline, CT_Blip etc.
- Table struct layout within `wordingo/` package (table.go)
- Image format detection and DPI parsing strategy
- Header/footer internal file layout
- List numbering def auto-generation algorithm (abstractNumId allocation)
- PageOrientation enum values
- Hyperlink relationship cleanup (duplicate URI detection)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Product & Requirements
- `.planning/REQUIREMENTS.md` — API-04..09 normative text (tables, images, headers/footers, lists, hyperlinks, page setup)
- `.planning/ROADMAP.md` — Phase 5 goal, success criteria, plan breakdown (05-01 Tables+images, 05-02 HdrFtr+pageSetup, 05-03 Lists+hyperlinks)
- `.planning/PROJECT.md` — Constraints (stdlib only, MIT, Go 1.23+), Key Decisions table

### Prior Phase Context (carried forward — must respect)
- `.planning/phases/04-content-api/04-CONTEXT.md` — D-01..D-14 (builder chaining pattern, Warnings pattern, X() escape hatch, SetStyle/SetAlignment/AddRun)
- `.planning/phases/03-document-model/03-CONTEXT.md` — D-04 (OpenTemplate strips HdrFtrRef — Pitfall 4 reversal needed in Phase 5)
- `.planning/phases/02-style-engine/02-CONTEXT.md` — D-13 (numbering resolution level), D-06 (theme color handler)
- `.planning/phases/01-foundation/01-CONTEXT.md` — D-04 (per-part byte diff round-trip), D-09..D-12 (error/warning patterns)

### External Specifications (no local copies — cite in plan as needed)
- ISO/IEC 29500 Part 1 — §17.4 (Tables), §17.6 (Headers/Footers), §17.8 (Hyperlinks), §17.13 (Page Setup), §17.3 (Numbering), §20 (DrawingML)
- ECMA-376 Part 2 (OPC) — part creation, relationship management, content types

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/wml/table.go` — CT_Tbl, CT_TblPr, CT_TblGrid, CT_Tr, CT_Tc, CT_TcPr (hMerge/vMerge, borders, shading, widths) all modeled and ready
- `internal/wml/document.go` — CT_Body.Tbl field already alongside P; CT_Hdr, CT_Ftr, CT_SectPr.HdrFtrRef, CT_PgSz, CT_PgMar, CT_SectPr.TitlePg all modeled
- `internal/wml/numbering.go` — CT_Numbering, CT_Num, CT_AbstractNum, CT_Lvl, CT_NumFmt, CT_LvlText, CT_Start — full list-backed numbering types
- `internal/wml/properties.go` — CT_NumPr (numId, ilvl), CT_Shd (reusable for table cells), CT_Jc, CT_Spacing
- `internal/wml/styles.go` — CT_Style.Type for table/paragraph/numbering styles
- `internal/style/numbering.go` — Phase 2 numbering resolver for list definition resolution
- `create.go` — `defaultSectPr()`, content type/relationship constants (ctMain, relStyles, relTheme, etc.)
- `template.go` — OpenTemplate strips HdrFtrRef (Phase 3 design); Phase 5 reverses this by cloning header/footer parts

### Established Patterns
- Wrapper-over-schema with `X()` escape hatch — every public type wraps WML CT_* and exposes it via X()
- Builder chaining — methods return receiver for fluent API
- Deferred errors via `Warnings()` — non-fatal issues accumulated, checked at Save
- Per-part byte diff for round-trip assertion
- CT_Body holds `P []*CT_P` for paragraphs AND `Tbl []*CT_Tbl` for tables — same body container

### Integration Points
- `wordingo.go` — `Document.Body` holds `Tbl []*CT_Tbl` alongside `P []*CT_P`; needs `AddTable()` alongside `AddParagraph()`
- `Document.Paragraphs()` returns paragraphs only — `Tables()` / `Table(i)` needed for table access
- Headers/footers need OPC part creation + relationship entries + content type overrides + sectPr reference assembly
- Images need new DrawingML types + media part creation + relationship from document.xml.rels + content type override
- Lists use existing numbering types — `doc.AddList` needs numbering def generation (auto abstractNumId allocation)
- CT_R currently lacks DrawingML inline field — needs `Drawing` field for inline images
- CT_P currently lacks Hyperlink child — needs `Hyperlink *CT_Hyperlink` field in CT_P
- OpenTemplate header handling: currently strips HdrFtrRef — Phase 5 must clone header/footer parts and fix rIds

</code_context>

<specifics>
## Specific Ideas

- AddTable(data [][]string) mirrors the simplicity of doc.AddParagraph("text") — two-line table creation for common case
- Images: EMU conversion helper needed (inches → EMU: inches * 914400)
- List numbering defs: auto-generate abstractNum with numId starting after max existing numId. Reuses Phase 2 CT_Numbering marshaling.
- Page break: CT_PPr.PageBreakBefore field exists but not exposed via Paragraph API yet
- Hyperlinks: CT_Hyperlink needs modeling with ID attr (rId) + child CT_R elements

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 05-rich-content*
*Context gathered: 2026-07-26*
