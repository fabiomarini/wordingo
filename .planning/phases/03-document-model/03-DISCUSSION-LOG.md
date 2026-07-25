# Phase 3: Document Model - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-07-25
**Phase:** 03-document-model
**Areas discussed:** API layout, Body model, Lazy loading, FromTemplate policy, Round-trip harness, Style protection, API surface, Body iteration, Test fixtures

---

## API Layout

| Option | Description | Selected |
|--------|-------------|----------|
| Public package now | Create wordingo/ with Document, Open/Create/FromTemplate/Save/Paragraphs. Phase 4 adds mutation methods. Matches PRD API sketch. | ✓ |
| Internal until Phase 4 | Keep Document in internal/ or create draft. Phase 4 creates public wordingo/ package with everything at once. | |

**User's choice:** Public package now
**Notes:** None

---

## Body Model

| Option | Description | Selected |
|--------|-------------|----------|
| Paragraph wrappers | Document.Paragraphs() []*Paragraph — each Paragraph wraps *wml.CT_P with read-only accessors (Style(), Text()). Phase 4 adds mutation methods. Matches PRD sketch. | ✓ |
| Raw CT_Body via X() | Document.Body() *wml.CT_Body via escape hatch. User navigates p/r/t directly on raw WML. Simpler but bypasses the wrapper model. | |

**User's choice:** Paragraph wrappers
**Notes:** None

---

## Lazy Loading

| Option | Description | Selected |
|--------|-------------|----------|
| Eager on Open | Parse body immediately during Open(). Simpler, body always available. <50ms for 100-page doc is acceptable. | ✓ |
| Lazy on first Paragraphs() | Keep raw bytes until first Paragraphs() call. Maximizes lazy loading — unparsed if user only calls Save/Warnings. | |

**User's choice:** Eager on Open
**Notes:** Supporting parts (headers, footers, etc.) stay lazy — parsed on first access

---

## FromTemplate Body Policy

| Option | Description | Selected |
|--------|-------------|----------|
| Two separate functions | FromTemplate(path) → clears body. OpenTemplate(path) → keeps existing body. Clear intent per function. | ✓ |
| Single function with flag | FromTemplate(path, keepBody bool) → single entry point, bool controls body clearing. Fewer functions. | |

**User's choice:** Two separate functions
**Notes:** FromTemplate + OpenTemplate. Reader variants too.

---

## Round-trip Harness

| Option | Description | Selected |
|--------|-------------|----------|
| Per-part byte diff | Open fixture → save to buffer → unzip both → diff each part byte-identical. Same approach as Phase 1 D-04. Works today. | ✓ |
| Part-level modified flag check | Assert opc.Part.IsModified() returns false for unmodified parts. Lighter but less thorough. | |
| Both combined | Modified flag check in unit tests + per-part byte diff in integration tests for full confidence. | |

**User's choice:** Per-part byte diff
**Notes:** Same as Phase 1 D-04

---

## Style Protection

| Option | Description | Selected |
|--------|-------------|----------|
| Automatic via OPC layer | Document never calls MarkModified on style parts unless user explicitly modifies styles. OPC Save raw-copies untouched parts. | ✓ |
| Explicit lock on style parts | Add a Document-level guard: read-only access to style parts during open/save. Any mutation to style parts requires explicit API call. Extra safety layer. | |

**User's choice:** Automatic via OPC layer
**Notes:** STYLE-ROUNDTRIP is automatic — no extra mechanism needed

---

## API Surface

| Option | Description | Selected |
|--------|-------------|----------|
| Full PRD set | Open(path), OpenReader(r,size), Create(), FromTemplate(path), FromTemplateReader(r,size), Save(path), WriteTo(w), Paragraphs(), Warnings(), Close(). | ✓ |
| Minimal: file paths only | Open(path), Create(), FromTemplate(path), Save(path). Add reader/writer variants in Phase 4. | |

**User's choice:** Full PRD set
**Notes:** Reader/writer variants from the start

---

## Body Iteration

| Option | Description | Selected |
|--------|-------------|----------|
| Paragraphs-only | Document.Paragraphs() []*Paragraph — returns only paragraphs. Tables accessed separately via Document.Tables() []*Table. Clear separation. | ✓ |
| Mixed content blocks | Document.Content() []Block — Block interface covers Paragraph and Table. More flexible but diverges from PRD Paragraphs() sketch. | |

**User's choice:** Paragraphs-only
**Notes:** Tables added in Phase 5 via Document.Tables()

---

## Test Fixtures

| Option | Description | Selected |
|--------|-------------|----------|
| Use existing fixtures | Phase 1 and style-engine testdata/. Open→save→per-part diff on existing fixtures. Zero unintended diffs. | |
| Add minimal fixture set | 3-4 minimal docx files: blank, single-paragraph, multi-heading, header+footer. Covers body, headers, footers, multiple parts. | ✓ |
| Add rich corpus | Full set: blank, styled, hosted (customXml), pre-populated-template-style with headers/footers. Tests every part type. | |

**User's choice:** Add minimal fixture set
**Notes:** 3-4 docx files under testdata/roundtrip/

---

## the agent's Discretion

- Internal file/function layout within `wordingo/` package
- Exact read-only accessor shape on `Paragraph` (Style() string, Text() string, X() *wml.CT_P)
- Fixture filenames and location within `testdata/roundtrip/`
- Whether to store raw body bytes for eager parse + diff-check or parse to CT_Document only
- `OpenTemplate` vs `OpenTemplateReader` naming (confirm with user during plan)

## Deferred Ideas

- Document.Tables() — Phase 5
- Paragraph mutation (AddRun, SetStyle, formatting) — Phase 4
- Edit operations — Phase 6
- Template merge — Phase 6
