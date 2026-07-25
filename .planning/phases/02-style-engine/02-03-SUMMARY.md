---
phase: 02-style-engine
plan: 03
subsystem: style engine
tags: opc, ooxml, clone, styles, numbering, fontTable, theme, settings
requires:
  - phase: 01-foundation
    provides: opc.Package, MarkModified, NextRID, ContentTypes
provides:
  - CloneStyles(src, dst *opc.Package) error — byte pass-through clone of the 5 style dependency-graph parts
  - ErrCloneTargetNotEmpty sentinel (D-08 fresh-empty-target policy)
  - cloneParts constant table (5 entries with OOXML relationship type URIs + content-type MIMEs)
affects: 02-04 (corpus validation), Phase 3 (Document.FromTemplate)
tech-stack:
  added: none (stdlib + internal/opc only)
  patterns: byte pass-through clone via MarkModified (D-09), precondition scan atomicity (D-08), leading-slash content-type Override key
key-files:
  created:
    - internal/style/cloner.go
    - internal/style/cloner_test.go
  modified:
    - internal/style/errors.go
key-decisions:
  - "Relationship Target is relative to word/document.xml's directory (strip 'word/' prefix) — opc.resolveTarget joins it with source part directory"
  - "ErrCloneTargetNotEmpty also used for nil-package defensive guard (wraps clear error)"
  - "No strings import needed in cloner apart from strings.TrimPrefix for rel target"
patterns-established:
  - "Clone precondition: atomic full scan before any byte copy (D-08)"
  - "Relationship target relative to source part dir, not package-absolute"
  - "struct{} constant table with name, relType, ct fields"
requirements-completed: [STYLE-CLONE-01]
coverage:
  - id: D1
    description: "CloneStyles byte-copies all 5 source parts into fresh-empty target with byte identity"
    requirement: "STYLE-CLONE-01"
    verification:
      - kind: unit
        ref: "internal/style/cloner_test.go#TestCloneStyles_FreshEmptyTarget"
        status: pass
      - kind: unit
        ref: "internal/style/cloner_test.go#TestCloneStyles_SaveRoundTrip"
        status: pass
    human_judgment: false
  - id: D2
    description: "ErrCloneTargetNotEmpty returned for non-empty target with atomicity (no partial clone)"
    requirement: "STYLE-CLONE-01"
    verification:
      - kind: unit
        ref: "internal/style/cloner_test.go#TestCloneStyles_ErrCloneTargetNotEmpty"
        status: pass
    human_judgment: false
  - id: D3
    description: "Absent source parts skipped silently; numbering.xml absence not an error"
    requirement: "STYLE-CLONE-01"
    verification:
      - kind: unit
        ref: "internal/style/cloner_test.go#TestCloneStyles_AbsentSourcePartSkip"
        status: pass
    human_judgment: false
  - id: D4
    description: "Relationships allocated with unique rIds via NextRID (OPC-06)"
    requirement: "STYLE-CLONE-01"
    verification:
      - kind: unit
        ref: "internal/style/cloner_test.go#TestCloneStyles_RelationshipAllocation"
        status: pass
    human_judgment: false
  - id: D5
    description: "Content-type Overrides registered with leading-slash keys (PATTERNS critical flag)"
    requirement: "STYLE-CLONE-01"
    verification:
      - kind: unit
        ref: "internal/style/cloner_test.go#TestCloneStyles_ContentTypeOverrides"
        status: pass
    human_judgment: false
  - id: D6
    description: "No xmlutil/wml imports in cloner.go (D-09 byte pass-through only)"
    requirement: "STYLE-CLONE-01"
    verification:
      - kind: unit
        ref: "internal/style/cloner_test.go#TestCloneStyles_NoReParse"
        status: pass
    human_judgment: false
  - id: D7
    description: "Second call to CloneStyles on same dst fails with ErrCloneTargetNotEmpty (idempotent)"
    requirement: "STYLE-CLONE-01"
    verification:
      - kind: unit
        ref: "internal/style/cloner_test.go#TestCloneStyles_IdempotentFails"
        status: pass
    human_judgment: false
  - id: D8
    description: "Nil package arguments produce clear error (no panic)"
    requirement: "STYLE-CLONE-01"
    verification:
      - kind: unit
        ref: "internal/style/cloner_test.go#TestCloneStyles_NilPackageDefensive"
        status: pass
    human_judgment: false
duration: 2 min
completed: 2026-07-25
status: complete
---

# Phase 02 Plan 03: Cloner — Byte Pass-Through Style Dependency Graph Clone

**CloneStyles with cloneParts constant table, ErrCloneTargetNotEmpty sentinel, and synthetic-fixture test suite — all 5 OOXML style parts (styles, numbering, fontTable, theme, settings) copied via MarkModified pass-through with valid relationships and content-type Overrides**

## Performance

- **Duration:** 2 min
- **Started:** 2026-07-25T18:52:37Z
- **Completed:** 2026-07-25T18:55:17Z
- **Tasks:** 1 (TDD: RED + GREEN)
- **Files modified:** 2 created, 1 modified

## Accomplishments

- `ErrCloneTargetNotEmpty` sentinel added to `internal/style/errors.go` (D-08 fresh-empty-target policy; 4th and final v1 style-package sentinel)
- `cloneParts` constant table with 5 entries: canonical OOXML relationship type URIs and content-type MIMEs for styles.xml, numbering.xml, fontTable.xml, theme/theme1.xml, settings.xml
- `CloneStyles(src, dst *opc.Package) error` — atomic precondition scan (D-08), byte pass-through via MarkModified (D-09), content-type Overrides with leading-slash key (PATTERNS critical flag), relationship allocation via NextRID (OPC-06)
- No xmlutil/wml imports — structurally enforced (D-09)
- Relationship.Target is relative to `word/document.xml`'s directory (opc.resolveTarget joins with source part dir)
- Comprehensive test suite: fresh-empty-target byte identity, ErrCloneTargetNotEmpty atomicity, absent-source-part skip, relationship allocation, content-type overrides, save round-trip, no-re-parse structural gate, idempotent-fails, nil-package defensive

## Task Commits

1. **Task 1 (RED): Add failing test + sentinel** — `f0ab3ed` (test: add failing test for CloneStyles + ErrCloneTargetNotEmpty)
2. **Task 1 (GREEN): Implement CloneStyles** — `646fd7a` (feat: implement CloneStyles with cloneParts constant table)

## Files Created/Modified

- `internal/style/cloner.go` — CloneStyles implementation + cloneParts constant table (183 lines, 0 xmlutil/wml references)
- `internal/style/cloner_test.go` — Comprehensive test suite covering all acceptance criteria (523 lines)
- `internal/style/errors.go` — Extended with `ErrCloneTargetNotEmpty` sentinel (additive var, no existing field changes)

## Decisions Made

- **Relationship Target is relative** to the source part's directory, not package-absolute. The OPC `resolveTarget` joins the target with the source part's directory (path.Dir); using `word/styles.xml` as the target would wrongly resolve to `word/word/styles.xml`. Fixed by stripping `word/` prefix via `strings.TrimPrefix`.
- **ErrCloneTargetNotEmpty for nil-guard**: Used as the wrapping sentinel for nil src/dst errors (consistent wrapping pattern).
- **No REFACTOR phase**: Implementation was minimal and clean after test feedback — no refactoring needed.

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

- **Relationship Target relative path**: The initial implementation used the full package-relative path (`word/styles.xml`) as the Relationship Target, but `opc.resolveTarget` joins the target with the source part's directory (`word/`), producing `word/word/styles.xml`. Fixed by stripping the `word/` prefix for relationship targets. All 5 parts are under `word/` so the stripping is deterministic. The plan's pseudocode (`Target: p.name`) under-specified this detail against the OPC relative-path convention.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- CloneStyles ready for 02-04 (corpus validation — Wave 3, depends on this plan + 02-01 + 02-02)
- `ErrCloneTargetNotEmpty` completes the 4-sentinel set (ErrStylesParseFailed, ErrThemeParseFailed, ErrNumberingParseFailed, ErrCloneTargetNotEmpty)
- Phase 3 `Document.FromTemplate()` will call CloneStyles under the hood

---
*Phase: 02-style-engine*
*Completed: 2026-07-25*
