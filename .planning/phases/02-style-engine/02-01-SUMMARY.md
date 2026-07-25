---
phase: 02-style-engine
plan: 01
subsystem: style-engine
tags: [style, resolver, basedOn-chain, memo, cycle-detection]
requires: []
provides:
  - Part.IsModified() accessor for cache invalidation (pre-existing)
  - ErrStylesParseFailed sentinel error
  - Resolver type with NewResolver, ResolveParagraph, ResolveRun, Warnings
  - Merged clone output through docDefaults → basedOn chain → direct formatting
  - Cycle detection (visited-set), latentStyles fallback, dangling-ref warnings
  - Synthetic test suite covering 14 behavior cases
affects:
  - 02-02 (theme color concretization, numbering level pPr merge)
  - 02-03 (real-fixture harness)

tech-stack:
  added: []
  patterns:
    - Pointer-field nil-merge semantics (shallow override + composite-leaf deep merge)
    - Warnings() []string accumulator (Phase 1 D-12 pattern)
    - Memo-by-styleId with cache invalidation via Part.IsModified
    - Clone-before-return (merge into empty CT_PPr/CT_RPr)

key-files:
  created:
    - internal/style/errors.go
    - internal/style/resolver.go
    - internal/style/resolver_test.go
  modified: []

key-decisions:
  - "Task 1 (Part.IsModified accessor) pre-existing — no code change needed"
  - "Resolver stores memo cache including docDefaults; cache invalidated via Part.IsModified dirty flag (Option A)"
  - "mergeColor preserves ThemeColor/ThemeShade/ThemeTint attributes for plan 02-02 concretization"
  - "NumPr passed through as shallow override — no lvl.PPr merge in this plan"

patterns-established:
  - "Deep merge for CT_Spacing, CT_Ind, CT_RFonts, CT_Color (per-attribute independence)"
  - "Shallow override for all other pointer fields on CT_PPr/CT_RPr"
  - "collectChain order: child-first return, root-first apply (reverse iteration)"

requirements-completed: [STYLE-RESOLVE-01, STYLE-RESOLVE-02, STYLE-RESOLVE-03]

coverage:
  - id: D1
    description: "Part.IsModified() accessor for cache invalidation"
    verification:
      - kind: unit
        ref: "internal/opc/package.go#func (p *Part) IsModified() bool"
        status: pass
    human_judgment: false
  - id: D2
    description: "Resolver builds and returns deep-merged CT_PPr/CT_RPr clones (D-01, STYLE-RESOLVE-01)"
    verification:
      - kind: unit
        ref: "internal/style/resolver_test.go#TestResolveParagraph_Heading2Chain"
        status: pass
      - kind: unit
        ref: "internal/style/resolver_test.go#TestResolveParagraph_DocDefaultsOnly"
        status: pass
    human_judgment: false
  - id: D3
    description: "ResolveRun returns merged rPr (D-04, STYLE-RESOLVE-02)"
    verification:
      - kind: unit
        ref: "internal/style/resolver_test.go#TestResolveRun_Heading2Chain"
        status: pass
    human_judgment: false
  - id: D4
    description: "Cycle detection via visited-set, last-good retained, warning appended (D-05)"
    verification:
      - kind: unit
        ref: "internal/style/resolver_test.go#TestResolveParagraph_CircularBasedOn"
        status: pass
    human_judgment: false
  - id: D5
    description: "Dangling basedOn ref warns and uses docDefaults (D-07)"
    verification:
      - kind: unit
        ref: "internal/style/resolver_test.go#TestResolveParagraph_DanglingBasedOn"
        status: pass
    human_judgment: false
  - id: D6
    description: "latentStyles fallback only when styleId missing (D-12)"
    verification:
      - kind: unit
        ref: "internal/style/resolver_test.go#TestResolveParagraph_LatentStylesFallback"
        status: pass
      - kind: unit
        ref: "internal/style/resolver_test.go#TestResolveParagraph_UnstyledSkipsLatent"
        status: pass
    human_judgment: false
  - id: D7
    description: "Memo caches by styleId; clone independent of cache (D-03, Pitfall 8)"
    verification:
      - kind: unit
        ref: "internal/style/resolver_test.go#TestResolveParagraph_MemoHit"
        status: pass
      - kind: unit
        ref: "internal/style/resolver_test.go#TestResolveParagraph_CloneIndependence"
        status: pass
    human_judgment: false
  - id: D8
    description: "Cache invalidates on styles.xml mutation via Part.IsModified (D-03)"
    verification:
      - kind: unit
        ref: "internal/style/resolver_test.go#TestResolveParagraph_CacheInvalidation"
        status: pass
    human_judgment: false
  - id: D9
    description: "ThemeColor passed through UNCHANGED (phase boundary with 02-02)"
    verification:
      - kind: unit
        ref: "internal/style/resolver_test.go#TestResolveParagraph_ThemeColorPassThrough"
        status: pass
    human_judgment: false
  - id: D10
    description: "NumPr passed through as shallow override (no lvl.PPr merge)"
    verification:
      - kind: unit
        ref: "internal/style/resolver_test.go#TestResolveParagraph_NumPrPassThrough"
        status: pass
    human_judgment: false

duration: 8min
completed: 2026-07-25
status: complete
---

# Phase 02 Style Engine Plan 01: Resolver Core Summary

**Chain-walker resolver with memoization, cycle detection, latentStyles fallback, and clone-before-return — the basedOn inheritance engine for effective paragraph and run properties**

## Performance

- **Duration:** 8 min
- **Started:** 2026-07-25T19:20:00Z
- **Completed:** 2026-07-25T19:28:00Z
- **Tasks:** 2 (Task 1 pre-existing, Tasks 2-3 executed)
- **Files modified:** 4 (2 created, 0 modified)

## Accomplishments

- `internal/style/errors.go` — `ErrStylesParseFailed` sentinel with `%w` wrapping convention
- `internal/style/resolver.go` — `Resolver` struct with lazy parse, memo cache by styleId, cycle detection, clone-before-return, cache invalidation via `Part.IsModified()`, and merge helpers for CT_PPr/CT_RPr composite leaves
- `internal/style/resolver_test.go` — 15 tests covering all behavior cases via synthetic OPC packages (no real .docx fixtures needed)

## Task Commits

Each task was committed atomically:

1. **Task 1: Part.IsModified() accessor** — pre-existing at `internal/opc/package.go:49-52` (already delivered)
2. **Task 2: errors.go + resolver.go** — `0d608dc` (feat)
3. **Task 3: resolver_test.go** — `6307b65` (test)

**Plan metadata:** pending (after SUMMARY.md + STATE.md commit)

## Files Created/Modified

- `internal/style/errors.go` — `ErrStylesParseFailed` sentinel error (NEW)
- `internal/style/resolver.go` — Resolver, NewResolver, ResolveParagraph, ResolveRun, Warnings, merge helpers, clone helpers, cycle detection (NEW)
- `internal/style/resolver_test.go` — 15 synthetic-fixture tests (NEW)

## Decisions Made

- **Cache invalidation via dirty flag (Option A):** Resolver checks `Part.IsModified()` on every Resolve entry; if true, drops both memo maps and re-decodes. O(1) per entry, re-uses existing Phase 1 state. No version counter needed for v1.
- **mergeColor preserves ThemeColor attributes:** Child attributes override parent attributes via deep-merge; ThemeColor is NOT cleared (plan 02-02 concretizes). This differs from the RESEARCH Pattern 3 example which clears ThemeColor on resolution — that belongs in 02-02, not here.
- **Shallow override for NumPr:** NumPr is treated as an opaque pointer field — the resolver does NOT merge numbering level pPr (deferred to plan 02-02). This ensures Phase 1 numbering types remain untouched.
- **Clone via merge-into-empty:** `clonePPr(p)` = `mergePPr(&CT_PPr{}, p)` reuses the same mergePPr implementation, guaranteeing clone and merge share semantics (Pitfall 8 defense).

## Deviations from Plan

### Pre-existing Condition

**Task 1: `Part.IsModified()` accessor already existed at `internal/opc/package.go:49-52`**

- **Note:** The plan specified adding this accessor as a new method, but it was already present in the codebase (likely delivered in Phase 1 or by a prior executor). No code changes were needed.
- **Impact:** None — Task 1 was effectively already complete.

### Auto-fixed Issues

None — plan executed exactly as written for Tasks 2-3.

---

**Total deviations:** None (one pre-existing condition noted)
**Impact on plan:** All verification criteria met. The phase boundary with 02-02 is strictly enforced: no ThemeColor clearing, no numbering.xml parse, no next-chain walk.

## Issues Encountered

None — all tests pass, boundary checks clean.

## Known Stubs

- `TestResolveParagraph_RealFixtures` is `t.Skip`-guarded — real .docx fixtures are user-authored per D-10/D-11. Plan 02-03 will wire the equality harness.

## Next Phase Readiness

- Style resolver core ready for plan 02-02 (theme color concretization + numbering level pPr merge)
- `Resolver` struct holds memo maps and `Warnings()` accumulator — 02-02's theme.go and numbering.go will add their own caches and invalidation
- Phase boundary comment on `ResolveParagraph`/`ResolveRun` explicitly documents ThemeColor and NumPr pass-through

---
*Phase: 02-style-engine*
*Completed: 2026-07-25*
