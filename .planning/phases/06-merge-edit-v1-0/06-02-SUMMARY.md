---
phase: 06-merge-edit-v1-0
plan: 02
subsystem: edit-ops
tags: [edit, insert, delete, body-element, paragraph-container, release]
requires:
  - phase: 06-01
    provides: merge engine, Merge() API
provides:
  - InsertBefore(target, text) — by pointer identity
  - InsertAfter(target, text) — by pointer identity
  - DeleteParagraph(target) — by pointer identity
  - DeleteRow(idx) — table row deletion with OOB error
  - SetText(s) — replace run text content
  - ReplaceText(old, new) — substring replacement in run
  - Body() — returns []BodyElement (interleaved paragraphs+tables)
  - ParagraphContainer interface for header/footer paragraph editing
  - UC1–UC4 acceptance tests
affects: []
tech-stack:
  added: []
  patterns:
    - Pointer-identity matching (p.ct == target.ct) for editing operations
    - BodyElement accessor for document-order iteration
    - ParagraphContainer dispatch (body vs header/footer)
    - DeleteRow bounds check returns error, not panic
key-files:
  created:
    - body.go — BodyElement, BodyElementType, Body(), BodyParagraph, BodyTable
    - body_test.go — Body accessor tested
    - edit_test.go — 11 tests (insert, delete, set/replace text, style round-trip)
    - uc_test.go — 4 acceptance tests (UC1–UC4)
  modified:
    - wordingo.go — InsertBefore, InsertAfter, DeleteParagraph, header/footer paragraph edit methods
    - body.go
    - table.go — DeleteRow
    - run.go — SetText, ReplaceText
    - paragraph.go — SetPageBreakBefore
requirements-completed:
  - EDIT-01
  - EDIT-02
  - EDIT-03
  - QUAL-01
  - QUAL-02
  - QUAL-03
---

## Summary

Built edit operations, BodyElement ordered accessor, and acceptance tests.

### Changes

- **body.go** (NEW, 62 lines): `BodyElementType` enum (ElementParagraph, ElementTable), `BodyElement` struct with Type/Para/Table, `Body()` method returning interleaved paragraphs+tables in document order. `BodyParagraph`/`BodyTable` wrappers with X() escape hatch.

- **edit_test.go** (NEW, 267 lines): 11 tests covering InsertBefore, InsertAfter, DeleteParagraph, DeleteParagraphNotFound, DeleteRow, DeleteRowOutOfBounds, SetText, ReplaceText, ReplaceTextNoMatch, and StyleRoundTrip (style parts byte-identical after edits).

- **uc_test.go** (NEW, 193 lines): 4 acceptance tests (UC1–UC4) from ROADMAP:
  - UC1: Template merge with reopen verification
  - UC2: Missing key warnings
  - UC3: Edit operations (InsertBefore, DeleteParagraph, merge)
  - UC4: Table edit with DeleteRow + merge

### Deviations

None. All operations by pointer identity per D-08. No index-based targeting.
