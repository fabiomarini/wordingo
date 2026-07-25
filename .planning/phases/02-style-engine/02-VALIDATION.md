---
phase: 2
slug: 02-style-engine
status: approved
nyquist_compliant: true
wave_0_complete: true
created: 2026-07-25
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (Go 1.23+ stdlib testing) |
| **Config file** | none — Go convention |
| **Quick run command** | `go test ./internal/style/... -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~4 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/style/... -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 4 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 1 | STYLE-RESOLVE-01 | Part.IsModified accessor exposes dirty flag without mutation | unit | `go test ./internal/opc/... -count=1` | ✅ | ✅ green |
| 02-01-02 | 01 | 1 | STYLE-RESOLVE-01 (D-01/D-04/D-12) | docDefaults → basedOn chain → direct pPr deep-merge, clone-before-return | unit | `go test ./internal/style/... -count=1 -run ResolveParagraph` | ✅ | ✅ green |
| 02-01-03 | 01 | 1 | STYLE-RESOLVE-02 | ResolveRun merges paragraph-chain rPr → rStyle-chain rPr → direct rPr | unit | `go test ./internal/style/... -count=1 -run ResolveRun` | ✅ | ✅ green |
| 02-01-04 | 01 | 1 | STYLE-RESOLVE-03 (D-05) | Cycle detection via visited-set, last-good retained, warning appended | unit | `go test ./internal/style/... -count=1 -run CircularBasedOn` | ✅ | ✅ green |
| 02-01-05 | 01 | 1 | STYLE-RESOLVE-03 (D-07) | Dangling basedOn → warning + docDefaults base | unit | `go test ./internal/style/... -count=1 -run DanglingBasedOn` | ✅ | ✅ green |
| 02-01-06 | 01 | 1 | STYLE-RESOLVE-01 (D-12) | latentStyles fallback only when styleId missing; unstyled skips it | unit | `go test ./internal/style/... -count=1 -run LatentStyles` | ✅ | ✅ green |
| 02-01-07 | 01 | 1 | STYLE-RESOLVE-01 (D-03) | Memo cache by styleId + cache invalidation via Part.IsModified | unit | `go test ./internal/style/... -count=1 -run MemoHit\|CacheInvalidation` | ✅ | ✅ green |
| 02-01-08 | 01 | 1 | STYLE-RESOLVE-01 (Pitfall 8) | Returned clones independent of memo cache | unit | `go test ./internal/style/... -count=1 -run CloneIndependence` | ✅ | ✅ green |
| 02-01-09 | 01 | 1 | STYLE-RESOLVE-03 | ThemeColor + NumPr passed through UNCHANGED (phase boundary) | unit | `go test ./internal/style/... -count=1 -run ThemeColorPassThrough\|NumPrPassThrough` | ✅ | ✅ green |
| 02-02-01 | 02 | 2 | STYLE-RESOLVE-03 (Pitfall 4) | CT_Num.LvlOverride struct extension; backward compatible | unit | `go test ./internal/wml/... -count=1 -run LvlOverride` | ✅ | ✅ green |
| 02-02-02 | 02 | 2 | STYLE-RESOLVE-03 (D-06) | Theme color concretization (themeEnumToElement, sysClr lastClr, tint/shade) | unit | `go test ./internal/style/... -count=1 -run Theme -v` | ✅ | ✅ green |
| 02-02-03 | 02 | 2 | STYLE-RESOLVE-03 (D-13) | Numbering level resolution + lvlOverride (startOverride, full lvl) | unit | `go test ./internal/style/... -count=1 -run 'ResolveLvl' -v` | ✅ | ✅ green |
| 02-02-04 | 02 | 2 | STYLE-RESOLVE-03 | Resolver integration: ResolveRun→theme, ResolveParagraph→numbering | integration | `go test ./internal/style/... -count=1 -run TestResolver_` | ✅ | ✅ green |
| 02-03-01 | 03 | 1 | STYLE-CLONE-01 (D-09) | CloneStyles byte-copies 5 parts via MarkModified + relationships + content types | unit | `go test ./internal/style/... -count=1 -run Clone -v` | ✅ | ✅ green |
| 02-03-02 | 03 | 1 | STYLE-CLONE-01 (D-08) | ErrCloneTargetNotEmpty — atomic precondition scan before copy | unit | `go test ./internal/style/... -count=1 -run ErrCloneTargetNotEmpty` | ✅ | ✅ green |
| 02-03-03 | 03 | 1 | STYLE-CLONE-01 | Absent source part skip silently; save round-trip byte identity | unit | `go test ./internal/style/... -count=1 -run AbsentSourcePartSkip\|SaveRoundTrip` | ✅ | ✅ green |
| 02-04-01 | 04 | 3 | STYLE-CLONE-02 | Heading2 chain resolves effective props matching expected.json (ROADMAP SC#4) | unit | `go test ./internal/style/... -count=1 -run Corpus_Heading2Chain` | ✅ | ✅ green |
| 02-04-02 | 04 | 3 | STYLE-CLONE-02 | Theme refs concretize over cloned packages (accent1, dark1, hyperlink) | unit | `go test ./internal/style/... -count=1 -run Corpus_ThemeRefs` | ✅ | ✅ green |
| 02-04-03 | 04 | 3 | STYLE-CLONE-02 | Multi-level numbering merges level pPr ind over cloned packages | unit | `go test ./internal/style/... -count=1 -run Corpus_MultiLevelNumbering` | ✅ | ✅ green |
| 02-04-04 | 04 | 3 | STYLE-CLONE-02 | Hostile paths (circular, dangling, missing numId) surface warnings | unit | `go test ./internal/style/... -count=1 -run Corpus_CircularBasedOn\|Corpus_DanglingBasedOn\|Corpus_MissingNumId` | ✅ | ✅ green |
| 02-04-05 | 04 | 3 | STYLE-CLONE-02 | Named styles from cloned template applicable by name | unit | `go test ./internal/style/... -count=1 -run Corpus_NamedStylesApplicableByName` | ✅ | ✅ green |

*Status: ✅ green · ❌ red · ⚠️ flaky*

---

## Per-Requirement Coverage

| Requirement | Plans | Test File | Tests | Status |
|-------------|-------|-----------|-------|--------|
| STYLE-RESOLVE-01 | 02-01 | `resolver_test.go` | TestResolveParagraph_DocDefaultsOnly, TestResolveParagraph_Heading2Chain, TestResolveParagraph_ThemeColorPassThrough, TestResolveParagraph_NumPrPassThrough, TestResolveParagraph_NoStylesPart | ✅ COVERED |
| STYLE-RESOLVE-02 | 02-01 | `resolver_test.go` | TestResolveRun_Heading2Chain, TestResolver_Warnings | ✅ COVERED |
| STYLE-RESOLVE-03 | 02-01, 02-02 | `resolver_test.go`, `theme_test.go`, `numbering_test.go` | TestResolveParagraph_CircularBasedOn, TestResolveParagraph_DanglingBasedOn, TestResolveParagraph_LatentStylesFallback, TestResolveColor_*, TestResolveLvl_*, TestResolver_* | ✅ COVERED |
| STYLE-CLONE-01 | 02-03 | `cloner_test.go` | TestCloneStyles_FreshEmptyTarget, TestCloneStyles_ErrCloneTargetNotEmpty, TestCloneStyles_AbsentSourcePartSkip, TestCloneStyles_RelationshipAllocation, TestCloneStyles_ContentTypeOverrides, TestCloneStyles_SaveRoundTrip, TestCloneStyles_NoReParse, TestCloneStyles_IdempotentFails, TestCloneStyles_NilPackageDefensive | ✅ COVERED |
| STYLE-CLONE-02 | 02-04 | `corpus_test.go` | TestCorpus_Heading2Chain, TestCorpus_ThemeRefs, TestCorpus_MultiLevelNumbering, TestCorpus_CircularBasedOn, TestCorpus_DanglingBasedOn, TestCorpus_MissingNumId, TestCorpus_CloneTargetNotEmpty, TestCorpus_NamedStylesApplicableByName | ✅ COVERED |

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 4s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-07-25
