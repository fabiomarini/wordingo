# v0.2.0 — wordingo

Pure Go, zero-dependency library for creating and editing Microsoft Word (.docx) documents.

## Install

```
go get github.com/fabiomarini/wordingo
```

## What's new in v0.2.0

### Branded templates keep their letterhead

`FromTemplate` / `FromTemplateReader` now clone the source's **headers and footers** —
including their relationship graphs and embedded media (logos) — over the empty body,
instead of dropping them. A template with a logo in its header produces documents that
carry the same letterhead.

- Header/footer parts are cloned with each part's own `.rels` graph; image parts are
  renamed to avoid collisions with the target package, and document-level relationship
  ids are freshly assigned
- Page geometry (`SectPr` header/footer references) is preserved
- `Header.AddImageBytes` / `Footer.AddImageBytes` — author a logo directly inside a
  header/footer via the part's own rels graph (previously images were document-level only)

### Reliability

- `opc.Save` now emits `.rels` parts for relationship sets whose rels part is not in
  `Parts` (previously such relationships were silently dropped from the saved package)

## Breaking changes

- `FromTemplate` semantics: the result now **retains** the template's headers/footers
  (previously an empty package). Callers relying on a letterhead-free shell should strip
  the cloned parts explicitly.

## What's included

Everything from v0.1.1: TOC fields, Word-faithful lists, lifecycle-safe `Open`,
deterministic re-encoding, tables, images, styles, markdown import.
