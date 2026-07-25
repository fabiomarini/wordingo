---
phase: 02-style-engine
plan: 04
subsystem: testing
tags: [corpus-validation, expected-json, synthetic-fixtures, heading2-chain, theme-refs, numbering, hostile-paths]

requires:
  - phase: 02-01
    provides: Resolver (NewResolver, ResolveParagraph, ResolveRun, Warnings, cycle detection)
  - phase: 02-02
    provides: themeCache.ResolveColor, numberingCache.ResolveLvl, resolver post-processing wiring
  - phase: 02-03
    provides: CloneStyles, ErrCloneTargetNotEmpty, buildFreshTargetZip, buildTargetWithStylesZip helpers

provides:
  - Synthetic fixture builders (heading-chain, theme-refs, multi-level-numbering, circular-basedon, dangling-basedon, missing-numid) for CI
  - expected.json schema files (6) defining expected effective CT_PPr/CT_RPr across style-rich + hostile paths
  - Field-by-field nil-check equality assertion helpers (assertEffectivePPr, assertEffectiveRPr, assertWarnings)
  - ROADMAP success criterion #4 coverage (Heading2 chain canonical case)
  - STYLE-CLONE-02 acceptance test (named styles from cloned template applicable by name)

affects: Phase 3 (STYLE-ROUNDTRIP validation may extend the corpus)

tech-stack:
  added: []
  patterns:
    - Synthetic fixture builder via zip.Writer + opc.Open (same pattern as cloner_test.go + opc_test.go)
    - expected.json schema with `null`=`nil` vs `"value"`=`present` semantics
    - Hostile warnings assertion via strings.Contains on Warnings() output

key-files:
  created:
    - internal/style/corpus_test.go
    - testdata/word/style-rich/heading-chain.expected.json
    - testdata/word/style-rich/theme-refs.expected.json
    - testdata/word/style-rich/multi-level-numbering.expected.json
    - testdata/style-engine/hostile/circular-basedon.expected.json
    - testdata/style-engine/hostile/dangling-basedon.expected.json
    - testdata/style-engine/hostile/missing-numid.expected.json
  modified: []

key-decisions:
  - "corpus test reads document.xml from source package (CloneStyles does NOT copy document.xml)"
  - "Single corpus_test.go file with shared assertion helpers and fixture builders (splitting would duplicate helpers across files)"
  - "assertion helper uses map[string]any with explicit field-by-field switch (not reflect.DeepEqual) for precise nil-vs-present error reporting"
  - "Hostile fixtures under testdata/style-engine/hostile/ (separate from style-rich/ for clean equality assertions)"

patterns-established:
  - "corpusResolveParagraph helper: CloneStyles → NewResolver → ResolveParagraph/ResolveRun on source document paragraphs"
  - "assertCloneCorrectness: after CloneStyles + Save + re-open, verify 5 style parts byte-identical"
  - "subtest-per-case for multi-case fixtures (theme-refs, multi-level-numbering)"

requirements-completed: [STYLE-CLONE-02]

coverage:
  - id: D01
    description: Heading2 chain (Heading2→Heading1→Normal) resolves effective CT_PPr/CT_RPr matching heading-chain.expected.json
    verification:
      - kind: unit
        ref: corpus_test.go#TestCorpus_Heading2Chain
        status: pass
    human_judgment: false
  - id: D02
    description: Theme color refs (accent1, dark1 sysClr, hyperlink) concretize to concrete hex over cloned packages
    verification:
      - kind: unit
        ref: corpus_test.go#TestCorpus_ThemeRefs
        status: pass
    human_judgment: false
  - id: D03
    description: Multi-level numbering merges level pPr ind into effective pPr over cloned packages
    verification:
      - kind: unit
        ref: corpus_test.go#TestCorpus_MultiLevelNumbering
        status: pass
    human_judgment: false
  - id: D04
    description: Circular basedOn (D-05) surfaces warning + no panic + last-good props
    verification:
      - kind: unit
        ref: corpus_test.go#TestCorpus_CircularBasedOn
        status: pass
    human_judgment: false
  - id: D05
    description: Dangling basedOn (D-07) surfaces warning + docDefaults merged
    verification:
      - kind: unit
        ref: corpus_test.go#TestCorpus_DanglingBasedOn
        status: pass
    human_judgment: false
  - id: D06
    description: Missing numId (D-07) surfaces warning + numPr preserved
    verification:
      - kind: unit
        ref: corpus_test.go#TestCorpus_MissingNumId
        status: pass
    human_judgment: false
  - id: D07
    description: Clone target not empty returns ErrCloneTargetNotEmpty (D-08)
    verification:
      - kind: unit
        ref: corpus_test.go#TestCorpus_CloneTargetNotEmpty
        status: pass
    human_judgment: false
  - id: D08
    description: Named styles from cloned template applicable by name (STYLE-CLONE-02)
    verification:
      - kind: unit
        ref: corpus_test.go#TestCorpus_NamedStylesApplicableByName
        status: pass
    human_judgment: false

duration: 18 min
completed: 2026-07-25
status: complete
---

# Phase 2 Plan 4: Corpus Validation Summary

**6 expected.json schema files + synthetic in-memory fixture builders asserting effective resolved props over cloned packages — Heading2 chain canonical case (ROADMAP #4), theme color concretization, multi-level numbering, hostile paths (circular/dangling/missing), and STYLE-CLONE-02 acceptance**

## Performance

- **Duration:** 18 min
- **Started:** 2026-07-25T21:07:30Z
- **Completed:** 2026-07-25T21:25:37Z
- **Tasks:** 1 (tdd - corpus harness + 6 expected.json schema files)
- **Files modified:** 7

## Accomplishments
- 9 test functions covering all corpus validation paths (Heading2 chain, theme refs, multi-level numbering, circular basedOn, dangling basedOn, missing numId, clone target not empty, named styles applicable by name, real fixtures skip)
- 6 synthetic fixture builders with zip.Writer producing hermetic OPC packages
- 6 expected.json schema files under testdata/word/style-rich/ (3) and testdata/style-engine/hostile/ (3)
- Field-by-field nil-check assertion helpers: assertEffectivePPr, assertEffectiveRPr, assertWarnings
- clone correctness assertion (byte-identity of 5 style parts after Save + re-open)
- All upstream plans (02-01 resolver, 02-02 theme/numbering, 02-03 cloner) exercised end-to-end over cloned packages

## Task Commits

Each task was committed atomically:

1. **Task 1: Build corpus_test.go + 6 expected.json schema files** - `8bad4dc` (test)

## Files Created/Modified
- `internal/style/corpus_test.go` - Corpus validation harness with 9 test functions, 6 synthetic fixture builders, expectedFixture schema types, assertion helpers
- `testdata/word/style-rich/heading-chain.expected.json` - Canonical Heading2 chain expected values (ROADMAP success criterion #4)
- `testdata/word/style-rich/theme-refs.expected.json` - Theme color refs (accent1, dark1 sysClr, hyperlink) expected values
- `testdata/word/style-rich/multi-level-numbering.expected.json` - 3-level numbering expected values (ilvl 0/1/2)
- `testdata/style-engine/hostile/circular-basedon.expected.json` - D-05 cycle expected values + warnings
- `testdata/style-engine/hostile/dangling-basedon.expected.json` - D-07 dangling ref expected values + warnings
- `testdata/style-engine/hostile/missing-numid.expected.json` - D-07 missing numId expected values + warnings

## Decisions Made
- `corpusResolveParagraph` helper reads paragraph XML from source package (CloneStyles only copies style parts, not document.xml)
- Single corpus_test.go file preserves shared assertion helpers for all test functions (no duplication across fixture-family test files)
- `map[string]any` assertion approach provides flexible field-by-field comparison with precise nil-vs-present error reporting per field
- Hostile fixtures in separate `testdata/style-engine/hostile/` directory keeps "style-rich" equality assertions clean from warning-string noise
- Synthetic fixture builders embed XML string literals for hermetic CI tests (no external file dependencies)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- Initial test failures due to parsing document.xml from cloned target (which has a generic empty paragraph from buildFreshTargetZip) instead of source — fixed by reading document.xml from source package
- Hostile fixture paths used incorrect `../` prefix in expectedFixturePath calls — fixed by using correct relative path within testdata/

## Next Phase Readiness
- Phase 2 style engine complete (all 4 plans executed: 02-01 resolver, 02-02 theme/numbering, 02-03 cloner, 02-04 corpus validation)
- Phase 3 (Document Model) can now consume the complete style stack for open/read/save and FromTemplate
- STYLE-CLONE-02 requirement completed (named styles from cloned template applicable by name)
- ROADMAP success criterion #4 satisfied (Heading2 chain canonical case tested)
- User-authored real .docx fixtures still pending (listed in user_setup artifacts)

---

*Phase: 02-style-engine*
*Completed: 2026-07-25*
