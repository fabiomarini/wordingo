---
phase: 05
plan: 02
subsystem: content
tags: header, footer, page-setup, section, orientation, margins, page-break, template-clone
requires:
  - phase: "05-01"
    provides: Tables + Images foundation, Document.nextHeaderID/nextFooterID counters
provides:
  - Header/Footer creation API (default, first, even variants) with OPC part, relationship, CT override, and sectPr linking
  - OpenTemplate header/footer part cloning with fresh rId fixup (reverses Phase 3 Pitfall 4)
  - Page setup API (orientation, paper size, margins, page break)
  - Section wrapper for forward-compatible multi-section v2
affects: 05-03 (lists/hyperlinks), 05-04 (formatting consolidation)
tech-stack:
  added: []
  patterns:
    - addHelperPart pattern for OPC part creation (rel + CT + part)
    - Section wrapper over CT_SectPr for forward-compatible API
    - Template header clone: copy part bytes, new rId, CT override, sectPr link fixup
key-files:
  created:
    - header.go — Header/Footer types, AddHeader/AddFooter, addHelperPart
    - page.go — Section type, Document Section/page/break methods
  modified:
    - format.go — HeaderVariant, FooterVariant, PageOrientation enums, PaperSize constants
    - template.go — OpenTemplateReader header/footer part cloning
    - paragraph.go — SetPageBreakBefore method
key-decisions:
  - "addHelperPart shared method for OPC part + rel + CT creation (pattern from image.go AddImageBytes)"
  - "Header/Footer AddParagraph mirrors Document.AddParagraph pattern"
  - "Section wrapper provides forward-compatible API for multi-section v2"
  - "Template header clone uses path.Join for target resolution, allocates fresh rIds via dstRels.NextRID()"
requirements-completed:
  - API-07
  - API-09
coverage:
  - id: D1
    description: "AddHeader(HeaderDefault) creates word/header1.xml part with relationship and sectPr link"
    verification:
      - kind: unit
        ref: "header.go#AddHeader"
        status: pass
    human_judgment: false
  - id: D2
    description: "AddFooter(FooterDefault) creates word/footer1.xml part with relationship and sectPr link"
    verification:
      - kind: unit
        ref: "header.go#AddFooter"
        status: pass
    human_judgment: false
  - id: D3
    description: "All three header/footer variants (default, first, even) supported"
    verification:
      - kind: unit
        ref: "format.go#HeaderVariant.String/FooterVariant.String"
        status: pass
    human_judgment: false
  - id: D4
    description: "Header/Footer types expose AddParagraph for content"
    verification:
      - kind: unit
        ref: "header.go#Header.AddParagraph/Footer.AddParagraph"
        status: pass
    human_judgment: false
  - id: D5
    description: "OpenTemplate clones template headers/footers with rId fixup"
    verification:
      - kind: unit
        ref: "template.go#OpenTemplateReader"
        status: pass
    human_judgment: false
  - id: D6
    description: "SetOrientation(Landscape) swaps PgSz W/H on default sectPr"
    verification:
      - kind: unit
        ref: "page.go#Section.SetOrientation"
        status: pass
    human_judgment: false
  - id: D7
    description: "SetPaperSize(PaperA4) sets correct twips dimensions"
    verification:
      - kind: unit
        ref: "page.go#Section.SetPaperSize"
        status: pass
    human_judgment: false
  - id: D8
    description: "SetMargins(t,r,b,l) sets PgMar fields"
    verification:
      - kind: unit
        ref: "page.go#Section.SetMargins"
        status: pass
    human_judgment: false
  - id: D9
    description: "AddPageBreak() creates paragraph with PageBreakBefore"
    verification:
      - kind: unit
        ref: "page.go#Document.AddPageBreak"
        status: pass
    human_judgment: false
  - id: D10
    description: "Section wrapper provides same API as Document-level methods"
    verification:
      - kind: unit
        ref: "page.go#Section"
        status: pass
    human_judgment: false
duration: 2 min
completed: 2026-07-26
status: complete
---

# Phase 5 Plan 2: Header/Footer + Page Setup Summary

**Header/Footer OPC part creation + sectPr linking + page setup API (orientation, paper, margins, page breaks, Section wrapper) + template header clone fixup**

## Performance

- **Duration:** 2 min
- **Started:** 2026-07-26T10:06:16Z
- **Completed:** 2026-07-26T10:07:46Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- HeaderVariant/FooterVariant enum types with String() -> "default"/"first"/"even"
- Header/Footer types wrapping CT_Hdr/CT_Ftr with AddParagraph() and X()
- Document.AddHeader(variant), Document.AddFooter(variant) — creates OPC part, relationship, content type override, sectPr link via HdrFtrRef/FtrRef
- addHelperPart reusable pattern for OPC part + relationship + content type creation
- OpenTemplateReader clones template header/footer parts with fresh rId allocation (reverse Phase 3 Pitfall 4)
- Template header/footer parts get content type overrides (T-05-06 mitigation)
- PageOrientation enum, PaperSize presets (Letter/A4/Legal) in twips
- Section type wrapping CT_SectPr with SetOrientation/SetPaperSize/SetMargins
- Document.Section() lazy-init accessor
- Document.SetOrientation/SetPaperSize/SetMargins chaining methods
- Document.AddPageBreak() — paragraph with PageBreakBefore
- Paragraph.SetPageBreakBefore(bool) for explicit control

## Task Commits

Each task was committed atomically:

1. **Task 1: Header/Footer API** - `91922ed` (feat)
2. **Task 2: Page Setup API + OpenTemplate header fixup** - `5a86ee6` (feat)

**Plan metadata:** (committed below with SUMMARY.md)

## Files Created/Modified

- `header.go` - Header/Footer types, AddHeader, AddFooter, addHelperPart
- `page.go` - Section type, Document Section/page/break methods
- `format.go` - HeaderVariant, FooterVariant, PageOrientation, PaperSize constants
- `template.go` - OpenTemplateReader header/footer part cloning
- `paragraph.go` - SetPageBreakBefore method

## Decisions Made

- addHelperPart shared method for OPC part + rel + CT creation (following image.go AddImageBytes pattern)
- Header/Footer AddParagraph mirrors Document.AddParagraph behavior
- Section wrapper provides forward-compatible API surface for unplanned multi-section v2
- Template header clone uses path.Join for target resolution from rel target
- Uses dstRels.NextRID() for fresh rId allocation (T-05-05 mitigation)
- Exposes ContentTypes.Overrides for cloned parts (T-05-06 mitigation)
- Page orientation swap logic via W/H comparison (T-05-07 accepted behavior — Word reads orientation from W/H ratio)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Header/footer API ready for use in document creation
- Page setup API ready for document-level formatting
- Template cloning includes header/footer parts with valid rIds
- Ready for 05-03 (Lists/Hyperlinks) or 05-04 (Formatting Consolidation)

---

*Phase: 05-rich-content*
*Completed: 2026-07-26*
