# Phase 2: Style Engine - Context

**Gathered:** 2026-07-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Given any template .docx, the library (a) clones its complete style dependency graph (styles.xml + numbering.xml + fontTable.xml + theme.xml + settings.xml, with relationships and content types updated) into a target document, and (b) resolves effective paragraph/run properties through the full inheritance chain — docDefaults → latentStyles → basedOn chain → direct formatting — with theme colors resolved to concrete hex, numbering definitions resolved to format/text/start + level pPr, and circular basedOn detected without infinite recursion. Requirements: STYLE-CLONE-01..02, STYLE-RESOLVE-01..03.

Round-trip safety (STYLE-ROUNDTRIP-01..02) and FromTemplate (CREATE-03, CREATE-04) belong to Phase 3 — this phase delivers clone + resolve primitives only.

</domain>

<decisions>
## Implementation Decisions

### Resolver Output & API Surface
- **D-01:** `Resolve()` returns freshly-built, deep-merged `*wml.CT_PPr` and `*wml.CT_RPr` clones (no new parallel `EffectiveProps` struct). Reuses existing WML types from Phase 1; matches the `X()` escape-hatch pattern. Callers receive clones — mutating them does not touch the source chain.
- **D-02:** `style.Resolver` is internal only (per PRD §9 architecture — `internal/style` not imported by users). Callers reach resolution via the `Document` wrapper in Phase 4 (e.g., `Paragraph.SetStyle("Heading1")` resolves under the hood). No public `style.Resolver` API in this phase; revisit exposure in Phase 4 once the content API shape is clear.
- **D-03:** Memoize by `styleId` — first `Resolve()` call per styleId walks the chain and caches the merged clone; subsequent calls return a fresh clone of the cache. Direct formatting (paragraph/run pPr/rPr) is merged on top per call. Cache invalidates when `styles.xml` is mutated.
- **D-04:** `Resolve(paragraph|run)` returns the full effective props (styleId chain + direct formatting merged into a fresh clone). Single entry point — caller never touches the cache directly. No separate `ResolveStyle(styleId)`-only method.

### Cycle + Theme Color Handling
- **D-05:** Circular basedOn chain (A → B → A): drop the cycle, use last-good effective props, append a warning to `doc.Warnings()`. Matches Phase 1 D-12 pattern. No hard error — document stays usable, problem stays visible.
- **D-06:** Theme colors resolve to concrete hex at resolve-time. The returned `CT_Color` carries explicit hex (e.g., `#1F4E79`), not the theme placeholder. The raw `theme1.xml` part is not mutated — resolution is a read-only view over the cached style chain. (Round-trip byte-identity in Phase 3 applies to raw parts, not resolved clones, so this is safe.)
- **D-07:** Missing theme color ref, missing styleId in basedOn chain, or unresolvable numbering numId/ilvl: append to `doc.Warnings()`, substitute a sensible default (`docDefaults`, absent `CT_Color`, no `numPr`). No sentinel errors for resolver misses — resilient by design, problems surface via warnings.

### Clone Conflict Policy
- **D-08:** Target conflict policy: source replaces wholesale. If the target document already has any of styles/numbering/theme/fontTable/settings, clone returns `ErrCloneTargetNotEmpty`. Clone is a fresh-empty-target operation — merging by styleId is a separate capability and belongs in a future phase (out of scope).
- **D-09:** Cloner treats the 5 source parts as raw bytes via OPC pass-through (Phase 1's `RawXML`/byte-copy path) + adds relationships + content types. No re-parse on clone. Resolver lazily parses `styles.xml`/`numbering.xml`/`theme1.xml` on first query via the existing `xmlutil` + `wml` unmarshal. Invalid source surfaces at resolve time as a warning (D-07), not at clone time.

### Validation Oracle
- **D-10:** Resolver validation uses hand-authored expected-values fixtures: 2–3 "style-rich" templates under `testdata/word/style-rich/` plus a JSON/markdown file per template listing expected effective `CT_PPr`/`CT_RPr` for specific paragraphs/runs. Expected values captured by hand from reading the template's `styles.xml`/`theme1.xml`. No Open XML SDK container in CI; no round-trip visual check as primary assertion.
- **D-11:** Fixture shape: (a) deep basedOn chain (Heading4 → Heading3 → Heading2 → Heading1 → Normal), (b) themeColor refs in styles + docDefaults, (c) multi-level numbering (numId + ilvl across 3 levels). Hostile fixtures (circular basedOn, dangling basedOn ref, missing numId) covered by the warning-path tests from D-05/D-07 — can live alongside or in a sibling dir; researcher to decide layout.

### Latent Styles + Numbering Depth
- **D-12:** latentStyles consulted only as fallback when an explicit styleId reference is missing from `styles.xml`. Unstyled paragraphs (no `pStyle`) skip latentStyles — they use `docDefaults` only. Matches Word's observed behavior where latentStyles defines default formatting for styles not yet instantiated.
- **D-13:** Numbering resolution depth: resolve numId + ilvl to the abstract numbering level's `numFmt`, `lvlText`, `start`, and the level's `pPr` (indentation merged into the paragraph's effective pPr). Does NOT include style-based numPr inheritance — that's beyond STYLE-RESOLVE-03 ("resolve numbering definitions"). Bounded scope.

### the agent's Discretion
- Internal package layout within `internal/style/` (resolver.go, cloner.go, theme.go, numbering.go — split or merged)
- Exact fixture filenames and JSON schema for expected-values
- Cache invalidation trigger mechanism (dirty flag on styles.xml write vs version counter)
- Whether hostile fixtures live under `testdata/word/style-rich/` or a sibling `testdata/style-engine/` dir
- Numbering level pPr merge strategy specifics (which fields override paragraph direct pPr vs complement)
- Safety-limit numeric thresholds reused from Phase 1's opc layer (path traversal on clone target, max parts)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Product & Requirements
- `.planning/PRD.md` — §6 FR-3 (Style Engine: FR-3.1 STYLE-CLONE, FR-3.2 STYLE-RESOLVE, FR-3.3 STYLE-ROUNDTRIP boundary with Phase 3), §9 architecture (`internal/style` layer, strict layering)
- `.planning/REQUIREMENTS.md` — STYLE-CLONE-01..02, STYLE-RESOLVE-01..03 normative text
- `.planning/PROJECT.md` — Constraints (stdlib only, MIT, Go 1.23+, <5MB), Key Decisions table (three separate style operations; style thesis before rich content)
- `.planning/ROADMAP.md` — Phase 2 goal, success criteria, plan breakdown (02-01 resolver, 02-02 theme+numbering, 02-03 cloner + corpus)

### Phase 1 Context (carried forward — must respect)
- `.planning/phases/01-foundation/01-CONTEXT.md` — D-01..D-12: fixture corpus + golden files, per-part byte diff, xmlutil token-stream wrapper, canonical OOXML prefixes on write, pointer-based optional fields, RawXML hoarding, sentinel errors + `%w` wrapping, `Warnings() []string`

### External Specifications (no local copies — cite in RESEARCH.md as needed)
- ISO/IEC 29500 Part 1 (WordprocessingML) — §17.7 (Style Definitions), §17.3 (Numbering), §17.9 (Theme)
- ECMA-376 Part 1 — style inheritance model (docDefaults → latentStyles → basedOn chain → direct formatting)
- Microsoft `DocumentFormat.OpenXml` (.NET SDK) — conceptual reference for style inheritance behavior; no code ported

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/wml/styles.go` — `CT_Styles`, `CT_Style`, `CT_DocDefaults`, `CT_RPrDefault`, `CT_PPrDefault`, `CT_LatentStyles`, `CT_LsdException`, `CT_BasedOn`, `CT_Next`, `CT_Link` (Phase 1 delivered) — the resolver reads these directly
- `internal/wml/numbering.go` — `CT_Numbering`, `CT_AbstractNum`, `CT_Num`, `CT_AbstractNumID`, `CT_Lvl` — numbering resolution types ready
- `internal/wml/properties.go` — `CT_PStyle`, `CT_RStyle`, `CT_NumPr`, `CT_NumId`, `CT_ILvl`, `CT_RFonts`, `CT_Color`, `CT_Sz`, `CT_OnOff`, `CT_Shd`, `CT_U`, `CT_Highlight`, `CT_VertAlign`, `CT_Jc`, `CT_Spacing`, `CT_Ind` — the fields the resolver merges
- `internal/xmlutil/` — namespace registry, `RawXML` token-blob hoarding, canonical-prefix encoder, safe decoder — resolver parses via this layer
- `internal/opc/` — `Open`/`Save`, pass-through preservation for unmodeled parts, high-water rId allocator — cloner uses pass-through path for the 5 source parts
- `testdata/` — Word/LibreOffice/Google Docs/hostile fixtures from Phase 1; `testdata/word/style-rich/` is the extension point for resolver fixtures (D-10)
- `defaults/` — blank-document defaults (`styles.xml`, `theme1.xml`, `fontTable.xml`, `settings.xml`, `webSettings.xml`) — reference for what a "complete style dependency graph" looks like

### Established Patterns
- Wrapper-over-schema with `X()` escape hatch — resolver returns raw WML types, not parallel structs
- Lazy part loading — resolver parses `styles.xml`/`numbering.xml`/`theme1.xml` on first query, not on clone
- Unknown-element hoarding via `RawXML` — parts not yet modeled still survive clone + round-trip
- URI-based namespace registry — resolver uses the same `xmlutil` decoder, immune to prefix variation
- Pointer-based optional fields (nil = absent) — merge logic uses nil checks to decide chain overrides
- Sentinel errors + `Warnings() []string` — cycle/missing-ref surface via warnings, not new error taxonomy

### Integration Points
- `internal/style/` package is new — depends on `internal/opc`, `internal/wml`, `internal/xmlutil` (per PRD §9 strict layering)
- Phase 3 `Document` open/read/save will use the cloner for `FromTemplate()` (CREATE-03) and the resolver for STYLE-ROUNDTRIP checks
- Phase 4 content API (`Paragraph.SetStyle`, `Run.SetBold`, ...) will call `Resolver.Resolve()` under the hood

</code_context>

<specifics>
## Specific Ideas

- Heading2 chain (Heading2 → Heading1 → Normal) is the canonical test case from ROADMAP success criteria #4 — must be in the style-rich fixture corpus
- Theme colors resolve to concrete hex (not placeholders) because resolved props are a render-ready view; raw parts stay untouched for Phase 3 byte-identity
- Clone is fresh-empty-target only — any existing style part in target = error. No merge story in v1; merge by styleId is a future-phase capability
- Resolver returns clones, never the cached merged result, so callers can safely mutate without touching shared state

</specifics>

<deferred>
## Deferred Ideas

- Merge-by-styleId clone policy (when target already has styles, merge instead of error) — future phase, not STYLE-CLONE-01
- Public `style.Resolver` API exposure — revisit in Phase 4 once content API shape is clear
- Full numbering inheritance including style-based numPr chain — beyond STYLE-RESOLVE-03; v2 candidate
- Style browser / diff-effective-props tooling — not a library concern
- Open XML SDK validator in CI — declined in Phase 1; remains declined for Phase 2 (hand-authored expected-values covers resolver correctness)

</deferred>

---

*Phase: 02-style-engine*
*Context gathered: 2026-07-25*
