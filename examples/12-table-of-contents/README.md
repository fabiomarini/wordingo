# Example 12: Table of Contents

## What it demonstrates

The table-of-contents API: `Headings()` (read the document outline from `Heading1`–`Heading9` styles and explicit outline levels), `AddTableOfContents` (append a TOC section built from the headings), `InsertTableOfContentsBefore` (place a TOC before a specific paragraph), `TOCOptions` (custom title, depth, and update-on-open control), plus reopen verification.

## How to run

```
cd examples/12-table-of-contents
go run main.go
```

## Output

- `output.docx` — a report with a "Table of Contents" section right after the subtitle and a second "Summary" section at the end (2-level depth).

## How it works

- `doc.Headings()` walks the body paragraphs and returns `{Level, Text, Style}` for every paragraph styled `Heading1`..`Heading9` (or carrying an explicit `w:outlineLvl`), in document order. Empty headings are skipped.
- `doc.InsertTableOfContentsBefore(target, nil)` inserts a TOC section before a given paragraph. The section is a real Word field: `<w:fldChar begin>` + `TOC \o "1-3" \h \z \u` + `<w:fldChar separate>` + the cached entries + `<w:fldChar end>`.
- Each cached entry is a hyperlink to a `_Toc…` bookmark placed on its heading, followed by a `PAGEREF` field for the page number — so the summary is navigable even in viewers that don't refresh fields.
- `w:updateFields w:val="true"` is written into `word/settings.xml`, so Word repopulates both TOCs — with live page numbers — the first time the file is opened. Pass `&TOCOptions{UpdateOnOpen: &false}` to skip that.
- `doc.AddTableOfContents(&TOCOptions{Title: "Summary", Levels: 2})` shows a custom title and a shallower depth (Heading1–2 only).

## Notes

- TOC entries themselves are plain paragraphs with inline formatting (no style references), so no "unknown style" warnings are produced even in template-based documents that lack the built-in `TOC1`–`TOC9` styles.
- Headings inside table cells, headers, and footers are not collected — the outline is body-paragraph based.
