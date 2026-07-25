# Phase 3: Document Model - Context

**Gathered:** 2026-07-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Library opens, reads, and saves existing documents with zero unintended diffs; creates documents from templates (blank or pre-populated). Delivers the public `wordingo/` API package with `Document` wrapper that boundaries Phase 1's OPC layer and Phase 2's style engine. Requirements: STYLE-ROUNDTRIP-01..02, CREATE-03, CREATE-04.

Round-trip fidelity (STYLE-ROUNDTRIP) is guaranteed by the OPC layer's automatic raw-copy of unmodified parts — Document never calls MarkModified on style parts. FromTemplate clones the style dependency graph via Phase 2's CloneStyles.

Phase 4 adds content mutation (AddParagraph, AddRun, formatting setters). Phase 5 adds tables. This phase is read-only body access + save.

</domain>

<decisions>
## Implementation Decisions

### Public API Package
- **D-01:** Create public `wordingo/` package in this phase with read-only `Document` wrapper. Phase 4 adds mutation methods (AddParagraph, AddRun, formatting). Matches PRD §8 API sketch.

### Body Read Model
- **D-02:** `Document.Paragraphs() []*Paragraph` returns typed paragraph wrappers over `*wml.CT_P`. Each `Paragraph` exposes read-only accessors: `Style()`, `Text()`, `X()` escape hatch. No Table support in Phase 3 — tables added in Phase 5 via `Document.Tables()`.

### Lazy Loading
- **D-03:** `word/document.xml` parsed eagerly into `CT_Document`/`CT_Body` on `Open()`. Supporting parts (headers, footers, numbering, settings, styles, theme, fontTable) stay lazy — parsed on first access via the OPC layer's existing lazy `Part.Open()`. Matches PRD NFR (<50ms for 100-page doc).

### FromTemplate Body Policy
- **D-04:** Two separate functions:
  - `FromTemplate(path)` — clones template style graph, clears body (CREATE-03)
  - `OpenTemplate(path)` — clones template style graph, keeps existing body (CREATE-04)
  Both available as reader variants: `FromTemplateReader(r, size)`, `OpenTemplateReader(r, size)`.

### Round-trip Verification
- **D-05:** Per-part byte diff (same as Phase 1 D-04). Open fixture → save to buffer → unzip both → diff each part byte-identical. Not whole-file ZIP diff (ZIP metadata varies).

### Style Parts Protection (STYLE-ROUNDTRIP)
- **D-06:** Automatic via OPC Save raw-copy behavior. Document never calls `opc.MarkModified` on style parts (styles, numbering, fontTable, theme, settings) during read-only or FromTemplate paths. Only explicit user style mutation (Phase 4+) triggers MarkModified. No extra guard layer needed.

### Full API Surface
- **D-07:** Full PRD §8 function set for Phase 3:
  - `Open(path string) (*Document, error)`
  - `OpenReader(r io.ReaderAt, size int64) (*Document, error)`
  - `Create() (*Document, error)`
  - `FromTemplate(path string) (*Document, error)`
  - `FromTemplateReader(r io.ReaderAt, size int64) (*Document, error)`
  - `OpenTemplate(path string) (*Document, error)`
  - `OpenTemplateReader(r io.ReaderAt, size int64) (*Document, error)`
  - `(*Document).Paragraphs() []*Paragraph`
  - `(*Document).Save(path string) error`
  - `(*Document).WriteTo(w io.Writer) (int64, error)`
  - `(*Document).Warnings() []string`
  - `(*Document).Close() error`

### Test Fixtures
- **D-08:** Minimal round-trip fixture set: 3-4 docx files under `testdata/roundtrip/` — blank, single-paragraph, multi-heading, header+footer. Used for per-part byte diff CI tests.

### the agent's Discretion
- Internal file/function layout within `wordingo/` package
- Exact read-only accessor shape on `Paragraph` (Style() string, Text() string, X() *wml.CT_P)
- Fixture filenames and location within `testdata/roundtrip/`
- Whether to store raw body bytes for eager parse + diff-check or parse to CT_Document only
- `OpenTemplate` vs `OpenTemplateReader` naming (confirm with user during plan)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Product & Requirements
- `.planning/PRD.md` — §5 UC3, UC4 (edit existing doc, pre-populated template), §6 FR-3.3 (STYLE-ROUNDTRIP), FR-5 (Template Creation), FR-9 (I/O flexibility), §8 API sketch (full Document/Paragraph shape), §9 Architecture (wordingo/ public root)
- `.planning/REQUIREMENTS.md` — STYLE-ROUNDTRIP-01..02, CREATE-03, CREATE-04 normative text
- `.planning/PROJECT.md` — Constraints (stdlib only, MIT, Go 1.23+, <5MB), Key Decisions (three separate style operations; style thesis before rich content)
- `.planning/ROADMAP.md` — Phase 3 goal, success criteria, plan breakdown (03-01 open/read/save, 03-02 FromTemplate)

### Prior Phase Context (carried forward — must respect)
- `.planning/phases/01-foundation/01-CONTEXT.md` — D-04 (per-part byte diff round-trip), D-09..D-12 (error/warning patterns, module identity)
- `.planning/phases/02-style-engine/02-CONTEXT.md` — D-08 (clone fresh-empty-target), D-09 (clone byte pass-through), D-06 (theme colors resolve-time, raw parts untouched for Phase 3 byte-identity)

### External Specifications (no local copies — cite in plan as needed)
- ISO/IEC 29500 Part 1 (WordprocessingML) — §17.4 (Document Body), §17.6 (Sections)
- ECMA-376 Part 2 (OPC) — package model, relationship semantics

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/opc/package.go` — `Open()`/`Save()`, per-part raw-copy pass-through (OPC-04), `Part.IsModified()`, `MarkModified()` — all ready for Document wrapper
- `internal/wml/document.go` — `CT_Document`, `CT_Body`, `CT_P`, `CT_R`, `CT_Text` — body types ready for parsing on Open
- `internal/style/cloner.go` — `CloneStyles()` byte pass-through for FromTemplate
- `internal/style/resolver.go` — `Resolver` for style resolution (used by Phase 4 content API, not directly by Phase 3)
- `internal/opc/relationships.go` — `NextRID()`, relationship graph validation (OPC-06)
- `internal/opc/contenttypes.go` — content type registry for override management
- `testdata/` — existing Word/LibreOffice/Google Docs fixtures from Phase 1; `testdata/style-engine/` fixtures from Phase 2

### Established Patterns
- Wrapper-over-schema with `X()` escape hatch — Document wraps `*opc.Package`, Paragraph wraps `*wml.CT_P`
- Lazy part loading — supporting parts parsed on first access, raw bytes otherwise
- Unknown-element hoarding via RawXML — round-trip safety for unmodeled XML
- Per-part byte diff for round-trip assertion (Phase 1 D-04)
- Clone is byte pass-through, fresh-empty-target only (Phase 2 D-08, D-09)

### Integration Points
- `wordingo/` root package is new — orchestrates `internal/opc`, `internal/wml`, `internal/style`
- Phase 4 adds mutation methods on Document/Paragraph — Phase 3's Document must expose enough for that to plug in
- Phase 5 adds tables — `Document.Paragraphs()` is paragraphs-only, `Document.Tables()` added later

</code_context>

<specifics>
## Specific Ideas

- FromTemplate/OpenTemplate clone styles via Phase 2's `CloneStyles` — existing, verified in Phase 2
- Body clearing for FromTemplate: create new empty body (one section, empty) via existing Create() defaults
- Round-trip test: open existing fixture → `doc.Paragraphs()` → no-op → save → diff each part
- Conformance handling: Open→Save preserves source conformance (already in OPC layer). Create/FromTemplate/OpenTemplate always write Transitional

</specifics>

<deferred>
## Deferred Ideas

- Document.Tables() — Phase 5, not Phase 3
- Paragraph mutation (AddRun, SetStyle, formatting) — Phase 4 content API
- Edit operations (insert/delete, replace text) — Phase 6
- Tables in body iteration — Phase 5
- Template merge ({{placeholder}}) — Phase 6

</deferred>

---

*Phase: 03-document-model*
*Context gathered: 2026-07-25*
