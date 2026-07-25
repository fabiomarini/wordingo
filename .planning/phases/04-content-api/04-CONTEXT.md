# Phase 4: Content API - Context

**Gathered:** 2026-07-26
**Status:** Ready for planning

<domain>
## Phase Boundary

Users create fully formatted text documents programmatically — paragraphs, runs, inline formatting (bold, italic, underline, font, size, color, highlight), paragraph layout (alignment, spacing, indentation), and named styles applied via the style engine. Building on Phase 3's read-only Document/Paragraph wrappers.

Requirements: API-01..03, QUAL-01..03. Tables deferred to Phase 5. Template merge and edit operations deferred to Phase 6.

Phase 4 is about WRITING content. Phase 3 handled READING content.

</domain>

<decisions>
## Implementation Decisions

### Run Formatting API (API-01)
- **D-01:** Per-attribute setters + a `RunFormat` struct for bulk setting. Setters: `SetBold(b bool)`, `SetItalic(b bool)`, `SetUnderline(u string)`, `SetFont(name string)`, `SetSize(pts float64)`, `SetColor(hex string)`, `SetHighlight(color string)`. `SetFormatting(RunFormat)` for bulk.
- **D-02:** Setters return `*Run` for builder chaining (`run.SetBold(true).SetSize(12)`). Errors deferred to `Warnings()` at Save time.
- **D-03:** v1 scope matches API-01 exactly: bold, italic, underline, font, size, color, highlight. Strike, subscript, superscript, smallCaps, etc. deferred beyond v1 (or to Phase 6 polish).
- **D-04:** `Paragraph.AddRun(text string) *Run` appends a new CT_R to the paragraph's run array. No `InsertRun(index)` or `RemoveRun(index)` in v1 — those are editing operations for Phase 6.

### Style Application & Resolver (API-03)
- **D-05:** `SetStyle` sets only the style reference (`pStyle`/`rStyle`) in XML. No pre-population of resolved properties. Word resolves at render time. `SetStyle(name string) *Paragraph` and `Run.SetStyle(name string) *Run`.
- **D-06:** `Paragraph.SetStyle(name string)` — method lives on Paragraph. Paragraph holds a back-reference to Document for access to the style resolver and `MarkModified` trigger.
- **D-07:** Resolver stays internal to the Document. No public `StyleResolver()` accessor. Users set style names, library resolves when needed internally.
- **D-08:** Run-level character styles supported: `run.SetStyle("Emphasis")` sets `rStyle` on run properties.

### Paragraph Formatting API (API-02)
- **D-09:** Per-category setters: `SetAlignment(a Alignment)`, `SetSpacing(s *ParSpacing)`, `SetIndent(i *ParIndent)`. Plus `SetFormatting(ParFormat)` for bulk. Consistent with run formatting (D-01).
- **D-10:** v1 scope matches API-02 exactly: alignment (left/center/right/justify), spacing (before/after/line), indentation (left/right/firstLine/hanging). KeepNext, keepLines, pageBreakBefore, widowControl, shading, tabs, outlineLvl deferred.
- **D-11:** Setters return `*Paragraph` for builder chaining.

### Error Handling (QUAL-02)
- **D-12:** Builder methods defer errors. Invalid input values (negative size, unknown style name, invalid color hex) logged as `Warnings()` checked at Save time. Matches existing Phase 1/2 pattern (`opc.Warnings()`).
- **D-13:** Invalid input → warning via `Warnings()`. Nil state (nil Document, nil Paragraph, method call after Close) → panic. Matches Go stdlib convention.
- **D-14:** `RunFormat` and `ParFormat` structs are pure data holders — no validation at creation time. Validation happens when values are applied via setter methods.

### the agent's Discretion
- Exact struct field names for `RunFormat` and `ParFormat`
- Alignment type enum values (`AlignmentLeft`, `AlignmentCenter`, etc.)
- Whether `ParSpacing` and `ParIndent` are standalone structs or inlined in `ParFormat`
- Internal file layout within `wordingo/` for formatting code
- `SaveFile` vs `Save` naming (both currently exist — standardize or keep both)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Product & Requirements
- `.planning/PRD.md` — §6 FR-7 (Style-aware content API), §8 API sketch (AddParagraph, AddRun, SetBold, SetItalic, SetFont, SetSize, SetColor, SetAlignment, SetStyle, X() escape hatch), §9 Architecture (layering, wrapper-over-schema)
- `.planning/REQUIREMENTS.md` — API-01..03, QUAL-01..03 normative text
- `.planning/PROJECT.md` — Constraints (stdlib only, MIT, Go 1.23+, <5MB public API), Key Decisions
- `.planning/ROADMAP.md` — Phase 4 goal, success criteria, plan breakdown (04-01 builders + formatting, 04-02 named styles + polish)

### Prior Phase Context (carried forward — must respect)
- `.planning/phases/03-document-model/03-CONTEXT.md` — D-01 (mutation methods added in Phase 4), D-02 (Paragraph wrapper with X()), D-03 (lazy loading), D-06 (style parts never MarkModified unless explicit)
- `.planning/phases/02-style-engine/02-CONTEXT.md` — D-06 (theme colors concretized at resolve-time), D-08/D-09 (clone pass-through)

### External Specifications (no local copies — cite in plan as needed)
- ISO/IEC 29500 Part 1 — §17.3 (Run Properties), §17.7 (Paragraph Properties), §17.9 (Style Inheritance)
- ECMA-376 Part 2 (OPC) — package model, relationship semantics

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `wordingo/paragraph.go` — `Paragraph` wrapper with read-only `Style()`, `Text()`, `X()` accessors. Phase 4 adds `SetStyle()`, `SetAlignment()`, `SetSpacing()`, `SetIndent()`, `AddRun()`, `SetFormatting()`.
- `wordingo/wordingo.go` — `Document` struct wrapping `*opc.Package` + `*wml.CT_Document`. Phase 4 adds `AddParagraph(text)`, resolver back-reference, MarkModified trigger.
- `internal/wml/document.go:59-76` — `CT_PPr` with all paragraph property fields ready (Jc, Spacing, Ind, PStyle, etc.)
- `internal/wml/document.go:90-111` — `CT_RPr` with all run property fields ready (B, I, U, Sz, Color, Highlight, RFonts, RStyle, etc.)
- `internal/wml/properties.go` — Formatting value types: `CT_Jc`, `CT_Spacing`, `CT_Ind`, `CT_RFonts`, `CT_Sz`, `CT_Color`, `CT_Highlight`, `CT_OnOff`, `CT_U`
- `internal/style/resolver.go` — `Resolver` with `ResolveParagraph(styleId)`, `ResolveRun(styleId)` for internal named style resolution
- `internal/opc/package.go` — `MarkModified(partName, data)` for triggering part dirty flags on content changes

### Established Patterns
- Wrapper-over-schema with `X()` escape hatch — every public type wraps a WML CT_* and exposes it via `X()`
- Builder chaining — methods return the receiver for fluent API (PRD §8)
- Deferred errors via `Warnings()` — non-fatal issues accumulated, checked at Save
- Per-part byte diff for round-trip assertion
- No panics for user errors — nil state panics match Go stdlib (e.g., nil map write)

### Integration Points
- `wordingo/` root package orchestrates `internal/opc`, `internal/wml`, `internal/style`
- Paragraph needs Document back-reference for resolver access + MarkModified
- Run wraps `CT_R`, created by Paragraph.AddRun — needs trigger for part modification
- Phase 5 adds AddTable — Phase 4 body mutation API must be extensible
- Phase 6 adds insert/delete — Phase 4 AddParagraph creates paragraphs, Phase 6 InsertParagraph/DeleteParagraph adds editing

</code_context>

<specifics>
## Specific Ideas

- PRD §8 API sketch shows the target shape: `doc.AddParagraph("Quarterly Report").Style("Title")`, `run.SetBold(true)`, `p.SetAlignment(AlignmentCenter)`. Follow this closely.
- Symmetric API across Run and Paragraph: both get per-attribute setters + bulk struct, both return self for chaining.
- Alignment enum: left, center, right, justify (maps to CT_Jc.Val: "left", "center", "right", "both")
- RunFormat struct optional — prefer individual setters, struct is convenience for bulk config

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 04-content-api*
*Context gathered: 2026-07-26*
