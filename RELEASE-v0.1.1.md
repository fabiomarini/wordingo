# v0.1.1 — wordingo

Pure Go, zero-dependency library for creating and editing Microsoft Word (.docx) documents.

## Install

```
go get github.com/fabiomarini/wordingo
```

## What's new in v0.1.1

### Table of contents
- `Document.Headings()` — read the document outline (Heading1–9 styles and explicit outline levels)
- `Document.AddTableOfContents(opts)` / `InsertTableOfContentsBefore(target, opts)` — generate a real Word TOC field (`TOC \o "1-N" \h \z \u`) with cached entries derived from the document's own headings
- Each entry links to a `_Toc…` bookmark placed on its heading, with a `PAGEREF` page-number field
- `TOCOptions{Title, Levels, UpdateOnOpen}` — custom title, depth clamp (1–9), and `w:updateFields` so Word repopulates the TOC on open
- Word-safe field layout: PAGEREF inside the hyperlink, TOC field end outside it (no repair prompts)

### Word-faithful lists
- `numbering.xml` now matches Word's own structure: `abstractNum` definitions before `num` instances, schema-ordered level children (`start`, `numFmt`, `lvlText`, `lvlJc`, `pPr` with hanging indents), and `hybridMultilevel` + `nsid`/`tmpl` metadata
- Bulleted and numbered lists render their markers reliably in Word and other consumers

### Reliability fixes
- **File lifecycle** — `Open(path)` keeps the source file alive for lazy part copies until `Close()`; saving over the still-open source file is rejected with a clear error
- **Deterministic output** — re-encoded XML parts are byte-reproducible (sorted `xmlns` declarations)
- **Round-trip fidelity** — paragraph-mark run properties stay hoarded verbatim; `mc:Ignorable` is emitted with its `xmlns:mc` declaration so extension content (w14/w15) survives re-encoding

## What's included

### Core
- Create blank documents with default styles, Letter, 1-inch margins
- Open existing .docx files for round-trip editing
- Templates: `FromTemplate` (fresh body, cloned styles) and `OpenTemplate` (preserved body)
- Save to file or any `io.Writer`

### Content
- Paragraphs, runs, and named styles (`Title`, `Heading1`–`Heading9`, `Normal`, `Subtitle`, `Quote`, etc.)
- Rich inline formatting: bold, italic, underline, font, size, color, highlight, character styles
- Alignment, spacing, indentation, page-break-before
- Tables: quick grid (`[][]string`) and fluent builder with borders, shading, cell width, merge-right/down, delete row
- Images: PNG + JPEG, auto DPI detection, explicit sizing in inches, locked aspect ratio
- Headers and footers: default, first-page, even-page variants with full paragraph editing
- Lists: bulleted, ordered, 9-level nesting, custom numbering formats
- Hyperlinks with chained run formatting
- **Table of contents built from the document's headings**

### Page setup
- Orientation: portrait / landscape
- Paper sizes: Letter, A4, Legal constants + custom twips
- Margins (twips), page breaks (add or set before)

### Document editing
- Template merge: `{{placeholder}}` replacement in body, tables, headers, footers (split-run safe)
- InsertBefore / InsertAfter / DeleteParagraph by pointer identity
- Run.SetText / ReplaceText, TableBuilder.DeleteRow (bounds-safe)
- Body() ordered iteration mixing paragraphs and tables

### Text & Markdown
- `ExtractText` — plain text from body/tables/headers/footers with scope control
- `ToMarkdown` — GitHub-flavored Markdown export (headings, bold, italic, tables, links, image data URIs, lists, code blocks)
- `CreateFromMarkdown` — build a document from Markdown
- `ImportMarkdown` — append Markdown to an existing document

### Validation
- `Warnings()` — non-fatal issues (unknown style refs, invalid colors, unused merge keys, markdown parse errors)

### Documentation
- [docs/](docs/) — comprehensive documentation (8 guides + full API reference)
- 12 standalone examples under [examples/](examples/) covering every feature
- [README.md](README.md) with colloguial walkthrough

## Design

- **Zero dependencies.** Only Go standard library. `go.mod` has no `require` directives.
- **Faithful round-trips.** Unknown XML elements preserved verbatim. Unchanged parts stay identical byte-for-byte.
- **Template cloning byte-for-byte.** Styles, fonts, colors, numbering, theme copied without re-encoding.
- **Escape hatch.** `X()` methods on all major types expose raw XML structs for anything the public API doesn't cover.
