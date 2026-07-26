# Example 10: Template to Document

## What it demonstrates

Both template entry points and a combined pipeline: (1) `FromTemplate` cloning a template's style dependency graph into a fresh empty body, then adding styled paragraphs and a table; (2) `OpenTemplate` preserving the template's existing body content (paragraphs and tables) while exposing it for additions; (3) `FromTemplate` combined with `Merge` and `InsertBefore` to build an invoice.

> Requires the `testdata/word/*.docx` files at the repo root; uses `runtime.Caller` to locate them, so run from this directory.

## How to run

```
cd examples/10-template-to-document
go run main.go
```

Running from a different directory will fail to locate `testdata/word/03-custom_styles.docx` and `04-custom_styles_plus_sample_text.docx`.

## Output

- `from-template-report.docx` — fresh body built on the cloned styles of `03-custom_styles.docx` (Q3 report with `Title`/`Subtitle`/`Heading1`/`Heading2` paragraphs and a metrics table).
- `open-template-appendix.docx` — `04-custom_styles_plus_sample_text.docx`'s original paragraphs preserved, plus an appended `Heading1`-style heading, a paragraph, and a 4-column builder table.
- `template-merge-invoice.docx` — `FromTemplate` + `Merge` + `InsertBefore` combined to build an invoice with header/body/table placeholders filled and a "Total Due" paragraph inserted before the first empty paragraph.

## Code walkthrough

- `testdataPath(parts...)` — helper that uses `runtime.Caller(0)` to find the source-file directory, then walks up twice (`Dir(Dir(filename))`) to reach the repo root, and joins with `testdata/word/<file>`. This is why the example must be run from its own directory.
- `doc1, err := wordingo.FromTemplate(customStylesPath)` — opens the template, clones `styles.xml`, `numbering.xml`, `fontTable.xml`, `theme1.xml`, `settings.xml` byte-for-byte into a fresh OPC package, and replaces the body with an empty single-section body. Style parts are never re-encoded after cloning.
- `doc1.AddParagraph("Q3 Financial Report").SetStyle("Title")` — uses the style IDs from the cloned `styles.xml`.
- `doc2, err := wordingo.OpenTemplate(samplePath)` — same style-clone pass, but the template's existing body paragraphs and tables are preserved. Header/footer `rId` references in `sectPr` are stripped to avoid dangling rIds in the cloned package; the source `PgSz`/`PgMar` are carried over.
- `printStyleIDs(path)` — diagnostic helper that opens the .docx as a ZIP, reads `word/styles.xml`, and regex-extracts every `w:styleId="..."` value to stdout. Demonstrates the kind of inspection you can do with the standard library before deciding which style IDs to use.
- `doc3` example: `wordingo.FromTemplate(customStylesPath)` again, then `h.AddParagraph("Invoice {{invoice_id}}")` and body/table `{{...}}` placeholders, then `doc3.Merge(map[string]string{...}, nil)` fills `invoice_id`, `customer_name`, and per-line item/qty/price/total placeholders. `doc3.InsertBefore(totalPara, "Total Due: $8,500")` adds the totals line before the first empty paragraph.