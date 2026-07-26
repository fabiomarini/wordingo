# Phase 6: Merge & Edit (v0.1.0) - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-07-26
**Phase:** 6-merge-edit-v1-0
**Areas discussed:** Merge API surface, Split-run placeholder handling, Paragraph targeting for edits, Mixed body order (P + Tbl), Table row deletion, ReplaceWithinRun vs SetText, v1.0 acceptance criteria

---

## Merge API Surface

| Option | Description | Selected |
|--------|-------------|----------|
| doc.Merge(data) — all parts | Single map, scans body+tables+headers+footers automatically | |
| doc.Merge(data, opts) — options | MergeOpts struct for targeting (ScopedParts) | ✓ |

**User's choice:** doc.Merge(data, opts) — options struct
**Notes:** User chose options struct with ScopedParts for targeting

| Option | Description | Selected |
|--------|-------------|----------|
| ScopedParts — what to scan | Booleans for Body/Headers/Footers/Tables | ✓ |
| PrefixSuffix — custom delimiters | MergeOpts with non-standard delimiters | |
| FuncMap — template functions | {{upper:name}} transformations | |

**User's choice:** ScopedParts — what to scan

| Option | Description | Selected |
|--------|-------------|----------|
| nil = scan all, warn unused | Scan all parts. Keys not found produce warnings | ✓ |
| nil = scan all, ignore unused | No warnings for unused keys | |

**User's choice:** nil = scan all, warn unused

---

## Split-Run Placeholder Handling

| Option | Description | Selected |
|--------|-------------|----------|
| First-fragment formatting (MERGE-02) | Merged text gets first run's formatting | |
| Merge all fragment formatting | Keep fragments, distribute formatting | ✓ (first choice) |
| First-fragment unless only one has format | Smarter heuristic | |

**User's choice:** Initially "Merge all fragment formatting", then clarified to "Keep it simple — first-fragment formatting"
**Notes:** User initially chose "Merge all" then clarified to "first-fragment formatting" when conflict was pointed out. Final: first-fragment formatting, all text in first run, other fragments deleted.

| Option | Description | Selected |
|--------|-------------|----------|
| Keep fragment structure, replace per-fragment | Proportional text distribution | |
| All text in first fragment, remove others | First run gets merged value | ✓ |
| All text in all fragments | Same content in every run | |

**User's choice:** All text in first fragment, remove others (then clarified to "first-fragment formatting")

| Option | Description | Selected |
|--------|-------------|----------|
| Concatenate adjacent runs, find {{, map back | Join texts, find placeholders, map to runs | ✓ |
| Scan runs sequentially for {{, }} | Simpler sequential scan | |

**User's choice:** Concatenate adjacent runs, find {{, map back

---

## Paragraph Targeting for Edits

| Option | Description | Selected |
|--------|-------------|----------|
| By Paragraph pointer | InsertBefore/InsertAfter/DeleteParagraph by reference | ✓ |
| By index | InsertParagraphAt/DeleteParagraphAt by index | |

**User's choice:** By Paragraph pointer

| Option | Description | Selected |
|--------|-------------|----------|
| InsertBefore + InsertAfter as Document methods | doc.InsertBefore(target, text) | ✓ |
| InsertAfter on Paragraph itself | para.InsertAfter(text) | |

**User's choice:** InsertBefore + InsertAfter as Document methods

| Option | Description | Selected |
|--------|-------------|----------|
| Same API — Header/Footer implement ParagraphContainer | Interface for body+header/footer editing | ✓ |
| Scope to body paragraphs only | Body-only, headers append-only | |

**User's choice:** Same API — Header/Footer implement ParagraphContainer

---

## Mixed Body Order (P + Tbl)

| Option | Description | Selected |
|--------|-------------|----------|
| Keep v1 limitation — tables after paragraphs | P slice first, Tbl second | |
| Unified body content with interleaved ordering | BodyElement interface, single slice | ✓ |

**User's choice:** Unified body content with interleaved ordering

| Option | Description | Selected |
|--------|-------------|----------|
| Public BodyElement interface | Paragraph and Table implement it | ✓ |
| Internal only — public API hides change | Internally interleaved, public stays same | |

**User's choice:** Public BodyElement interface

| Option | Description | Selected |
|--------|-------------|----------|
| Keep Paragraphs()+Tables() + add Body() | Convenience + ordered Body() | ✓ |
| Replace with Body() only | Single source of truth | |

**User's choice:** Keep Paragraphs()+Tables() + add Body()

---

## Table Row Deletion

| Option | Description | Selected |
|--------|-------------|----------|
| By row index on Table | table.DeleteRow(idx int) | ✓ |
| By RowBuilder pointer | row.Delete() | |

**User's choice:** By row index on Table

| Option | Description | Selected |
|--------|-------------|----------|
| On TableBuilder | tableBuilder.DeleteRow(idx int) | ✓ |
| On Document with table pointer | doc.DeleteTableRow(table, idx) | |

**User's choice:** On TableBuilder

---

## ReplaceWithinRun vs SetText

| Option | Description | Selected |
|--------|-------------|----------|
| Run.SetText(s string) | Replace entire run content | |
| Run.ReplaceText(old, new string) | Find and replace within run | |
| Both — SetText + ReplaceText | Full + partial replacement | ✓ |

**User's choice:** Both — SetText for full, ReplaceText for partial

---

## v1.0 Acceptance Criteria

| Option | Description | Selected |
|--------|-------------|----------|
| Golden file tests | Automated byte/visual diff | |
| Opens in Word without repair | Manual check | |
| Functional correctness | Automated content verification | ✓ |

**User's choice:** Opens in Word without repair

| Option | Description | Selected |
|--------|-------------|----------|
| All requirements + tests pass | MERGE-01..03, EDIT-01..03, all green | ✓ |
| Plus v1.0.0 git tag + release notes | Tag + release notes | |
| Just merge+edit working + opens repair | Lean | |

**User's choice:** All requirements + tests pass

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — close all QUAL reqs for v1.0 | Audit error patterns, public API, X() | ✓ |
| Only if they cause issues | Don't actively audit | |

**User's choice:** Yes — close all QUAL reqs for v1.0

**Free-text:** Release as v0.1.0, not v1.0

---

## the agent's Discretion

- Exact MergeOpts struct field naming and zero-value defaults
- BodyElement interface method signature and type-switch ergonomics
- ParagraphContainer interface shape
- DeleteParagraph behavior when paragraph belongs to header/footer
- ReplaceText algorithm (strings.Replace vs strings.ReplaceAll)
- Internal file layout: merge.go, edit.go, body.go
- Merge warning message format
- Row index bounds checking style (panic vs error)
- Numbering.xml merge (merge creates no new numbering defs — placeholder replacement only)

## Deferred Ideas

None — discussion stayed within phase scope.
