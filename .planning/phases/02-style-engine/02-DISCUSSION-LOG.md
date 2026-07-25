# Phase 2: Style Engine - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-07-25
**Phase:** 02-style-engine
**Areas discussed:** Resolver output & API surface, Cycle + theme color handling, Clone conflict policy, Validation oracle, Latent styles + numbering depth

---

## Resolver Output & API Surface

### Q1: What should Resolve() return as the effective properties?

| Option | Description | Selected |
|--------|-------------|----------|
| Deep-merged CT_PPr/CT_RPr clone | Freshly-built *wml.CT_PPr and *wml.CT_RPr deep-merged from chain; reuses existing types, no new schema; matches X() escape-hatch | ✓ |
| New EffectiveProps typed struct | Type-safe struct with typed fields; duplicates WML schema in parallel struct | |
| Both — CT clone + typed view | Merged CT_PPr/CT_RPr AND convenience EffectiveProps view over them | |

**User's choice:** Deep-merged CT_PPr/CT_RPr clone
**Notes:** Reuses Phase 1 types, no parallel model to keep in sync.

### Q2: How should callers reach the resolver?

| Option | Description | Selected |
|--------|-------------|----------|
| Internal only — via Document wrapper | style.Resolver internal; callers reach via Document (Paragraph.SetStyle resolves); matches PRD §9 | ✓ |
| Public standalone style.Resolver | Exposed as wordingo.NewResolver or doc.Resolver() | |
| Internal now, public in Phase 4 | Defer commitment, keep options open | |

**User's choice:** Internal only — via Document wrapper
**Notes:** Locked layering; revisit exposure in Phase 4.

### Q3: Resolver caching strategy?

| Option | Description | Selected |
|--------|-------------|----------|
| Memoize by styleId | First call per styleId computes; subsequent calls return cached clone; direct formatting merged per call | ✓ |
| No cache — recompute every call | Always walks chain fresh; no invalidation concerns | |
| Eager precompute at load | Precompute all styles at Document open; O(1) lookups later | |

**User's choice:** Memoize by styleId
**Notes:** Bounded memory; fast repeated SetStyle calls in Phase 4.

### Q4: Direct formatting handling on top of cached chain?

| Option | Description | Selected |
|--------|-------------|----------|
| Resolve(paragraph/run) → clone | Resolver merges direct pPr/rPr on top of cached styleId result; returns fresh clone | ✓ |
| Resolve(styleId) only — caller merges | Returns cached style-chain result only; caller merges direct formatting | |
| Both — ResolveStyle + Resolve | Two methods | |

**User's choice:** Resolve(paragraph/run) → clone
**Notes:** Single entry point; cache stays internal.

---

## Cycle + Theme Color Handling

### Q1: Circular basedOn chain handling?

| Option | Description | Selected |
|--------|-------------|----------|
| Drop cycle + warn | Stop walking at cycle, use last-good, append to doc.Warnings(); matches D-12 | ✓ |
| Error + stop | ErrCircularBasedOn sentinel; hard failure | |
| Break silently | No warning | |

**User's choice:** Drop cycle + warn
**Notes:** Document stays usable; problem visible via Warnings().

### Q2: Theme color resolution timing?

| Option | Description | Selected |
|--------|-------------|----------|
| Resolve to concrete hex at resolve-time | Returns CT_Color with explicit hex; render-ready view | ✓ |
| Preserve theme ref until serialization | Caller or serializer resolves; preserves round-trip info | |
| Both — theme ref + resolved hex | Duality | |

**User's choice:** Resolve to concrete hex at resolve-time
**Notes:** Raw theme1.xml stays untouched; Phase 3 byte-identity applies to raw parts, not resolved clones.

### Q3: Missing theme color / styleId / numbering ref?

| Option | Description | Selected |
|--------|-------------|----------|
| Warn + substitute default | Append to doc.Warnings(); substitute docDefaults/nil color/no numPr | ✓ |
| Error on missing ref | Sentinel errors; caller must handle | |
| Silently drop unresolved | No warning | |

**User's choice:** Warn + substitute default
**Notes:** Resilient by design; problems surface via warnings.

---

## Clone Conflict Policy

### Q1: Conflict policy when target already has style parts?

| Option | Description | Selected |
|--------|-------------|----------|
| Source replaces wholesale; non-empty target errors | ErrCloneTargetNotEmpty; clone is fresh-empty-target op | ✓ |
| Merge by styleId, source wins | Complex merge logic | |
| Skip existing parts | Keep target's | |

**User's choice:** Source replaces wholesale; non-empty target errors
**Notes:** Merge is a separate capability — future phase.

### Q2: How does the cloner treat the 5 source parts?

| Option | Description | Selected |
|--------|-------------|----------|
| Raw bytes pass-through, lazy parse | Via OPC pass-through + relationships + content types; resolver parses lazily | ✓ |
| Parse + validate + re-serialize | Catches source issues early; risks byte-identity | |
| Hybrid — raw for theme/settings, parsed for styles/numbering | Targeted | |

**User's choice:** Raw bytes pass-through, lazy parse
**Notes:** Leverages Phase 1 pass-through; byte-identical source parts; invalid source surfaces at resolve time as warning.

---

## Validation Oracle for Resolver

### Q1: How to prove resolver matches Word's effective property resolution?

| Option | Description | Selected |
|--------|-------------|----------|
| Hand-authored expected-values fixtures | 2-3 style-rich templates + JSON expected CT_PPr/CT_RPr per paragraph; runs in CI | ✓ |
| Open XML SDK as reference oracle (offline) | .NET SDK dumps effective props; commit as expected | |
| Round-trip visual check in Word | Manual Word-open at sign-off only | |

**User's choice:** Hand-authored expected-values fixtures
**Notes:** Self-contained CI; no external tools.

### Q2: Where do validation fixtures live and what shape?

| Option | Description | Selected |
|--------|-------------|----------|
| testdata/word/style-rich/ + expected JSON | Extend Phase 1 fixture pattern | ✓ |
| Separate testdata/style-engine/ corpus | Decoupled from producer-spread corpus | |
| Both — real + hostile fixtures | Real templates + edge cases | |

**User's choice:** testdata/word/style-rich/ + expected JSON
**Notes:** Reuses Phase 1 D-03 pattern.

---

## Latent Styles + Numbering Depth

### Q1: How should the resolver treat latentStyles?

| Option | Description | Selected |
|--------|-------------|----------|
| Fallback for missing styleId only | Consulted when explicit styleId missing from styles.xml; unstyled paragraphs skip latentStyles | ✓ |
| Active in every resolve chain | Aggressive; risk of over-applying | |
| Ignore latentStyles in Phase 2 | Defer to v2 | |

**User's choice:** Fallback for missing styleId only
**Notes:** Matches Word's observed behavior.

### Q2: How deep does numbering resolution go?

| Option | Description | Selected |
|--------|-------------|----------|
| Format + text + start + level pPr | CT_Lvl numFmt, lvlText, start + level pPr merged | ✓ |
| Return numbering references only | Thin resolver; pushes logic to Phase 4 | |
| Full inheritance including style-based numPr | Scope creep beyond STYLE-RESOLVE-03 | |

**User's choice:** Format + text + start + level pPr
**Notes:** Bounded scope; enough for Phase 4 list rendering.

---

## the agent's Discretion

- Internal package layout within `internal/style/`
- Exact fixture filenames and JSON schema for expected-values
- Cache invalidation trigger mechanism
- Hostile fixture directory layout
- Numbering level pPr merge strategy specifics
- Safety-limit numeric thresholds reused from Phase 1's opc layer

## Deferred Ideas

- Merge-by-styleId clone policy — future phase, not STYLE-CLONE-01
- Public `style.Resolver` API exposure — revisit in Phase 4
- Full numbering inheritance including style-based numPr chain — beyond STYLE-RESOLVE-03; v2 candidate
- Style browser / diff-effective-props tooling — not a library concern
- Open XML SDK validator in CI — remains declined for Phase 2
