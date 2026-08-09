# Quick Task: Let's add the necessary API and structure to add the Table of contents / summary to a document leveraging the headings in the document itself

**Date:** 2026-08-09
**Branch:** main

## What Changed

- **New public API** (`toc.go`):
  - `Document.Headings() []Heading` — reads the document outline from body paragraphs styled `Heading1`–`Heading9` or carrying an explicit `w:outlineLvl` (level 1..9, trimmed text, style id), in document order.
  - `Document.AddTableOfContents(opts *TOCOptions) (*TOC, error)` — appends a TOC section at the end of the body: a title paragraph with a real Word TOC field (` TOC \o "1-N" \h \z \u `), one entry paragraph per heading up to the configured depth, and the field closing character.
  - `Document.InsertTableOfContentsBefore(target *Paragraph, opts *TOCOptions) (*TOC, error)` — same section inserted before a target paragraph; returns `ErrTOCTargetNotFound` for nil/foreign targets.
  - `TOCOptions{Title, Levels, UpdateOnOpen}` — custom title (default "Table of Contents"), depth clamp 1..9 (default 3), and update-on-open switch (default true: patches `word/settings.xml` with `w:updateFields w:val="true"` so Word repopulates the TOC with live page numbers on open).
  - `TOC` handle with `Title()` accessor; `Heading` value type.
- **Generated structure**: each static entry is a hyperlink to a `_Toc…` bookmark placed on its heading (via new `w:bookmarkStart`/`w:bookmarkEnd` support) plus a `PAGEREF` field for the page number; entries use inline formatting (no style references) so no unknown-style warnings occur in template documents; right dot-leader tab position derived from the section page width (9360 twips fallback). Headings with no text are skipped; an empty-heading document gets a hint paragraph; repeated calls keep bookmark ids/names unique (existing bookmarks in body and table cells are scanned).
- **New WML types** (`internal/wml/fields.go`): `CT_FldChar`, `CT_InstrText` (xml:space preservation), `CT_BookmarkStart`, `CT_BookmarkEnd`, `CT_UpdateFields`; `CT_P`/`CT_R`/`CT_RPr`/`CT_Settings`/`CT_Hyperlink` extended (fldChar/instrText/bookmarks/noProof/updateFields/history).
- **Fixed encoder bug** (`internal/xmlutil/encoder.go`): `mc:Ignorable` was emitted without an `xmlns:mc` declaration (unbound prefix → invalid XML whenever w14/w15 extension content was re-encoded, e.g. settings.xml); xmlns declarations are now emitted in sorted order for byte-deterministic output.
- **Fixed file lifecycle bug** (`open.go`, `wordingo.go`): `Open(path)` closed the source file before `opc.Save` lazily copied unmodified parts, so saving an opened document failed with "file already closed". The file now stays open until `Document.Close()`, and `Save` rejects writing over the still-open source file.
- **Docs & example**: README feature row + "Table of contents / summary" section, `docs/api-reference.md` (methods + types), `docs/overview.md` row, new `examples/12-table-of-contents/` with reopen verification.

## Files Modified

- `toc.go`, `toc_test.go` (new)
- `internal/wml/fields.go` (new)
- `internal/wml/document.go`, `internal/wml/hyperlink.go`, `internal/wml/styles.go` (extended types)
- `internal/wml/wml_test.go` (field round-trip tests)
- `internal/xmlutil/encoder.go` (mc:Ignorable fix + deterministic xmlns), `internal/xmlutil/xmlutil_test.go` (encoder tests)
- `open.go`, `wordingo.go` (file lifecycle)
- `README.md`, `docs/api-reference.md`, `docs/overview.md`
- `examples/12-table-of-contents/main.go`, `examples/12-table-of-contents/README.md` (new)

## Verification

- `go build ./...`, `go vet ./...` clean; `gofmt` clean on all new/modified code.
- `go test ./...` — all packages pass, including new tests: `TestHeadings*`, `TestAddTableOfContents*`, `TestInsertTableOfContents*`, `TestTOC*` (structure, options, empty doc, bookmark uniqueness/continuation, round-trip, opened-file flow, table element order, no-style-warnings), `TestRoundTrip_Fields`, `TestRoundTrip_UpdateFields`, `TestEncoder_MCIgnorableDeclaresMC`, `TestEncoder_NoDuplicateMCDeclaration`, `TestEncoder_DeterministicOutput`, `TestOpenKeepsSourceFileForLazyParts`. `go test -race` passes.
- End-to-end: generated docx validated with Python expat (all 9 XML parts well-formed), OPC relationship targets all resolve, content types fully covered; TOC field instruction, 4 PAGEREF fields, 4 heading bookmarks, 4 hyperlink anchors, and settings `w:updateFields` verified in the saved package.
- Example `examples/12-table-of-contents` runs end-to-end (build doc → outline → TOC at start + summary at end → save → reopen → checks pass, no warnings).
- Failure modes: nil/foreign insert target → `ErrTOCTargetNotFound`; empty-heading document → hint paragraph + closed field; save over the open source file → rejected; `UpdateOnOpen=false` → settings.xml untouched.
