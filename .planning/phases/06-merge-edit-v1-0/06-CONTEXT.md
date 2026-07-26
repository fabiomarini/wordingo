# Phase 6: Merge & Edit (v0.1.0) - Context

**Gathered:** 2026-07-26
**Status:** Ready for planning

<domain>
## Phase Boundary

Template merge (replace `{{key}}` placeholders in paragraphs, table cells, headers, footers with data) + edit operations (insert/delete paragraphs, delete table rows, replace run text) + v0.1.0 release (UC1–UC4 pass end-to-end, all requirements closed).

Requirements: MERGE-01..03, EDIT-01..03, QUAL-01..03. Release tagged as v0.1.0 (not v1.0).

</domain>

<decisions>
## Implementation Decisions

### Merge API (MERGE-01)
- **D-01:** `doc.Merge(data map[string]string, opts *MergeOpts)` — options struct, not single-map method. Matches PRD FR-7.1 sketch with extensibility.
- **D-02:** `MergeOpts` has `ScopedParts` struct with booleans: `Body`, `Headers`, `Footers`, `Tables`. All default true. Single struct covering which parts to scan for placeholders.
- **D-03:** `nil` opts = scan all parts, warn on unused keys in data. Unused keys produce warnings via `Warnings()`, never silent corruption (MERGE-03).

### Split-Run Placeholder Handling (MERGE-02)
- **D-04:** Detect split-run placeholders by concatenating adjacent run texts, finding `{{key}}`, mapping character ranges back to original runs. Handles arbitrary split patterns.
- **D-05:** Merged value goes in the first fragment run with its formatting. Other fragment runs are deleted. Matches "first-fragment formatting" approach. Not proportional distribution.

### Paragraph Editing (EDIT-01)
- **D-06:** Paragraphs identified by pointer: `doc.InsertBefore(target *Paragraph, text string) *Paragraph`, `doc.InsertAfter(target *Paragraph, text string) *Paragraph`, `doc.DeleteParagraph(target *Paragraph)`.
- **D-07:** Header/Footer paragraphs support same editing API via a `ParagraphContainer` interface.
- **D-08:** No index-based paragraph targeting — pointer/reference only.

### Mixed Body Order (CT_Body P + Tbl)
- **D-09:** Public `BodyElement` interface with concrete types `BodyParagraph` and `BodyTable`. Single `[]BodyElement` slice preserves document-order interleaving of paragraphs and tables.
- **D-10:** Keep existing `Paragraphs()` and `Tables()` convenience methods for backward compatibility. Add `Body() []BodyElement` as the canonical ordered accessor.

### Table Row Deletion (EDIT-01)
- **D-11:** `TableBuilder.DeleteRow(idx int)` — by row index on the table builder. Consistent with existing builder pattern.

### Run Text Replacement (EDIT-02)
- **D-12:** Two methods on `Run`: `SetText(s string)` replaces entire run content; `ReplaceText(old, new string)` for targeted replace within run text.

### v0.1.0 Release Acceptance
- **D-13:** Version tagged as `v0.1.0`, not `v1.0`. Phase rename from ROADMAP "v1.0" to v0.1.0.
- **D-14:** Acceptance: MERGE-01..03 and EDIT-01..03 pass all criteria. All existing tests green. Generated documents open in Word without repair dialog.
- **D-15:** QUAL-01..03 formally closed: audit all public methods for error-not-panic, verify io.ReaderAt/io.Writer I/O, confirm single public package with X() on all wrappers.

### the agent's Discretion
- Exact `MergeOpts` struct field naming and zero-value defaults
- `BodyElement` interface method signature and type-switch ergonomics
- `ParagraphContainer` interface shape
- `DeleteParagraph` behavior when paragraph belongs to header/footer — delegates to Header/Footer's own slice management
- `ReplaceText` algorithm (strings.Replace vs strings.ReplaceAll)
- Internal file layout: merge.go, edit.go, body.go
- Merge warning message format
- Checkpoint checkpoint ID generation for roll-forward merge
- Row index bounds checking style (panic vs error)
- Numbering.xml merge (merge creates no new numbering defs — placeholder replacement only)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Product & Requirements
- `.planning/PRD.md` — §6 FR-7 (template merge), FR-8 (editing), §8 API sketch (Merge, InsertAfter, SetText), UC1–UC4 use cases, §5 (user stories)
- `.planning/REQUIREMENTS.md` — MERGE-01..03, EDIT-01..03, QUAL-01..03 normative text
- `.planning/PROJECT.md` — Constraints (stdlib only, MIT, Go 1.23+), Key Decisions table
- `.planning/ROADMAP.md` — Phase 6 goal, success criteria, plan breakdown (06-01 merge engine, 06-02 edit ops + release)

### Prior Phase Context (carried forward — must respect)
- `.planning/phases/05-rich-content/05-CONTEXT.md` — D-24 (tables after all paragraphs v1 limitation — Phase 6 reverses with BodyElement interleaving)
- `.planning/phases/04-content-api/04-CONTEXT.md` — D-04 (InsertRun/RemoveRun deferred to Phase 6 — editorial ops now)
- `.planning/phases/03-document-model/03-CONTEXT.md` — D-02 (Paragraph wrapper), D-03 (lazy loading), D-05 (per-part round-trip diff)
- `.planning/phases/01-foundation/01-CONTEXT.md` — D-04 (per-part byte diff), D-09..D-12 (error/warning patterns)

### Existing Code (read during scout)
- `wordingo.go` — Document struct, AddParagraph, Paragraphs(), Warnings(), serializeBody, dirty flag
- `paragraph.go` — Paragraph wrapper with X(), AddRun, SetAlignment, SetSpacing, SetStyle
- `run.go` — Run wrapper with SetBold/SetItalic/SetColor/SetSize/SetStyle
- `table.go` — TableBuilder, RowBuilder, CellBuilder, AddTable, AddTableBuilder
- `list.go` — ListBuilder, numbering.xml read/write helpers
- `header.go` — Header/Footer wrappers, addHelperPart, AddHeader, AddFooter
- `template.go` — FromTemplate, OpenTemplate (body clone + HdrFtrRef fixup)
- `internal/wml/document.go` — CT_Body (P + Tbl + SectPr), CT_P (R + Hyperlink), CT_R (T + RPr), CT_Text (Value)

### External Specifications (no local copies — cite in plan as needed)
- ISO/IEC 29500 Part 1 — §17.3 (Run Properties), §17.7 (Paragraph Properties), §17.4 (Tables), §17.6 (Headers/Footers)
- ECMA-376 Part 2 (OPC) — part creation, relationship management, content types

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `wordingo.go` — `Document.warnings` accumulation + `warn()` method; `d.dirty` flag pattern
- `internal/wml/document.go` — `CT_Text` with `Value` string (replaced during merge); `CT_P` with `R []*CT_R` runs
- `internal/wml/hyperlink.go` — `CT_Hyperlink` with child runs — merge must skip hyperlink runs or handle separately
- `header.go` — `addHelperPart` for OPC part creation (reusable if merge creates helper parts)
- `list.go` — `readOrCreateNumbering`/`writeNumbering` pattern for part read/write
- `table.go` — `CT_Tbl` with `Tr []*CT_Tr`, `DeleteRow` needs to splice this slice
- `open.go` — doc-level pattern for opening/reading

### Established Patterns
- Wrapper-over-schema with `X()` escape hatch — BodyElement wrappers should follow this
- Builder chaining — methods return receiver for fluent API
- Deferred errors via `Warnings()` — missing merge keys use this pattern
- Per-part byte diff for round-trip assertion (STYLE-ROUNDTRIP)
- `d.dirty` flag + `serializeBody()` for body content changes

### Integration Points
- `CT_Body.P` and `CT_Body.Tbl` currently separate slices — BodyElement unifies these
- `Document.dirty` flag triggers body re-serialize — merge + edit must set this
- `Document.Paragraphs()` returns `[]*Paragraph` — InsertBefore/After needs to find paragraph's index in body
- Header/footer paragraphs are not tracked by Document.Paragraphs() — ParagraphContainer interface exposes them for editing
- Merge needs access to header/footer parts and table cell paragraphs — walk all content parts
- Numbering.xml already managed by list.go helpers — merge doesn't touch it
- ReplaceText on Run needs to modify CT_Text.Value and set dirty flag

</code_context>

<specifics>
## Specific Ideas

- Merge implementation: walk body paragraphs, then tables (each cell paragraph), then header/footer parts. Per D-02, opts.ScopedParts controls which are scanned.
- Split-run detection: join all run texts with a map from char offset → run index. Find `{{key}}` in joined text. Delete all runs in the fragment range except the first; set first run's text to merged value.
- BodyElement could be a simple struct union with a type enum, not a full interface, to avoid heap allocations in a hot path. Agent discretion.
- ParagraphContainer: Header and Footer both have P []*CT_P. Interface adds InsertParagraphAt/DeleteParagraphAt internal methods.
- v0.1.0 release: no breaking changes expected for existing API. Additive only.
- ReplaceText: strings.ReplaceAll(run.ct.T.Value, old, new) — simple. Could add count parameter later.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 06-merge-edit-v0-1-0*
*Context gathered: 2026-07-26*
