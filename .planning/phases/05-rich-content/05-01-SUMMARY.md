---
phase: 05-rich-content
plan: 01
subsystem: wml
tags: tables, images, drawingml, hyperlink, wml-types
requires:
  - phase: 04-content-api
    provides: CT_P, CT_R, CT_Tbl, Document body API
provides:
  - Table builder API (grid tables + builder chain)
  - Image embedding API (PNG/JPEG with DrawingML inline)
  - Shared WML types (CT_Drawing, CT_Hyperlink, CT_P.Hyperlink, CT_R.Drawing)
  - Content type and relationship constants for headers, footers, images, hyperlinks
affects:
  - 05-02 (Headers/Footers) — CT_Hyperlink, ctHeader/ctFooter, relHeader/relFooter
  - 05-03 (Lists/Hyperlinks/PageSetup) — CT_Hyperlink, CT_P.Hyperlink, relHyperlink

tech-stack:
  added: image/png (stdlib), image/jpeg (stdlib)
  patterns: Fluent builder chain (TableBuilder/RowBuilder/CellBuilder), DrawingML inline construction

key-files:
  created:
    - internal/wml/drawing.go — 25 DrawingML struct types
    - internal/wml/hyperlink.go — CT_Hyperlink type
    - table.go — TableBuilder, RowBuilder, CellBuilder, AddTable, AddTableBuilder
    - image.go — AddImage, AddImageBytes, EMU helpers, DPI detection, Run.SetImageWidth/SetImageHeight
  modified:
    - internal/wml/document.go — CT_P.Hyperlink, CT_R.Drawing fields
    - internal/wml/namespaces.go — NSPicture constant
    - internal/wml/table.go — CT_TcPr.Borders, CT_TcBorders
    - create.go — ctHeader/ctFooter/ctPng/ctJpeg, relHeader/relFooter/relImage/relHyperlink
    - wordingo.go — nextImageID/nextHeaderID/nextFooterID on Document
    - open.go — initialize counters in OpenReader
    - template.go — initialize counters in FromTemplateReader/OpenTemplateReader
    - format.go — TableBorders, BorderDef types
    - run.go — SetImageWidth, SetImageHeight methods

key-decisions:
  - "Tables append after all paragraphs (v1 limitation per D-24) — no ordered interleaving"
  - "AddImageBytes builds full DrawingML inline structure with stretch-to-fill and locked aspect ratio"
  - "Image auto-size at 3in default width with proportional height, overridable via SetImageWidth/SetImageHeight"
  - "Image DPI detection via JFIF APP0 (JPEG) and pHYs (PNG) chunks for accurate display size"
  - "CT_Anchor defined as placeholder with RawXML hoarding — anchored images deferred"
  - "nextImageID/nextHeaderID/nextFooterID init to 1 in all constructors (no scan of existing rels for v1)"

requirements-completed:
  - API-04
  - API-05

coverage:
  - id: D1
    description: "Shared WML types — CT_Drawing with 25 DrawingML structs, CT_Hyperlink, CT_P.Hyperlink, CT_R.Drawing"
    requirement: API-04
    verification:
      - kind: unit
        ref: "go build ./..."
        status: pass
      - kind: unit
        ref: "internal/wml/drawing.go#L1-L25 type declarations"
        status: pass
    human_judgment: false
  - id: D2
    description: "Content type and relationship constants for images, headers, footers, and hyperlinks"
    requirement: API-05
    verification:
      - kind: unit
        ref: "grep relImage create.go"
        status: pass
    human_judgment: false
  - id: D3
    description: "Table builder API — AddTable for grid tables, AddTableBuilder with fluent chain, Tables() accessor"
    requirement: API-04
    verification:
      - kind: unit
        ref: "TableBuilder chain test in TestTableBuilder"
        status: pass
    human_judgment: false
  - id: D4
    description: "Image embedding API — AddImageBytes creates media part, relationship, content type, and DrawingML inline element"
    requirement: API-05
    verification:
      - kind: unit
        ref: "TestImageEmbed — zip has word/media/, content type override, drawing markup"
        status: pass
    human_judgment: false
  - id: D5
    description: "Run.SetImageWidth/SetImageHeight — mutates DrawingML inline extent and pic SpPr extent"
    requirement: API-05
    verification:
      - kind: unit
        ref: "TestImageEmbed — Extent Cx/Cy after SetImageWidth(2.0)/SetImageHeight(2.0)"
        status: pass
    human_judgment: false
  - id: D6
    description: "Sequence counters (nextImageID/nextHeaderID/nextFooterID) initialized in all Document constructors"
    requirement: API-04
    verification:
      - kind: unit
        ref: "grep nextImageID wordingo.go open.go template.go"
        status: pass
    human_judgment: false

duration: 45 min
completed: 2026-07-26
status: complete
---

# Phase 5 Plan 1: Tables + Images Summary

**Table builder API (grid + fluent chain) and image embedding API (PNG/JPEG with full DrawingML inline structure), including shared WML types and constants that unblock all Phase 5 plans**

## Performance

- **Duration:** 45 min
- **Started:** 2026-07-26T10:20:00Z
- **Completed:** 2026-07-26T11:05:00Z
- **Tasks:** 3
- **Files modified:** 13

## Accomplishments

- **Shared WML types (Task 1):** 25 DrawingML struct types (CT_Drawing, CT_Inline, CT_Pic, CT_BlipFill, CT_Graphic, etc.), CT_Hyperlink with r:id and child runs, CT_P.Hyperlink and CT_R.Drawing fields, namespace constants (NSPicture), content type constants (ctPng, ctJpeg, ctHeader, ctFooter), relationship constants (relHeader, relFooter, relImage, relHyperlink), and nextImageID/nextHeaderID/nextFooterID sequence counters on Document.

- **Table builder API (Task 2):** Simple grid table via `AddTable([][]string)`, complex tables via `AddTableBuilder()` with fluent chain (`SetTableStyle → Row().Cell().SetText().SetBold()`), `RowBuilder.SetBorders()` on all cells, `CellBuilder` with shading, bold, width, merge (MergeRight/MergeDown), and `Tables()` accessor. Tables append at end of body (v1 per D-24).

- **Image embedding API (Task 3):** `AddImage(path)` and `AddImageBytes(name, data, ct)` create word/media/ part, relationship from document.xml, content type override, and full DrawingML inline element (wp:inline → a:graphic → pic:pic → a:blipFill → a:blip with r:embed). DPI detection from JFIF APP0 (JPEG) and pHYs (PNG) chunks. Auto-size at 3in default width. `Run.SetImageWidth(inches)` and `Run.SetImageHeight(inches)` mutate the inline extent.

## Task Commits

Each task was committed atomically:

1. **Task 1: Shared WML types, fields, and constants** — `52fa88e` (feat)
2. **Task 2: Table builder API** — `3c27139` (feat)
3. **Task 3: Image embedding API** — `fa06426` (feat)

**Plan metadata:** (pending metadata commit)

## Files Created/Modified

### Created (4)
- `internal/wml/drawing.go` — 25 DrawingML struct types (CT_Drawing, CT_Inline, CT_Pic, CT_BlipFill, CT_Graphic, etc.)
- `internal/wml/hyperlink.go` — CT_Hyperlink with r:id attribute and child runs
- `table.go` — TableBuilder, RowBuilder, CellBuilder with fluent chain API
- `image.go` — AddImage, AddImageBytes, EMU conversion, DPI detection

### Modified (9)
- `internal/wml/document.go` — CT_P.Hyperlink field added, CT_R.Drawing field added
- `internal/wml/namespaces.go` — NSPicture constant
- `internal/wml/table.go` — CT_TcPr.Borders, CT_TcBorders type
- `create.go` — ctHeader/ctFooter/ctPng/ctJpeg, relHeader/relFooter/relImage/relHyperlink
- `wordingo.go` — Document.nextImageID/nextHeaderID/nextFooterID
- `open.go` — initialize counters in OpenReader
- `template.go` — initialize counters in FromTemplateReader/OpenTemplateReader
- `format.go` — TableBorders, BorderDef types
- `run.go` — SetImageWidth, SetImageHeight methods

## Decisions Made

- **Tables-at-end v1:** Tables always render after all paragraphs per D-24. No ordered interleaving in v1. Future phases can add paragraph-index-based placement.
- **Auto-size at 3in:** Images default to 3in width with proportional height, overridable via SetImageWidth/SetImageHeight. DPI-aware sizing when JFIF/pHYs metadata available.
- **CT_Anchor placeholder:** CT_Anchor defined with RawXML hoarding for future anchored image support. Current implementation uses inline images only.
- **Counter initialization:** nextImageID/nextHeaderID/nextFooterID start at 1 in all constructors. No scan of existing package relationships in v1 — advanced users can set d.nextImageID manually before AddImageBytes.

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

- `boolPtr` helper in image.go conflicted with existing `boolPtr` in format_test.go (test package uses `package wordingo`, not `_test`). Renamed to `ptrBool` to resolve.

## Known Stubs

None

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: filesystem-read | image.go | `AddImage(path)` reads arbitrary file via `os.ReadFile(path)` — mitigated by Go stdlib path handling |
| threat_flag: part-creation | image.go | `AddImageBytes` creates new OPC parts and relationships — mitigated by OPC-07 MaxPartBytes cap |

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- **Shared WML types** (CT_Drawing, CT_Hyperlink, CT_P.Hyperlink, CT_R.Drawing) available for plans 05-02 and 05-03
- **Content type and relationship constants** for headers, footers, images, and hyperlinks deployed in create.go
- **Sequence counters** (nextImageID, nextHeaderID, nextFooterID) initialized on Document in all constructors
- Ready for **05-02 (Headers/Footers)**, **05-03 (Lists/Hyperlinks/PageSetup)**

---

*Phase: 05-rich-content*
*Completed: 2026-07-26*
