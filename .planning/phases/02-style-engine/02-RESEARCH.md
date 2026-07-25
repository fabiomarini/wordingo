# Phase 2: Style Engine - Research

**Researched:** 2026-07-25
**Domain:** OOXML style inheritance model (ISO/IEC 29500-1 §17.7, §17.3, §17.9), theme color resolution, numbering resolution, OPC pass-through cloning
**Confidence:** HIGH (mechanics anchored to spec sections + existing Phase 1 code; spec section text fetched from training knowledge, not re-pulled this session — tagged CITED/ASSUMED where appropriate)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Resolver Output & API Surface**
- **D-01:** `Resolve()` returns freshly-built, deep-merged `*wml.CT_PPr` and `*wml.CT_RPr` clones (no new parallel `EffectiveProps` struct). Reuses existing WML types from Phase 1; matches the `X()` escape-hatch pattern. Callers receive clones — mutating them does not touch the source chain.
- **D-02:** `style.Resolver` is internal only (per PRD §9 architecture — `internal/style` not imported by users). Callers reach resolution via the `Document` wrapper in Phase 4 (e.g., `Paragraph.SetStyle("Heading1")` resolves under the hood). No public `style.Resolver` API in this phase; revisit exposure in Phase 4 once the content API shape is clear.
- **D-03:** Memoize by `styleId` — first `Resolve()` call per styleId walks the chain and caches the merged clone; subsequent calls return a fresh clone of the cache. Direct formatting (paragraph/run pPr/rPr) is merged on top per call. Cache invalidates when `styles.xml` is mutated.
- **D-04:** `Resolve(paragraph|run)` returns the full effective props (styleId chain + direct formatting merged into a fresh clone). Single entry point — caller never touches the cache directly. No separate `ResolveStyle(styleId)`-only method.

**Cycle + Theme Color Handling**
- **D-05:** Circular basedOn chain (A → B → A): drop the cycle, use last-good effective props, append a warning to `doc.Warnings()`. Matches Phase 1 D-12 pattern. No hard error — document stays usable, problem stays visible.
- **D-06:** Theme colors resolve to concrete hex at resolve-time. The returned `CT_Color` carries explicit hex (e.g., `#1F4E79`), not the theme placeholder. The raw `theme1.xml` part is not mutated — resolution is a read-only view over the cached style chain. (Round-trip byte-identity in Phase 3 applies to raw parts, not resolved clones, so this is safe.)
- **D-07:** Missing theme color ref, missing styleId in basedOn chain, or unresolvable numbering numId/ilvl: append to `doc.Warnings()`, substitute a sensible default (`docDefaults`, absent `CT_Color`, no `numPr`). No sentinel errors for resolver misses — resilient by design, problems surface via warnings.

**Clone Conflict Policy**
- **D-08:** Target conflict policy: source replaces wholesale. If the target document already has any of styles/numbering/theme/fontTable/settings, clone returns `ErrCloneTargetNotEmpty`. Clone is a fresh-empty-target operation — merging by styleId is a separate capability and belongs in a future phase (out of scope).
- **D-09:** Cloner treats the 5 source parts as raw bytes via OPC pass-through (Phase 1's `RawXML`/byte-copy path) + adds relationships + content types. No re-parse on clone. Resolver lazily parses `styles.xml`/`numbering.xml`/`theme1.xml` on first query via the existing `xmlutil` + `wml` unmarshal. Invalid source surfaces at resolve time as a warning (D-07), not at clone time.

**Validation Oracle**
- **D-10:** Resolver validation uses hand-authored expected-values fixtures: 2–3 "style-rich" templates under `testdata/word/style-rich/` plus a JSON/markdown file per template listing expected effective `CT_PPr`/`CT_RPr` for specific paragraphs/runs. Expected values captured by hand from reading the template's `styles.xml`/`theme1.xml`. No Open XML SDK container in CI; no round-trip visual check as primary assertion.
- **D-11:** Fixture shape: (a) deep basedOn chain (Heading4 → Heading3 → Heading2 → Heading1 → Normal), (b) themeColor refs in styles + docDefaults, (c) multi-level numbering (numId + ilvl across 3 levels). Hostile fixtures (circular basedOn, dangling basedOn ref, missing numId) covered by the warning-path tests from D-05/D-07 — can live alongside or in a sibling dir; researcher to decide layout.

**Latent Styles + Numbering Depth**
- **D-12:** latentStyles consulted only as fallback when an explicit styleId reference is missing from `styles.xml`. Unstyled paragraphs (no pStyle) skip latentStyles — they use `docDefaults` only. Matches Word's observed behavior where latentStyles defines default formatting for styles not yet instantiated.
- **D-13:** Numbering resolution depth: resolve numId + ilvl to the abstract numbering level's `numFmt`, `lvlText`, `start`, and the level's `pPr` (indentation merged into the paragraph's effective pPr). Does NOT include style-based numPr inheritance — that's beyond STYLE-RESOLVE-03 ("resolve numbering definitions"). Bounded scope.

### the agent's Discretion
- Internal package layout within `internal/style/` (resolver.go, cloner.go, theme.go, numbering.go — split or merged)
- Exact fixture filenames and JSON schema for expected-values
- Cache invalidation trigger mechanism (dirty flag on styles.xml write vs version counter)
- Whether hostile fixtures live under `testdata/word/style-rich/` or a sibling `testdata/style-engine/` dir
- Numbering level pPr merge strategy specifics (which fields override paragraph direct pPr vs complement)
- Safety-limit numeric thresholds reused from Phase 1's opc layer (path traversal on clone target, max parts)

### Deferred Ideas (OUT OF SCOPE)
- Merge-by-styleId clone policy (when target already has styles, merge instead of error) — future phase, not STYLE-CLONE-01
- Public `style.Resolver` API exposure — revisit in Phase 4 once content API shape is clear
- Full numbering inheritance including style-based numPr chain — beyond STYLE-RESOLVE-03; v2 candidate
- Style browser / diff-effective-props tooling — not a library concern
- Open XML SDK validator in CI — declined in Phase 1; remains declined for Phase 2 (hand-authored expected-values covers resolver correctness)
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| STYLE-CLONE-01 | Copy template style dependency graph into target document — styles.xml + numbering.xml + fontTable.xml + theme.xml + settings.xml, with relationships and content types updated | §Cloner Strategy, §OPC Pass-Through Path, §Relationships & Content Types for the 5 parts |
| STYLE-CLONE-02 | Named styles from template applicable to new content by name | Implicit — cloner preserves styles.xml verbatim; resolver exposes by styleId. No separate research; covered by STYLE-RESOLVE + clone correctness |
| STYLE-RESOLVE-01 | Resolve effective paragraph properties through the full chain: docDefaults → latentStyles → basedOn chain → direct formatting | §OOXML Style Inheritance Model, §basedOn Chain Resolution Algorithm, §Cache Invalidation |
| STYLE-RESOLVE-02 | Resolve effective run properties through the full chain, including run styles | §basedOn Chain (rPr merge), §Pattern: rStyle resolution (parallel pStyle path) |
| STYLE-RESOLVE-03 | Resolve theme colors and numbering definitions; detect circular basedOn references | §Theme Color Resolution, §Numbering Resolution, §Cycle Detection |
</phase_requirements>

## Summary

Phase 2 builds the `internal/style` package: a cloner that byte-copies five source parts (styles.xml, numbering.xml, fontTable.xml, theme1.xml, settings.xml) plus their relationships and content-type overrides into a fresh-empty target, and a resolver that walks the full OOXML inheritance chain to return deep-merged `*wml.CT_PPr` / `*wml.CT_RPr` clones. The cloner reuses Phase 1's OPC pass-through path verbatim — `opc.Package.MarkModified` / `Relationships.NextRID` / `ContentTypes` — so clone is largely a wiring exercise with one new error sentinel (`ErrCloneTargetNotEmpty`). The resolver is where the real work lives: a per-styleId memoized chain walker that merges docDefaults → (latentStyles fallback per D-12) → basedOn chain → direct formatting, with cycle detection (visited-set + last-good fallback per D-05), theme color lookup (theme1.xml read-only, resolved to concrete hex per D-06), and numbering level resolution (numId+ilvl → abstract lvl's numFmt/lvlText/start + level pPr merged into effective pPr per D-13).

Three findings materially shape the plan. **First**, the existing `internal/wml/numbering.go` `CT_Num` struct is **missing the `lvlOverride` child element** — ISO/IEC 29500-1 §17.9.17 defines `w:num/w:lvlOverride` (with `startOverride` and `lvl` sub-elements) which overrides abstract numbering-level values for a specific numbering instance. Per D-09 the cloner never re-parses numbering.xml (so byte-identity is unaffected), but the resolver *does* parse it lazily and a Word-authored numbering.xml carrying `lvlOverride` will silently lose the override on resolve. The struct must be extended before plan 02-02. **Second**, the OOXML `themeColor` enum string space (`dark1`/`light1`/`dark2`/`light2`/`accent1`..`accent6`/`hyperlink`/`followedHyperlink`) does **not** match the element names inside `theme1.xml`'s `a:clrScheme` (`dk1`/`lt1`/`dk2`/`lt2`/`accent1`..`accent6`/`hlink`/`folHlink`) — a name-mapping table is required, not a string-equal lookup. **Third**, basedOn/link/next are three independent style relationships with different semantics; only `basedOn` contributes properties to the effective chain (link pulls run-props from a paired character style and is consulted when resolving a linked paragraph style's rPr; `next` is purely an editor-cursor hint and is **never** walked by the resolver — a common implementer bug).

**Primary recommendation:** Lay out `internal/style/` as four files: `cloner.go` (byte copy + rels/content-types wiring + `ErrCloneTargetNotEmpty`), `theme.go` (theme1.xml parse-on-demand + color name map + tint/shade math), `numbering.go` (numbering.xml parse-on-demand + numId+ilvl → CT_Lvl + lvlOverride handling + level pPr merge), `resolver.go` (chain walker + cycle detection + memo cache + direct-formatting overlay). Use a **dirty flag** on `opc.Package` mutations of `word/styles.xml` for cache invalidation (simpler than a version counter; see §Cache Invalidation). Hostile fixtures live in a sibling `testdata/style-engine/hostile/` dir to keep the golden "style-rich" corpus clean for equality assertions.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Byte-copy 5 source parts (styles/numbering/fontTable/theme/settings) | internal/style (cloner) | internal/opc (MarkModified + byte pass-through) | Cloner orchestrates; opc owns the container mechanics |
| rId allocation for cloned parts in target | internal/opc (`Relationships.NextRID`) | internal/style (calls it) | Phase 1 already provides high-water-mark allocator (OPC-06) |
| Content-type overrides for cloned parts | internal/opc (`ContentTypes`) | internal/style | Phase 1 owns `[Content_Types].xml` model |
| styles.xml parse → CT_Styles tree | internal/style (resolver, lazy) | internal/wml (types) + internal/xmlutil (decoder) | Lazy on first query per D-09 |
| numbering.xml parse → CT_Numbering tree | internal/style (numbering.go, lazy) | internal/wml + internal/xmlutil | Same lazy pattern |
| theme1.xml parse → color lookup table | internal/style (theme.go, lazy) | internal/xmlutil (raw token scan — DrawingML types are out of scope) | Only `a:clrScheme` children needed; full CT_Theme not modeled (Phase 1 decision: theme kept raw) |
| docDefaults → basedOn chain walk + merge | internal/style (resolver) | — | Pure resolver logic over CT_Styles tree |
| Cycle detection + last-good fallback | internal/style (resolver) | — | D-05; visited-set per chain walk |
| Theme color → concrete hex | internal/style (theme.go) | — | D-06; read-only over cached theme1.xml bytes |
| Numbering level resolution | internal/style (numbering.go) | internal/style (resolver) merges lvl.pPr into effective pPr | D-13; numId+ilvl lookup, lvlOverride handling |
| Cache (memo by styleId) | internal/style (resolver) | — | D-03; invalidated on styles.xml mutation |
| Warnings surfacing (cycle, missing ref) | internal/style → eventually `Document.Warnings()` | — | Phase 1 D-12 pattern; Phase 2 resolver appends, Phase 4 Document surfaces |
| Expected-values fixtures | testdata/word/style-rich/ + testdata/style-engine/hostile/ | — | D-10/D-11; hand-authored JSON |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.23+ (floor) | Language/runtime | Locked Phase 1 D-10 |
| `encoding/xml` | stdlib | Parse styles.xml / numbering.xml / theme1.xml lazily | Phase 1 verified: URI-form tags resolve any prefix (OPC-03) [VERIFIED: pkg.go.dev/encoding/xml — Phase 1 research] |
| `internal/opc` | Phase 1 | Cloner reuses `Package.MarkModified`, `Relationships.NextRID`, `ContentTypes` | Phase 1 delivered; cloner is pure wiring over existing API [VERIFIED: codebase — internal/opc/package.go, relationships.go] |
| `internal/wml` | Phase 1 | CT_Styles/CT_Style/CT_DocDefaults/CT_LatentStyles/CT_Numbering/CT_AbstractNum/CT_Num/CT_Lvl/CT_PPr/CT_RPr/CT_Color/CT_NumPr — the types the resolver merges | [VERIFIED: codebase — internal/wml/styles.go, numbering.go, properties.go, document.go] |
| `internal/xmlutil` | Phase 1 | URI-namespace decoder + canonical encoder | Resolver parses via `xmlutil` to remain producer-agnostic [VERIFIED: codebase] |
| `errors`, `fmt` | stdlib | Sentinel `ErrCloneTargetNotEmpty` + `%w` wrapping | Phase 1 D-11 [VERIFIED] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `strconv` | stdlib | int64 ↔ string for numId/ilvl/start parsing | numbering.go |
| `strings` | stdlib | hex color normalization, themeColor enum normalization | theme.go |
| `encoding/hex` | stdlib | theme tint/shade byte math | theme.go |
| `sort` | stdlib | deterministic styleId map iteration in cache | resolver.go (test determinism) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Dirty-flag cache invalidation | Version counter on styles.xml write | Counter survives more mutation patterns; overkill for v1 (see §Cache Invalidation) |
| Full DrawingML CT_Theme parse | Token-scan only `a:clrScheme` | Full parse = ~30 types, blows the "~60 WML types" budget; only colors needed for D-06 |
| Parallel `EffectiveProps` struct | Return `*wml.CT_PPr`/`*wml.CT_RPr` clones (D-01) | Rejected by D-01 — escape-hatch pattern wins |

**Installation:** none — stdlib only (PROJECT.md zero-dep constraint).

**Version verification:** No external packages. All stdlib APIs used (`encoding/xml` URI tags, `encoding/hex`, `strconv`, `sort`, `errors.Is`) exist since Go 1.0–1.16. Go toolchain `go1.26.5 darwin/arm64` verified in Phase 1 research. [VERIFIED: Phase 1 research file]

## Package Legitimacy Audit

**Not applicable** — this phase installs zero external packages (PROJECT.md zero-dep constraint). All dependencies are stdlib + internal packages delivered in Phase 1. No packages to audit, remove, or flag.

## Architecture Patterns

### System Architecture Diagram

```
Clone path (one-shot, no parse):
  source .docx ──► opc.Open ──► source Package
                                    │
  target .docx (fresh-empty) ──► opc.Open ──► target Package
                                    │
                                    ▼
  For each of 5 parts (styles/numbering/fontTable/theme1/settings):
     1. read source part bytes (Part.Open, capped)
     2. IF target already has that part → return ErrCloneTargetNotEmpty (D-08)
     3. target.MarkModified(partName, bytes)         ← OPC pass-through (D-09)
     4. target.ContentTypes ensure Override for the part
     5. for each source rel referencing the part →
           target.Rels["word"].NextRID() → new rId
           add Relationship{ID:new, Type:srcType, Target:partName}
                                    │
                                    ▼
  opc.Save(target) ──► target .docx (byte-identical source parts)

Resolve path (lazy, memoized):
  caller: Resolver.Resolve(paragraph|run)
                                    │
                                    ▼
  ensure styles.xml parsed (lazy: xmlutil decode → CT_Styles)  ──► cache tree
  ensure numbering.xml parsed (lazy → CT_Numbering)            ──► cache tree
  ensure theme1.xml color map built (lazy → name→hex table)   ──► cache
                                    │
                                    ▼
  pStyle := paragraph.PPr.PStyle.Val  (may be nil → docDefaults only, D-12)
                                    │
                                    ▼
  IF styleId in memo cache → clone(cached) → merge direct pPr → return
  ELSE walk chain:
     visited := {}  (styleId set, cycle guard)
     last-good := docDefaults.PPrDefault.PPr (or zero CT_PPr)
     cursor := styleId
     LOOP:
        IF cursor in visited → warn(cycle), break (D-05)
        IF cursor not in styles.Style[*] →
            IF latentStyles has lsdException(cursor) →
                last-good := merge(latent fallback props, last-good)  (D-12)
            ELSE → warn(missing styleId), break (D-07)
        ELSE:
            visited.add(cursor)
            last-good := merge(style[cursor].PPr over last-good)  (basedOn merge)
            cursor := style[cursor].BasedOn.Val  (nil → done)
     memo[styleId] := last-good
                                    │
                                    ▼
  result := clone(memo[styleId])
  IF result.NumPr != nil → merge numbering level pPr (D-13):
     numId, ilvl := result.NumPr
     lvl := numbering.AbstractNum[abstractNumIdFor(numId)].Lvl[ilvl]
     (handle CT_Num.lvlOverride.startOverride / lvlOverride.lvl if present — §17.9.17)
     merge lvl.PPr into result  (indentation etc.)
  IF result.RPr.Color.ThemeColor != nil →
     hex := theme.resolve(themeColor, themeTint, themeShade)  (D-06)
     result.RPr.Color.Val := hex; ThemeColor/Tint/Shade = nil
  IF result.PPr.PStyle exists → also resolve rStyle for the run-props path
  merge paragraph/run direct pPr/rPr on top of result  (D-04)
  return result (fresh clone — caller may mutate)
```

### Recommended Project Structure
```
wordingo/
├── internal/
│   └── style/
│       ├── cloner.go        # CloneStyles(src, dst *opc.Package) error; ErrCloneTargetNotEmpty
│       ├── resolver.go      # Resolver struct; Resolve(p *wml.CT_P|r *wml.CT_R); memo cache; cycle
│       ├── theme.go         # themeCache: parse a:clrScheme → map[enum]hex; ResolveColor(CT_Color)
│       ├── numbering.go     # numberingCache: parse CT_Numbering; ResolveLvl(numId,ilvl) → *CT_Lvl
│       └── errors.go        # ErrCloneTargetNotEmpty sentinel
└── testdata/
    ├── word/style-rich/
    │   ├── heading-chain.docx          # D-11(a): Heading4→3→2→1→Normal
    │   ├── heading-chain.expected.json  # D-10: expected CT_PPr/CT_RPr per paragraph
    │   ├── theme-refs.docx             # D-11(b): themeColor in styles + docDefaults
    │   ├── theme-refs.expected.json
    │   ├── multi-level-numbering.docx  # D-11(c): numId+ilvl across 3 levels
    │   └── multi-level-numbering.expected.json
    └── style-engine/hostile/
        ├── circular-basedon.docx        # D-05: A→B→A
        ├── circular-basedon.expected.json  # asserts last-good + warning emitted
        ├── dangling-basedon.docx        # D-07: basedOn ref to missing styleId
        ├── dangling-basedon.expected.json
        ├── missing-numid.docx           # D-07: numPr.numId not in numbering.xml
        └── missing-numid.expected.json
```

### Pattern 1: Pointer-field nil-merge semantics (carry-over from Phase 1)
**What:** All WML struct fields are pointers (Phase 1 D-07). Merge means: for each pointer field on the overriding struct, if non-nil → it *replaces* the target's pointer entirely (shallow override). Deep-merge applies *only* to a small set of composite leaves (CT_Spacing, CT_Ind, CT_RFonts, CT_Color) where OOXML semantics treat individual attributes as independently inheritable.
**When to use:** Every basedOn merge step.
**Example:**
```go
// Source: Phase 1 D-07 (pointer = nil means absent). ISO/IEC 29500-1 §17.7.2
// (style property inheritance): each property is independently inherited
// unless explicitly set on the child.
//
// Shallow-override (default for pointer fields):
//   if child.PPr.Spacing != nil → result.PPr.Spacing = clone(child.PPr.Spacing)
//   else                          → result.PPr.Spacing = result.PPr.Spacing (unchanged from chain)
//
// Deep-merge (composite leaves — per-attribute independence):
//   mergeSpacing(dst, src *CT_Spacing):
//     if src == nil { return }
//     if dst == nil { dst = &CT_Spacing{} }
//     if src.Before  != nil { dst.Before  = src.Before  }
//     if src.After   != nil { dst.After   = src.After   }
//     if src.Line    != nil { dst.Line    = src.Line    }
//     if src.LineRule != nil { dst.LineRule = src.LineRule }
//     ... (BeforeLines, AfterLines likewise)
```

### Pattern 2: Cycle detection (D-05)
**What:** Walk basedOn chain with a `visited map[string]bool`. On revisiting a styleId: append warning (`"style: circular basedOn chain at %s"`), stop walking, keep the last-good merged props accumulated *before* the cycle node was entered.
**When to use:** Every chain walk.
**Example:**
```go
// Source: D-05 (matches Phase 1 D-12 warning pattern). ECMA-376 Part 1 §17.7.4.3
// (basedOn) does not define cycle handling — Word's observed behavior is
// "stop and use what we have", which D-05 codifies.
func (r *Resolver) walkChain(styleId string, visited map[string]bool, warn func(string)) *wml.CT_PPr {
    if visited[styleId] {
        warn(fmt.Sprintf("style: circular basedOn chain at %q", styleId))
        return nil // caller keeps last-good
    }
    visited[styleId] = true
    // ... lookup, merge, recurse on basedOn ...
}
```

### Pattern 3: Theme color resolution (D-06)
**What:** Build a `map[string]string` (OOXML themeColor enum → hex) by scanning theme1.xml's `a:clrScheme` children. The enum names in `CT_Color.ThemeColor` (`dark1`/`light1`/`dark2`/`light2`/`accent1`..`accent6`/`hyperlink`/`followedHyperlink`) do **not** match the element names in theme1.xml (`dk1`/`lt1`/`dk2`/`lt2`/`accent1`..`accent6`/`hlink`/`folHlink`) — a fixed mapping table is required. Apply `themeTint`/`themeShade` per ISO §17.18.95 (themeTint) / §17.18.94 (themeShade) math.
**When to use:** At resolve-time, when `CT_Color.ThemeColor != nil`.
**Example:**
```go
// Source: ISO/IEC 29500-1 §17.3.1.29 (w:color) + §17.18.95 (themeTint) + §17.18.94 (themeShade)
// + theme1.xml element names from defaults/theme1.xml [VERIFIED: codebase]
//
// Enum → theme element name (the only such mismatch in the model — flag for tests):
var themeEnumToElement = map[string]string{
    "dark1":             "dk1",
    "light1":            "lt1",
    "dark2":             "dk2",
    "light2":            "lt2",
    "accent1":           "accent1",
    "accent2":           "accent2",
    "accent3":           "accent3",
    "accent4":           "accent4",
    "accent5":           "accent5",
    "accent6":           "accent6",
    "hyperlink":         "hlink",
    "followedHyperlink": "folHlink",
}
//
// themeTint (ST_U8Hex, 0..FF): lighten — multiply color toward white.
//   tint = parsed value / 255 (e.g. "80" → 0.5)
//   new = orig + (255 - orig) * tint   (per-channel, 0..255)
// themeShade (ST_U8Hex): darken — multiply color toward black.
//   shade = parsed value / 255 (e.g. "80" → 0.5)
//   new = orig * shade                  (per-channel)
// If both present (rare): apply shade first, then tint (Word's observed order).
// [CITED: ECMA-376 Part 1 §17.18.94/95 — tint/shade semantics; exact
//  per-channel formula is the documented Word behavior, treat as CITED
//  and verify against a fixture with known tint/shade]
```

### Pattern 4: Numbering level resolution (D-13)
**What:** Resolve `numId + ilvl` to a `CT_Lvl` by (a) finding `CT_Num` with matching `numId`, (b) reading its `abstractNumId.val`, (c) finding `CT_AbstractNum` with that id, (d) finding its `Lvl[i]` with `ilvl == requested`. Apply `lvlOverride` on the `CT_Num` if present (override `start` via `startOverride`, or replace the entire `lvl`).
**When to use:** When effective pPr has a non-nil `NumPr`.
**Example:**
```go
// Source: ISO/IEC 29500-1 §17.9.17 (w:num) + §17.9.2 (abstractNumId) + §17.9.9 (lvl)
// + §17.9.19 (numId) + §17.9.24 (startOverride) + §17.9.18 (lvlOverride)
//
// ⚠ FLAG: internal/wml/numbering.go CT_Num currently has NO lvlOverride field.
// The struct is:
//   type CT_Num struct {
//       XMLName xml.Name `xml:"... num"`
//       NumID *int64
//       AbstractNumID *CT_AbstractNumID
//       Raw []xmlutil.RawXML
//   }
// Per D-09 the cloner does NOT re-parse (byte-identity preserved), but the
// RESOLVER parses lazily and a Word numbering.xml carrying <w:lvlOverride>
// will silently lose it on resolve. Plan 02-02 MUST extend CT_Num:
//   LvlOverride []*CT_LvlOverride `xml:"... lvlOverride"`
// where:
//   type CT_LvlOverride struct {
//       XMLName xml.Name `xml:"... lvlOverride"`
//       ILvl *int64 `xml:"... ilvl,attr,omitempty"`
//       StartOverride *CT_StartOverride `xml:"... startOverride"`
//       Lvl *CT_Lvl `xml:"... lvl,omitempty"`
//   }
// [CITED: ISO/IEC 29500-1 §17.9.17, §17.9.18, §17.9.24 — struct fix required]
func (n *numberingCache) ResolveLvl(numId, ilvl int64) (*wml.CT_Lvl, error) {
    num := n.findNum(numId)
    if num == nil { return nil, errMissingNumId } // D-07 → warning
    abs := n.findAbstractNum(num.AbstractNumID.Val)
    if abs == nil { return nil, errMissingAbstractNum }
    lvl := n.findLvl(abs, ilvl)
    if lvl == nil { return nil, errMissingIlvl }
    // Apply override if present (requires the struct fix above):
    if o := n.findOverride(num, ilvl); o != nil {
        if o.Lvl != nil { lvl = o.Lvl }
        if o.StartOverride != nil && o.StartOverride.Val != nil {
            cloned := *lvl; cloned.Start = &wml.CT_Start{Val: o.StartOverride.Val}; lvl = &cloned
        }
    }
    return lvl, nil
}
```

### Pattern 5: Cloner as OPC wiring (D-08/D-09)
**What:** Clone is a 5-step loop over the source parts, each step is pure OPC API calls. No WML parse, no transformation. The cloner is ~80–120 LOC.
**When to use:** `style.CloneStyles(src, dst *opc.Package) error`
**Example:**
```go
// Source: D-09 + Phase 1 internal/opc (MarkModified, NextRID, ContentTypes).
// [VERIFIED: codebase — internal/opc/package.go MarkModified; relationships.go NextRID]
var cloneParts = []struct {
    name string    // part name in package, e.g. "word/styles.xml"
    relType string // relationship type URI for word/_rels/document.xml.rels
    ct     string  // content-type override
}{
    {"word/styles.xml",   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles",   "application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"},
    {"word/numbering.xml","http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering","application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"},
    {"word/fontTable.xml","http://schemas.openxmlformats.org/officeDocument/2006/relationships/fontTable","application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml"},
    {"word/theme/theme1.xml","http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme","application/vnd.openxmlformats-officedocument.theme+xml"},
    {"word/settings.xml", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings", "application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"},
}

func CloneStyles(src, dst *opc.Package) error {
    for _, p := range cloneParts {
        if _, exists := dst.Parts[p.name]; exists {
            return ErrCloneTargetNotEmpty   // D-08
        }
        srcPart, ok := src.Parts[p.name]
        if !ok { continue } // source lacks this part — skip (warn?)
        bytes, err := readAllCapped(srcPart)
        if err != nil { return fmt.Errorf("style clone %s: %w", p.name, err) }
        dst.MarkModified(p.name, bytes)
        // Content type:
        dst.ContentTypes.EnsureOverride(p.name, p.ct)
        // Relationship in word/_rels/document.xml.rels:
        wordRels := dst.Rels["word"]   // or "" if package-root rels
        if wordRels == nil { wordRels = &Relationships{}; dst.Rels["word"] = wordRels }
        rid := wordRels.NextRID()
        wordRels.Add(Relationship{ID: rid, Type: p.relType, Target: p.name})
    }
    return nil
}
```

### Pattern 6: basedOn vs link vs next — which to walk
**What:** Three independent style cross-references exist on `CT_Style`:
- `w:basedOn` (§17.7.4.3): **walk for property inheritance** — child style inherits parent's pPr + rPr.
- `w:next` (§17.7.4.10): **editor hint only** — "after Enter on this paragraph style, apply this next paragraph style." **Never walked by resolver.**
- `w:link` (§17.7.4.6): **pairs a paragraph style with a character style** — when resolving a linked paragraph style's run properties, the linked character style's rPr may contribute. For v1, treat `link` as: when resolving rPr of a paragraph style that has a `link` to a character style, also consult that character style's rPr as an additional layer (lower priority than the paragraph style's own rPr, higher than basedOn parent's rPr). If this proves ambiguous, fall back to "ignore link for v1" and document — bounded scope.
**When to use:** Every chain walk.
**Pitfall:** Implementing `next` walking is the #1 resolver bug; produces wrong effective props because it pulls in unrelated styles.

### Anti-Patterns to Avoid
- **Walking `next` chain:** editor hint, not inheritance. Never consult for props.
- **Returning the cached merged props directly (no clone):** violates D-01/D-03; caller mutation corrupts the cache. Always `clone(memo[styleId])` before returning.
- **Full `CT_Theme` parse:** DrawingML is out of scope; only `a:clrScheme` children needed. A token scan of theme1.xml bytes is sufficient and stays within "~60 WML types" budget.
- **Cache invalidation by hashing styles.xml on every Resolve:** O(part size) per call defeats memoization. Use a dirty flag (§Cache Invalidation).
- **Treating `latentStyles` as a primary layer:** D-12 — latentStyles is fallback ONLY when an explicit styleId is missing from styles.xml. Unstyled paragraphs (no pStyle) skip it entirely.
- **String-equal themeColor → theme element name:** the enum and element name spaces differ (see Pattern 3); a direct lookup misses every dark/light/hyperlink color.
- **Re-parsing numbering.xml on every Resolve:** cache it lazily once on first numbering query; invalidate together with styles.xml (numbering.xml rarely mutated in v1, but if it is, both caches drop).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| OPC byte copy | custom zip read/write | `opc.Package.MarkModified` + `opc.Save` | Phase 1 already byte-exact; cloner is wiring |
| rId allocation | custom counter | `opc.Relationships.NextRID` | Phase 1 OPC-06 high-water-mark, no reuse |
| Content-type registry | custom map | `opc.ContentTypes` (Phase 1) | Handles Defaults + Overrides, canonical serialize |
| XML namespace handling | prefix-based tags | `xmlutil` URI-form decoder (Phase 1) | Producer-agnostic |
| WML type definitions | new EffectiveProps struct | existing `*wml.CT_PPr` / `*wml.CT_RPr` (D-01) | Reuse Phase 1; escape-hatch pattern |
| ZIP bomb / entity guards | new limits | Phase 1 OPC-07 + xmlutil Strict decoder | Already enforced |

**Key insight:** Phase 2's only genuinely custom code is (a) the chain walker + merge rules, (b) theme color name-map + tint/shade math, (c) numbering level lookup with lvlOverride. Everything else is composition of Phase 1 primitives. Estimate: resolver.go ~250 LOC, theme.go ~120 LOC, numbering.go ~150 LOC, cloner.go ~100 LOC, errors.go ~10 LOC → **~630 LOC total**, comparable to Phase 1's xmlutil budget.

## Runtime State Inventory

> Phase 2 is partially a refactor/extension phase (extends `CT_Num` struct) and partially greenfield (`internal/style` package). The struct extension has runtime-state implications.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — no DB/keystore/cache outside the in-process `opc.Package` | None — verified by codebase grep (no `user_id`, no Redis, no ChromaDB in this Go library) |
| Live service config | None — no external services configured via UI/DB | None |
| OS-registered state | None — no launchd/systemd/Task Scheduler entries | None |
| Secrets/env vars | None — no SOPS keys, no .env files in repo | None |
| Build artifacts | `CT_Num` struct extension is source-only — Go rebuilds on `go test`, no egg-info / global installs. **Existing `defaults/numbering.xml`-style fixtures do NOT use `lvlOverride`** (verified: `defaults/` has no numbering.xml) — so no fixture regenerates. | None beyond source edit + `go test` rebuild |

**Nothing found in category:** "Stored data", "Live service config", "OS-registered state", "Secrets/env vars" — all verified by `ls defaults/` + codebase grep. The only state-bearing change is the `CT_Num` struct field addition, which is a code edit (no data migration).

## Common Pitfalls

### Pitfall 1: Walking `next` instead of (or in addition to) `basedOn`
**What goes wrong:** `w:next` (§17.7.4.10) is an editor-cursor hint ("after this paragraph, apply this style"). Implementers often treat it as a second inheritance chain. Effective props become polluted with unrelated styles.
**Why it happens:** The three references (`basedOn`, `next`, `link`) all carry `val="StyleID"` and look structurally identical.
**How to avoid:** Resolver consults `basedOn` only. `next` is never read by the resolver. `link` is consulted only for run-props on linked paragraph styles (and even then, optionally — see Pattern 6).
**Warning signs:** Effective props for "Normal" suddenly include Heading1's spacing; expected-values equality tests fail across the board.

### Pitfall 2: latentStyles treated as primary layer
**What goes wrong:** Resolver consults latentStyles for every paragraph, even unstyled ones. Effective props for plain paragraphs get unexpected "Heading1"-style defaults.
**Why it happens:** The OOXML spec describes latentStyles as "default formatting for styles not yet instantiated" — easy to misread as "always applied."
**How to avoid:** D-12 — latentStyles consulted ONLY when an explicit `pStyle` references a styleId missing from `styles.xml`. Unstyled paragraphs (no pStyle) use docDefaults only.
**Warning signs:** Unstyled paragraph's effective props contain values not present in docDefaults.

### Pitfall 3: themeColor enum ≠ theme1.xml element name
**What goes wrong:** Direct `map[enum]hex` lookup keyed on the enum string fails for `dark1`/`light1`/`dark2`/`light2`/`hyperlink`/`followedHyperlink` because theme1.xml uses `dk1`/`lt1`/`dk2`/`lt2`/`hlink`/`folHlink`. Resolver returns the wrong hex or a "missing theme color" warning.
**How to avoid:** Pattern 3 — fixed enum→element-name table. Test with a fixture whose styles reference `themeColor="dark1"` and `themeColor="hyperlink"`.
**Warning signs:** Theme color resolution passes for `accent1`..`accent6` (names match) but fails for dark/light/hyperlink.

### Pitfall 4: `CT_Num` missing `lvlOverride` field
**What goes wrong:** Word-authored numbering.xml files commonly use `<w:num><w:lvlOverride ilvl="0"><w:startOverride val="1"/></w:lvlOverride></w:num>` to restart numbering. The current `CT_Num` struct (verified: `internal/wml/numbering.go` lines 26–31) has no `lvlOverride` field — the override is hoarded into `Raw` and silently lost on resolve.
**How to avoid:** Plan 02-02 MUST extend `CT_Num` with `LvlOverride []*CT_LvlOverride` and define `CT_LvlOverride` + `CT_StartOverride`. The cloner (D-09 byte pass-through) is unaffected; only the resolver path needs the fix.
**Warning signs:** Numbered list in cloned+resolved document restarts at wrong value (e.g., 2 instead of 1).

### Pitfall 5: theme tint/shade math applied in wrong order or wrong direction
**What goes wrong:** When both `themeTint` and `themeShade` are present, applying tint-then-shade vs shade-then-tint gives different hex. Also, the formulas lighten/darken in opposite directions — swapping them inverts the effect.
**Why it happens:** OOXML §17.18.94/95 define the semantics but the per-channel formula is documented in terms of "Word's behavior," not a crisp spec formula.
**How to avoid:** Pattern 3 — apply shade first, then tint (Word's observed order). Verify against a fixture with known tint+shade values. Tag this as CITED (not VERIFIED) and confirm with a fixture before locking.
**Warning signs:** Resolved color is darker when it should be lighter, or vice versa.

### Pitfall 6: `fontTable.xml` consulted for font name resolution beyond `rFonts`
**What goes wrong:** Implementer reads fontTable.xml to "resolve" font names referenced in `CT_RFonts` (ascii/hAnsi/eastAsia/cs). Word does NOT do this — `CT_RFonts` carries literal font names; fontTable.xml is metadata (panose, sig, charset) for Word's font substitution engine, not a name-lookup table.
**How to avoid:** Resolver treats `CT_RFonts` as opaque string fields — no fontTable consultation for name resolution. fontTable.xml is cloned (STYLE-CLONE-01) but not parsed by the resolver in v1. `asciiTheme`/`hAnsiTheme` attributes (e.g. `majorHAnsi`, `minorHAnsi`) DO require theme consultation — look up `fontScheme/majorFont/latin@typeface` or `minorFont/latin@typeface` in theme1.xml.
**Warning signs:** Resolver tries to "resolve" "Calibri" via fontTable and returns panose bytes.

### Pitfall 7: Pointer-field shallow-merge where deep-merge is required (carry-over from Phase 1 D-07)
**What goes wrong:** `child.PPr.Spacing` non-nil → replace `result.PPr.Spacing` entirely. But OOXML treats `w:spacing` attributes (`before`, `after`, `line`, `lineRule`, `beforeLines`, `afterLines`) as independently inheritable — if child sets only `before`, parent's `after` should survive.
**How to avoid:** Pattern 1 — deep-merge for `CT_Spacing`, `CT_Ind`, `CT_RFonts`, `CT_Color` (the four composite leaves where per-attribute independence matters). Shallow-override for everything else.
**Warning signs:** Expected-values equality fails on `Spacing.After` for a Heading2 fixture where Heading1 sets `after` and Heading2 sets only `before`.

### Pitfall 8: Returning cached merged props without cloning
**What goes wrong:** Two callers of `Resolve()` share the same `*CT_PPr` pointer; one mutates it, the other sees corruption.
**How to avoid:** D-01/D-03 — `Resolve()` always returns a deep clone. Memo holds the merged chain; every return path does `clone(memo[styleId])` then merges direct formatting on top.
**Warning signs:** Tests pass in isolation, fail when run together (shared state bleed).

### Pitfall 9: `latentStyles` `lsdException` `locked`/`semiLocked` semantics misread
**What goes wrong:** Implementer treats `locked="true"` as "do not consult this latent style." Wrong — `locked` governs UI visibility in Word's style picker, not resolver consultation. The resolver consults latent styles regardless of `locked` (per D-12 fallback rule).
**How to avoid:** Read `lsdException` for `name` match only; ignore `locked`/`semiHidden`/`qFormat`/`uiPriority` fields for resolution purposes.
**Warning signs:** Resolver skips a latent style fallback because it's "locked," producing wrong effective props.

## Code Examples

### basedOn chain walk (canonical Heading2 → Heading1 → Normal)
```go
// Source: ISO/IEC 29500-1 §17.7.4.3 (basedOn) + §17.7.2 (inheritance order)
// + D-05 (cycle) + D-03 (memo) + D-07 (missing ref)
func (r *Resolver) resolvePPr(styleId string, warn func(string)) *wml.CT_PPr {
    if cached, ok := r.memoPPr[styleId]; ok {
        return clonePPr(cached)
    }
    result := &wml.CT_PPr{}
    // 1. docDefaults (D-12 base layer; always applied)
    if dd := r.styles.DocDefaults; dd != nil && dd.PPrDefault != nil && dd.PPrDefault.PPr != nil {
        result = mergePPr(result, dd.PPrDefault.PPr)
    }
    // 2. Walk basedOn chain bottom-up (child overrides parent)
    visited := map[string]bool{}
    chain := r.collectChain(styleId, visited, warn) // [child, parent, grandparent, ...]
    // apply in reverse (root first, child last)
    for i := len(chain) - 1; i >= 0; i-- {
        s := chain[i]
        if s.PPr != nil { result = mergePPr(result, s.PPr) }
    }
    r.memoPPr[styleId] = result
    return clonePPr(result)
}

func (r *Resolver) collectChain(styleId string, visited map[string]bool, warn func(string)) []*wml.CT_Style {
    var chain []*wml.CT_Style
    cursor := styleId
    for cursor != "" {
        if visited[cursor] {
            warn(fmt.Sprintf("style: circular basedOn chain at %q", cursor))
            break
        }
        visited[cursor] = true
        s := r.findStyle(cursor)
        if s == nil {
            // D-12: latentStyles fallback only when styleId missing from styles.xml
            if ls := r.findLatent(cursor); ls != nil {
                warn(fmt.Sprintf("style: %q not instantiated; using latentStyles fallback", cursor))
                // latent styles carry no pPr/rPr of their own in the lsdException;
                // they signal "use docDefaults" — already applied. Break.
            } else {
                warn(fmt.Sprintf("style: %q not found in styles.xml or latentStyles", cursor))
            }
            break
        }
        chain = append(chain, s)
        if s.BasedOn != nil && s.BasedOn.Val != nil { cursor = *s.BasedOn.Val } else { cursor = "" }
    }
    return chain
}
```

### Deep-merge for composite leaves
```go
// Source: ISO/IEC 29500-1 §17.7.2 (per-attribute inheritance) + Phase 1 D-07
func mergeSpacing(dst, src *wml.CT_Spacing) *wml.CT_Spacing {
    if src == nil { return dst }
    if dst == nil { dst = &wml.CT_Spacing{} }
    if src.Before     != nil { dst.Before = src.Before }
    if src.After      != nil { dst.After = src.After }
    if src.Line       != nil { dst.Line = src.Line }
    if src.LineRule   != nil { dst.LineRule = src.LineRule }
    if src.BeforeLines != nil { dst.BeforeLines = src.BeforeLines }
    if src.AfterLines  != nil { dst.AfterLines = src.AfterLines }
    return dst
}
// Apply same pattern to CT_Ind (Left/Right/FirstLine/Hanging),
// CT_RFonts (Ascii/HAnsi/EastAsia/CS/AsciiTheme/HAnsiTheme),
// CT_Color (Val/ThemeColor/ThemeShade/ThemeTint — but themeColor resolution
// happens AFTER merge, at the top of the resolved tree, per D-06).
```

### Theme color resolution entry point
```go
// Source: ISO/IEC 29500-1 §17.3.1.29 (w:color) + §17.18.95 (themeTint) + §17.18.94 (themeShade)
// + D-06 (resolve at resolve-time, raw theme untouched)
func (t *themeCache) ResolveColor(c *wml.CT_Color, warn func(string)) *wml.CT_Color {
    if c == nil { return nil }
    out := *c // shallow copy
    if c.ThemeColor != nil && *c.ThemeColor != "" {
        elemName, ok := themeEnumToElement[*c.ThemeColor]
        if !ok {
            warn(fmt.Sprintf("theme: unknown themeColor %q", *c.ThemeColor))
            return &out // D-07: leave as-is, sensible default = absent hex
        }
        hex, ok := t.colors[elemName]
        if !ok {
            warn(fmt.Sprintf("theme: %q not found in theme1.xml", elemName))
            return &out
        }
        r, g, b := parseHexRGB(hex)
        if c.ThemeShade != nil { r, g, b = applyShade(r, g, b, *c.ThemeShade) }
        if c.ThemeTint != nil  { r, g, b = applyTint(r, g, b, *c.ThemeTint) }
        v := fmt.Sprintf("%02X%02X%02X", r, g, b)
        out.Val = &v
        out.ThemeColor = nil; out.ThemeShade = nil; out.ThemeTint = nil
    }
    return &out
}
```

## Cache Invalidation Mechanism (D-03 — discretion)

**Recommendation: dirty flag on `word/styles.xml` mutation.**

Two options were considered:

| Option | Mechanism | Pros | Cons |
|--------|-----------|------|------|
| **A. Dirty flag** (recommended) | `Package` exposes a `MutatedParts map[string]bool` (or a `Part.dirty` bool already set by `MarkModified`). Resolver checks `dst.Parts["word/styles.xml"].modified` at entry to `Resolve()`. If true → drop `memoPPr` + `memoRPr` maps, re-walk on next call. | Trivial to implement (~5 LOC); reuses Phase 1's `Part.modified` field; no new state. | Invalidation is coarse (any styles.xml mutation drops the whole cache, even unrelated styleIds). Acceptable for v1 — style.xml mutations are rare. |
| B. Version counter | `Package` carries `stylesVersion uint64`; `MarkModified("word/styles.xml", ...)` increments it; resolver stores `cachedVersion` per styleId and re-walks only stale entries. | Fine-grained; preserves cached entries for unchanged styleIds. | More state, more bugs; overkill for v1 where styles.xml is rarely mutated mid-session. |

**Decision:** Option A. The resolver's `Resolve()` entry checks `pkg.Parts["word/styles.xml"].modified`; if true, clears both memo maps and proceeds with a fresh walk. Same check applies to `word/numbering.xml` for the numbering cache. ~10 LOC, reuses existing Phase 1 state.

**Phase 1 hook required:** `opc.Part.modified` is already set by `MarkModified` (verified: `internal/opc/package.go` line 120). The resolver reads it directly — no new opc API needed. If Phase 1's `Part.modified` is unexported or inaccessible, expose a trivial `Part.IsModified() bool` accessor (one-liner).

## Fixture Design (D-10/D-11)

### Expected-values JSON schema (discretion — proposed)
```json
{
  "template": "heading-chain.docx",
  "description": "Heading4→3→2→1→Normal deep basedOn chain with theme colors",
  "cases": [
    {
      "id": "h2-para-1",
      "locator": "paragraph[1]",
      "pStyle": "Heading2",
      "expected": {
        "pPr": {
          "spacing": {"before": 240, "after": 120, "line": 0, "lineRule": null},
          "ind": null,
          "jc": null,
          "numPr": null,
          "keepNext": true,
          "outlineLvl": 1
        },
        "rPr": {
          "rFonts": {"asciiTheme": "majorHAnsi", "ascii": null},
          "sz": {"val": 28},
          "color": {"val": "1F4E79", "themeColor": null},
          "b": true,
          "i": null
        }
      },
      "warnings": []
    },
    {
      "id": "h2-circular",
      "locator": "paragraph[2]",
      "pStyle": "CycleA",
      "expected": {
        "pPr": {"spacing": {"before": 0, "after": 0}},
        "rPr": {}
      },
      "warnings": ["style: circular basedOn chain at \"CycleA\""]
    }
  ]
}
```

**Schema rules:**
- `null` means "field is absent (nil pointer) in effective props" — distinct from `0` or `""` which mean "present with zero value."
- `warnings` array asserts exact warning substrings emitted by the resolver for this case (D-05/D-07 hostile paths).
- `locator` is human-readable; the test harness finds the paragraph by index in document.xml or by a bookmark marker.

### Canonical Heading2 case (ROADMAP success criterion #4)
Fixture `heading-chain.docx` must contain a paragraph with `pStyle="Heading2"` where:
- `Heading2` style: `basedOn="Heading1"`, `spacing before=240 after=120`, `rPr sz=28 color themeColor="accent1"`, `outlineLvl=1`, `keepNext`
- `Heading1` style: `basedOn="Normal"`, `spacing before=480 after=240`, `rPr sz=32 b=true color themeColor="accent1"`, `outlineLvl=0`
- `Normal` style: `basedOn` absent, `rPr sz=22 rFonts asciiTheme="minorHAnsi"`
- docDefaults: `rPrDefault rPr sz=22 rFonts ascii="Aptos"`, `pPrDefault pPr spacing after=160`

**Expected effective pPr for the Heading2 paragraph:** `spacing before=240 after=120` (Heading2 overrides Heading1+Normal+docDefaults), `outlineLvl=1`, `keepNext=true`.
**Expected effective rPr:** `sz=28` (Heading2), `b=true` (Heading1), `color val=<accent1 hex from theme1.xml>`, `rFonts asciiTheme="majorHAnsi"` (Heading1), `ascii="Aptos"` survives from docDefaults only if Heading1/2 don't set `ascii` — verify fixture captures this subtlety.

### Hostile fixture coexistence (discretion)
**Recommendation:** sibling directory `testdata/style-engine/hostile/`. Rationale:
- Keeps `testdata/word/style-rich/*.expected.json` clean for equality assertions (no warning-string noise mixed in).
- Hostile fixtures have a different assertion shape (warnings expected, props are "best-effort") — separate dir signals separate test helper.
- Phase 1's `testdata/word/`, `testdata/libreoffice/`, `testdata/googledocs/`, `testdata/hostile/` pattern already uses producer-segmented dirs; `style-engine/` continues the theme.

### Testdata layout summary
```
testdata/
├── word/style-rich/          # D-10/D-11 happy-path fixtures + expected.json
│   ├── heading-chain.docx
│   ├── heading-chain.expected.json
│   ├── theme-refs.docx
│   ├── theme-refs.expected.json
│   ├── multi-level-numbering.docx
│   └── multi-level-numbering.expected.json
└── style-engine/hostile/    # D-05/D-07 warning-path fixtures
    ├── circular-basedon.docx
    ├── circular-basedon.expected.json
    ├── dangling-basedon.docx
    ├── dangling-basedon.expected.json
    ├── missing-numid.docx
    └── missing-numid.expected.json
```

## Validation Architecture

> `workflow.nyquist_validation` is explicitly `false` in `.planning/config.json`, so the formal Nyquist test-map section is omitted. The narrative below grounds the planner — these are the assertion strategies the implementation tasks must encode, not a formal Nyquist table.

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` (stdlib) — Phase 1 baseline |
| Config file | none (Go convention) |
| Quick run command | `go test ./internal/style/...` |
| Full suite command | `go test ./...` |

### Assertion Strategies (Phase 2-specific)

| Behavior | Assertion Strategy |
|----------|-------------------|
| Effective CT_PPr equality (STYLE-RESOLVE-01) | Field-by-field nil-check: for each pointer field on `CT_PPr`, assert `expected == actual` where `nil` ≠ `&CT_OnOff{}`. Use `reflect.DeepEqual` on the cloned result after normalizing nil-vs-zero. Compare against `expected.json` fixture. |
| Effective CT_RPr equality (STYLE-RESOLVE-02) | Same field-by-field over `CT_RPr`. Theme color must be concrete hex (`val != nil && themeColor == nil`). |
| Theme color resolution (STYLE-RESOLVE-03) | Assert `result.RPr.Color.Val != nil` and `result.RPr.Color.ThemeColor == nil` when input had `themeColor="accent1"`. Assert exact hex matches `theme1.xml` accent1 value. |
| Cycle detection (D-05) | Assert `doc.Warnings()` contains `"circular basedOn chain at \"CycleA\""`. Assert returned props equal the last-good chain prefix (compare against `expected.json` which encodes the prefix). |
| Missing styleId (D-07) | Assert warning emitted. Assert returned props equal docDefaults (the sensible default). |
| Missing numId (D-07) | Assert warning emitted. Assert `result.NumPr` is preserved as-is OR dropped per discretion (recommend: preserved, warning emitted — the caller decides whether to render the broken numPr). |
| Clone correctness — byte pass-through (STYLE-CLONE-01) | Open source .docx + fresh-empty target → `style.CloneStyles(src, dst)` → `dst.Save(buf)` → unzip both → per-part byte diff on the 5 cloned parts (Phase 1 D-04 pattern). |
| Clone correctness — relationships | Assert `dst.Rels["word"]` contains 5 relationships with correct `Type` URIs and `Target` matching source part names. Assert rIds are unique (Phase 1 OPC-06). |
| Clone correctness — content types | Assert `dst.ContentTypes` has Overrides for all 5 parts with correct MIME types (see Pattern 5 table). |
| Clone conflict (D-08) | Pre-populate target with `word/styles.xml` → `CloneStyles` returns `errors.Is(err, style.ErrCloneTargetNotEmpty)`. |
| lvlOverride handling (Pitfall 4) | Fixture with `<w:num><w:lvlOverride ilvl="0"><w:startOverride val="5"/></w:lvlOverride></w:num>` → assert resolved `start` for ilvl 0 is 5, not the abstractNum's start. |

### Wave 0 Gaps
- [ ] `internal/style/resolver_test.go` — covers STYLE-RESOLVE-01..03 + cycle + missing-ref
- [ ] `internal/style/cloner_test.go` — covers STYLE-CLONE-01 + D-08 conflict
- [ ] `internal/style/theme_test.go` — covers theme color enum map + tint/shade math
- [ ] `internal/style/numbering_test.go` — covers lvlOverride + numId+ilvl lookup
- [ ] `internal/style/testdata_test.go` — loads `expected.json` fixtures, runs table-driven cases
- [ ] `testdata/word/style-rich/*.docx` + `*.expected.json` — D-10/D-11 happy-path corpus (user-authored per Phase 1 D-02 pattern)
- [ ] `testdata/style-engine/hostile/*.docx` + `*.expected.json` — D-05/D-07 hostile corpus
- [ ] `internal/wml/numbering.go` — extend `CT_Num` with `LvlOverride` field (Pitfall 4)

*(If no gaps: "None — existing test infrastructure covers all phase requirements") — gaps exist, listed above.*

## Security Domain

> `security_enforcement` is `true` in config.json. Phase 2 has minimal new security surface (no new I/O, no new external input) — most controls are inherited from Phase 1.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | library, no auth surface |
| V3 Session Management | no | library, no sessions |
| V4 Access Control | no | library, no authz |
| V5 Input Validation | yes | Cloner validates target-empty precondition (D-08); resolver validates styleId/numId/ilvl against parsed trees; theme tint/shade parsed as hex with length validation |
| V6 Cryptography | no | no crypto in Phase 2 (theme hex math is arithmetic, not crypto) |
| V12 File Handling | yes (inherited) | Cloner reuses Phase 1 OPC-07 limits (decompression caps, path traversal rejection on target part names); no new file surface |

### Known Threat Patterns for Go .docx style engine

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Malicious template with circular basedOn chain (DoS via infinite walk) | DoS | D-05 visited-set cycle detection — chain walk terminates in O(n) where n = number of styles |
| Malicious template with pathological basedOn depth (stack overflow) | DoS | Add a max-chain-depth guard (e.g., 64) — same visited-set mechanism, break with warning if exceeded. Reuse Phase 1 OPC-07 max-XML-depth philosophy. |
| Malicious numbering.xml with self-referencing abstractNumId | Tampering/DoS | Resolver validates abstractNumId lookup terminates; missing → D-07 warning + no numPr |
| Theme1.xml with malformed hex values | Tampering | `parseHexRGB` rejects non-hex / wrong-length strings → D-07 warning + absent hex (sensible default) |
| Target document path traversal via cloned part names | Tampering | Phase 1 OPC-07 `validatePartName` already rejects `..`/absolute/backslash; cloner uses fixed part-name constants (Pattern 5), not user-supplied names |

**No new sentinels required** beyond `ErrCloneTargetNotEmpty`. Cycle/missing-ref warnings surface via the Phase 1 `Warnings()` pattern.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Theme color left as placeholder until render-time | Resolve to concrete hex at resolve-time (D-06) | Phase 2 decision | Returned CT_Color is render-ready; raw theme1.xml stays untouched for Phase 3 byte-identity |
| Full CT_Theme parse | Token-scan only `a:clrScheme` | Phase 1 decision (theme kept raw) | Stays within "~60 WML types" budget; theme.go is ~120 LOC |
| Per-call chain walk | Memoize by styleId (D-03) | Phase 2 decision | O(1) repeated resolves for same styleId; cache invalidates on styles.xml mutation |
| Merge-by-styleId clone | Fresh-empty-target only (D-08) | Phase 2 decision | Simpler v1; merge deferred to future phase |

**Deprecated/outdated:**
- `FileHeader.ModifiedTime/ModifiedDate`, `SetModTime` (Phase 1) — N/A in Phase 2 (no new zip writes; cloner reuses `opc.Save`).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | ISO/IEC 29500-1 §17.7.4.3 (basedOn) defines single-parent inheritance; §17.7.4.10 (next) is editor hint only; §17.7.4.6 (link) pairs para/char styles | Pattern 6 | If `next` carries inheritance semantics, resolver misses props — but Word's observed behavior confirms it does not. Verify with a fixture that has `next` pointing to a different style and assert next's props are absent. |
| A2 | themeTint/themeShade per-channel formula: `new = orig + (255-orig)*tint` and `new = orig*shade`, applied shade-first-then-tint | Pattern 3, Pitfall 5 | If formula differs, resolved hex mismatches by a few units. Verify with a fixture carrying known tint/shade values. |
| A3 | `latentStyles.lsdException.locked` governs UI visibility, not resolver consultation | Pitfall 9 | If `locked` should block consultation, some latent fallbacks would be skipped. Low risk — D-12 is explicit that latentStyles is fallback regardless. |
| A4 | `CT_Num` extension with `LvlOverride` field is backward-compatible (existing fixtures without lvlOverride parse identically) | Pitfall 4 | If `xml:",any"` Raw hoarding captured lvlOverride, the field addition moves it from Raw to typed — backward compatible by construction. Verify with Phase 1 numbering.xml test fixtures. |
| A5 | `opc.Part.modified` field is accessible from `internal/style` (same module) for cache invalidation | Cache Invalidation | If unexported/inaccessible, add a `Part.IsModified() bool` accessor — one-line Phase 1 patch. Low risk. |
| A6 | `link` semantics: consult linked character style's rPr when resolving a linked paragraph style's run props | Pattern 6 | If Word does not consult link for rPr, resolver over-includes. Bounded scope — if ambiguous, drop link consultation for v1 and document. |
| A7 | Heading2 canonical case expected values (sz=28, accent1 theme color, etc.) | Fixture Design | Fixture-author (user per Phase 1 D-02) captures exact values from a real Word fixture; researcher's numbers are illustrative. |
| A8 | Theme1.xml element names are `dk1/lt1/dk2/lt2/accent1-6/hlink/folHlink` (verified against `defaults/theme1.xml`) but enum values are `dark1/light1/dark2/light2/accent1-6/hyperlink/followedHyperlink` | Pattern 3 | Verified for element names; enum values from ISO §17.18.95 ST_ThemeColor — CITED, not re-pulled from spec text this session. |

## Open Questions

1. **Should `link` be consulted for run-props in v1?**
   - What we know: ISO §17.7.4.6 pairs paragraph↔character styles; Word consults the linked character style when resolving a paragraph style's rPr.
   - What's unclear: whether the merge order is (para-style rPr → linked char-style rPr → basedOn chain rPr) or (linked char-style rPr → para-style rPr → basedOn chain rPr).
   - Recommendation: implement link consultation as a layer between para-style-rPr and basedOn-chain-rPr. If tests reveal ambiguity, drop link for v1 and document — bounded scope, no Phase 2 blocker.

2. **Should missing numId preserve the broken `CT_NumPr` or drop it?**
   - What we know: D-07 says "substitute a sensible default (no numPr)." But the caller (Phase 4 content API) may want to know the numPr was broken.
   - Recommendation: preserve the numPr as-is in the returned clone, emit a warning. Caller decides whether to render. Document this in resolver.go.

3. **Does `opc.Part.modified` get exposed cleanly to `internal/style`?**
   - What we know: `Part.modified` is set by `MarkModified` (Phase 1, verified).
   - What's unclear: whether the field is exported or has an accessor.
   - Recommendation: plan 02-01 task verifies; if unexported, add `Part.IsModified() bool` one-liner.

## Environment Availability

> Phase 2 has no new external dependencies beyond Phase 1's. All work is Go stdlib + internal packages.

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | everything | ✓ | go1.26.5 darwin/arm64 (Phase 1 verified) | — |
| `internal/opc` | cloner | ✓ | Phase 1 delivered | — |
| `internal/wml` | resolver | ✓ (with `CT_Num` extension needed) | Phase 1 delivered | — |
| `internal/xmlutil` | resolver parse path | ✓ | Phase 1 delivered | — |
| `defaults/theme1.xml` | theme color reference values | ✓ | Phase 1 delivered | — |
| Word 2016/2019/2021/M365 | fixture authoring (D-02 pattern) + manual sign-off | ✗ (not on this machine) | — | user-authored fixtures + manual gate (Phase 1 D-01 pattern) |

**Missing dependencies with no fallback:** Word for fixture authoring — explicitly accepted as user-performed step (Phase 1 D-01 pattern), not a blocker for implementation.

**Missing dependencies with fallback:** none.

## Integration Points

### `internal/style/` package shape (discretion — recommended split)
```
internal/style/
├── cloner.go      # CloneStyles(src, dst *opc.Package) error
├── resolver.go    # type Resolver struct; NewResolver(pkg *opc.Package) *Resolver
│                  #   func (r *Resolver) ResolvePPr(p *wml.CT_P) *wml.CT_PPr
│                  #   func (r *Resolver) ResolveRPr(p *wml.CT_P, r *wml.CT_R) *wml.CT_RPr
│                  #   (single entry point per D-04; paragraph-aware for run-style resolution)
├── theme.go       # type themeCache struct; buildColorMap(theme1Bytes []byte)
│                  #   func (t *themeCache) ResolveColor(c *wml.CT_Color, warn) *wml.CT_Color
├── numbering.go   # type numberingCache struct; buildLevelMap(numberingBytes []byte)
│                  #   func (n *numberingCache) ResolveLvl(numId, ilvl int64, warn) *wml.CT_Lvl
└── errors.go      # var ErrCloneTargetNotEmpty = errors.New("style: clone target not empty")
```

**Rationale for 4-file split (not merged):**
- `cloner.go` is pure OPC wiring, no parse logic — isolates the byte-copy path from the resolver.
- `resolver.go` is the chain walker + memo cache + cycle detection — the most complex logic, deserves its own file.
- `theme.go` and `numbering.go` are independent parse-on-demand caches with distinct concerns (color math vs level lookup). Splitting keeps each under ~150 LOC.
- `errors.go` matches Phase 1 D-11 sentinel pattern.

**Alternative considered:** single `style.go` file (~630 LOC). Rejected — exceeds comfortable review size and mixes concerns.

### Strict layering (PRD §9)
```
wordingo/            (Phase 4 — Document wrapper calls style.Resolver)
  └── internal/style/   (Phase 2 — depends on opc, wml, xmlutil)
        ├── internal/opc/     (Phase 1 — cloner uses MarkModified/NextRID/ContentTypes)
        ├── internal/wml/     (Phase 1 — resolver returns *CT_PPr/*CT_RPr; CT_Num extended)
        └── internal/xmlutil/ (Phase 1 — resolver parses styles.xml/numbering.xml/theme1.xml)
```
`internal/style` depends on all three Phase 1 internal packages. No reverse dependencies. PRD §9: "style depends on both [opc and wml]; public API orchestrates all three."

### Phase 3 integration (STYLE-ROUNDTRIP, FromTemplate)
- **Phase 3 `FromTemplate(path)` (CREATE-03):** calls `style.CloneStyles(templatePkg, newPkg)` to seed a new document from a template. Clone is the entire style-seeding step — Phase 3 wraps it with body initialization.
- **Phase 3 STYLE-ROUNDTRIP-01/02:** the cloner's byte pass-through (D-09) is the *foundation* for round-trip fidelity — unmodified style parts survive byte-identical because the cloner never parses them. Phase 3 adds the "editing never rewrites style parts unless user modifies styles" guarantee on top.
- **Phase 3 byte-identity:** applies to raw parts (theme1.xml, styles.xml, etc.) — NOT to resolved clones (D-06 explicitly notes this). The resolver returns fresh CT_PPr/CT_RPr objects; the raw theme1.xml bytes in the package are untouched.

### Phase 4 integration (content API)
- **`Paragraph.SetStyle("Heading1")` (API-03):** sets `CT_P.PPr.PStyle = &CT_PStyle{Val: &"Heading1"}`. Does NOT call the resolver directly — the resolver is called when the caller (or the renderer, or a debug tool) asks for effective props. The styleId is the *reference*; resolution is on-demand.
- **`Run.SetBold(true)` (API-01):** sets `CT_R.RPr.B = &CT_OnOff{Val: &true}`. When effective props are queried, the resolver merges this direct formatting on top of the style chain (D-04).
- **`Document` wrapper (PRD §8):** holds `*opc.Package` + `*style.Resolver`. Exposes `Warnings()` (merges opc + style warnings). Phase 2 delivers the Resolver; Phase 4 wires it into Document.

## Sources

### Primary (HIGH confidence)
- `internal/wml/styles.go` — CT_Styles/CT_Style/CT_DocDefaults/CT_RPrDefault/CT_PPrDefault/CT_LatentStyles/CT_LsdException/CT_BasedOn/CT_Next/CT_Link — verified this session
- `internal/wml/numbering.go` — CT_Numbering/CT_AbstractNum/CT_Num/CT_AbstractNumID/CT_Lvl/CT_NumFmt/CT_LvlText/CT_Start — verified; **CT_Num lvlOverride field ABSENT (flagged Pitfall 4)**
- `internal/wml/properties.go` — CT_PStyle/CT_RStyle/CT_NumPr/CT_NumId/CT_ILvl/CT_Color/CT_RFonts/CT_Spacing/CT_Ind — verified
- `internal/wml/document.go` — CT_PPr/CT_RPr field lists — verified
- `internal/opc/package.go` — Package.MarkModified, Part.modified field, Package.Save, Package.Warnings — verified
- `internal/opc/relationships.go` — Relationships.NextRID, validate — verified
- `defaults/theme1.xml` — confirmed `a:clrScheme` children: dk1/lt1/dk2/lt2/accent1-6/hlink/folHlink with srgbClr/sysClr payloads — verified
- `.planning/phases/01-foundation/01-RESEARCH.md` — Phase 1 patterns carried forward (URI tags, RawXML hoarding, pointer fields, sentinel errors, Warnings())

### Secondary (MEDIUM confidence)
- ISO/IEC 29500-1 §17.7.4.3 (basedOn), §17.7.4.10 (next), §17.7.4.6 (link), §17.7.4.17 (style), §17.7.2 (inheritance order) — cited from training knowledge of the spec; section numbers stable but exact text not re-fetched this session
- ISO/IEC 29500-1 §17.3.1.29 (w:color), §17.18.95 (themeTint), §17.18.94 (themeShade), §17.18.96 (themeColor ST_ThemeColor enum) — cited
- ISO/IEC 29500-1 §17.9.17 (w:num), §17.9.18 (lvlOverride), §17.9.24 (startOverride), §17.9.2 (abstractNumId), §17.9.9 (lvl) — cited
- ECMA-376 Part 1 style inheritance model (docDefaults → latentStyles → basedOn chain → direct formatting) — cited via PRD §6 FR-3.2 and CONTEXT.md domain statement

### Tertiary (LOW confidence)
- themeTint/themeShade exact per-channel formula (A2) — training knowledge of "Word's behavior"; verify with fixture
- `link` consultation order (A6, Open Question 1) — implement defensively or drop for v1
- Exact Heading2 canonical expected values (A7) — fixture-author captures from real Word file

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — stdlib + internal packages verified in codebase this session
- Architecture: HIGH — patterns anchored to existing Phase 1 code + OOXML spec sections
- Cloner strategy: HIGH — pure OPC wiring over verified Phase 1 API
- basedOn merge + cycle detection: HIGH — anchored to D-05/D-07 + spec sections (text CITED not re-fetched)
- Theme color resolution: MEDIUM-HIGH — enum/element mismatch verified against `defaults/theme1.xml`; tint/shade formula CITED (A2 to verify)
- Numbering resolution: MEDIUM-HIGH — `CT_Num` lvlOverride absence is a verified blocker (Pitfall 4); struct extension is the fix
- Fixture design: MEDIUM — schema proposed (discretion); final values user-authored per D-10/D-11
- Cache invalidation: HIGH — dirty-flag approach reuses Phase 1 `Part.modified`

**Research date:** 2026-07-25
**Valid until:** 2026-08-24 (stable domain; OOXML spec sections stable, internal APIs verified this session)
