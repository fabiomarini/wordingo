---
phase: 05-rich-content
plan: 03
subsystem: content-api
tags: [lists, numbering, hyperlinks, wordprocessingml, go]
requires:
  - phase: 05-01
    provides: CT_Hyperlink type, CT_P.Hyperlink field, relHyperlink constant
provides:
  - List builder API: AddList, AddListFromSlice, AddNumberingDef, AddItem
  - Hyperlink API: Paragraph.AddHyperlink with rId + External relationship
  - Auto-generated numbering definitions with 9-level depth
  - Template numbering preservation (scan+merge, no ID collision)
affects: []
tech-stack:
  added: []
  patterns:
    - Numbering read+append+write cycle via readOrCreateNumbering / writeNumbering
    - ListBuilder fluent builder (AddItem chaining, X() escape hatch)
    - Hyperlink part creation through OPC relationship layer
key-files:
  created:
    - list.go
    - hyperlink.go
    - list_test.go
  modified: []
key-decisions:
  - "No dedup on hyperlink URIs — each AddHyperlink call creates new rId (~50 bytes per extra relationship, negligible)"
  - "numId=0 reserved for Word's ListNumber (Pitfall 3); auto-generated numIds start at 1"
  - "AbstractNumId numbering starts at 0 (no reserved range needed)"
  - "Bullet chars cycle through 9 distinct Unicode glyphs for levels 0-8"
requirements-completed:
  - API-06
  - API-08
coverage:
  - id: D1
    description: Ordered/bulleted list creation with auto-generated numbering definitions
    requirement: API-06
    verification:
      - kind: unit
        ref: list_test.go#TestList_Ordered
        status: pass
      - kind: unit
        ref: list_test.go#TestList_Bulleted
        status: pass
      - kind: unit
        ref: list_test.go#TestList_MultiLevel
        status: pass
      - kind: unit
        ref: list_test.go#TestList_AddListFromSlice
        status: pass
      - kind: unit
        ref: list_test.go#TestList_AddNumberingDef
        status: pass
      - kind: unit
        ref: list_test.go#TestList_TemplateNumberingPreserved
        status: pass
    human_judgment: false
  - id: D2
    description: Hyperlink API on paragraphs with External relationship and Run chaining
    requirement: API-08
    verification:
      - kind: unit
        ref: list_test.go#TestHyperlink_Basic
        status: pass
      - kind: unit
        ref: list_test.go#TestHyperlink_FormattingChaining
        status: pass
      - kind: unit
        ref: list_test.go#TestHyperlink_MultipleCalls
        status: pass
    human_judgment: false
duration: 1 min
completed: 2026-07-26
status: complete
---

# Phase 5 Plan 3: Lists+Hyperlinks Summary

**List builder API (ordered/bulleted with auto-generated numbering defs) and hyperlink API on paragraphs (rId allocation + External relationships)**

## Performance

- **Duration:** 1 min
- **Started:** 2026-07-26T10:12:01Z
- **Completed:** 2026-07-26T10:13:28Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- `doc.AddList(ordered)` creates auto-numbered list with full 9-level abstractNum + num, returns ListBuilder
- `doc.AddListFromSlice(items, ordered)` flat list convenience method
- `doc.AddNumberingDef(fmt, start)` custom numbering definitions
- `ListBuilder.AddItem(text, level)` creates paragraph with NumPr linking (numId + ilvl)
- Ordered: `numFmt=decimal`, `lvlText="%N."` for each level N+1
- Bulleted: `numFmt=bullet`, distinct bullet chars per level (•, ◦, ▪, ▫, •, ◦, ▪, ▫, •)
- Numbering read+append+write cycle preserves template numbering (collision-free ID allocation)
- `para.AddHyperlink(text, uri)` creates CT_Hyperlink with External relationship via NextRID()
- Returned Run chains formatting: `p.AddHyperlink("click", "url").SetBold(true).SetColor("0563C1")`
- Hyperlink no-dedup policy: new rId per call (~50 bytes/extra relationship)

## Task Commits

Each task was committed atomically:

1. **Task 1: List builder API** - `03c668b` (feat)
2. **Task 2: Hyperlink API** - `b17aa16` (feat)
3. **Tests** - `8f217e6` (test — list + hyperlink verification)

**Plan metadata:** (committed below as docs commit)

## Files Created/Modified

- `list.go` - ListBuilder type, Document.AddList/AddListFromSlice/AddNumberingDef, numbering read/write helpers
- `hyperlink.go` - Paragraph.AddHyperlink method with rId/relationship management
- `list_test.go` - 10 test cases covering lists and hyperlinks

## Decisions Made

- **No hyperlink URI dedup** — each AddHyperlink call creates a new rId. Extra relationships are ~50 bytes each, negligible cost for typical documents. D-23 agent discretion applied.
- **numId=0 reserved** for Word's built-in ListNumber definition (Pitfall 3). Auto-generated numIds start at 1.
- **AbstractNumId starts at 0** — no reserved range needed for abstractNum IDs.
- **Bullet chars cycle** through 9 distinct Unicode glyphs for visual differentiation across levels 0-8.
- **Ordered level text** uses `%N.` pattern where N = level index + 1 (matches Word's default).
- **Numbering ID allocation** scans ALL existing entries in numbering.xml to avoid collisions with template numbering (Pitfall 3 mitigation).

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- xmlutil.Encoder uses full element syntax (`<w:ilvl w:val="0"></w:ilvl>` not self-closing `<w:ilvl w:val="0"/>`), requiring relaxed string matching in tests. No functional impact.

## Next Phase Readiness

- List and hyperlink APIs complete
- Ready for 05-04 (Formatting Consolidation) or parallel rich-content plans
- Template numbering merge verified — existing templates with numbering.xml are preserved

---
*Phase: 05-rich-content*
*Completed: 2026-07-26*
