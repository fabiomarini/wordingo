# Example 11: Text Extraction & Markdown

## What it demonstrates

Phase 06.1 text/markdown surface: `ExtractText` (plain text from body/tables/headers/footers with scope control and custom separator), `ToMarkdown` (GFM export with headings, inline formatting, pipe tables, hyperlinks, image data URIs, list markers, code blocks), `CreateFromMarkdown` (new `Document` from a markdown string), `ImportMarkdown` (append markdown to an existing doc), scoped parts control, and round-trip verification (~18 `verifyContains`/`verify*` checks).

## How to run

```
cd examples/11-text-extraction-markdown
go run main.go
```

## Output

- `output.docx` — the source document built by `Create()` with header/footer, styled headings, inline formatting, a table, lists, a code block, and an inline image.
- `output.txt` — `doc.ExtractText(nil)` result for the source document (all parts, `"\n"` separator).
- `output.md` — `doc.ToMarkdown(nil)` GFM export of the source document.
- `imported-source.md` — the source markdown string passed to `CreateFromMarkdown`.
- `imported-output.txt` — `imported.ExtractText(nil)` round-trip text (markdown → doc → text).
- `imported-output.md` — `imported.ToMarkdown(nil)` round-trip markdown (after `ImportMarkdown` appends more content).

## Code walkthrough

- `text, err := doc.ExtractText(nil)` — pulls plain text from body, tables, headers, and footers, joined by `"\n"` (the default when `Separator` is empty). Read-only: does not set `d.dirty`.
- `doc.ExtractText(&wordingo.ExtractOpts{ScopedParts: wordingo.ScopedParts{Body:true, Tables:false, Headers:false, Footers:false}})` — toggle which parts are extracted per call; `ScopedParts` is the same struct used by `Merge` and `ToMarkdown` for consistent scope control.
- `doc.ExtractText(&wordingo.ExtractOpts{ScopedParts:..., Separator:" | "})` — custom separator between extracted chunks (the bullet example uses `" | "`).
- `md, err := doc.ToMarkdown(nil)` — emits GFM: `Heading1`..`Heading6` paragraph styles become `#`..`######`; runs emit `**bold**` / `*italic*` / `~~strike~~`; tables become pipe tables with a separator row; hyperlinks resolve to `[text](url)` from the document relationships; inline images become `![alt](data:image/png;base64,...)` data URIs; list paragraphs emit `1.`/`-` markers; paragraphs whose runs use `Consolas`/`Courier`/`Courier New` emit fenced code blocks.
- `imported, err := wordingo.CreateFromMarkdown(sourceMD)` — builds a brand-new `Document` from a markdown string. Headings map to `Heading1`..`Heading6` paragraph styles; ordered/bulleted lists (with nesting) get numbering definitions merged into `word/numbering.xml`; pipe tables become grid tables; fenced code blocks become `Consolas` paragraphs; inline bold/italic/code/links/data-URI images are applied via `Run` formatting / `AddHyperlink` / `AddImageBytes`.
- `imported.ImportMarkdown("\n## Appended Section\n\n...")` — parses and appends more markdown content to an already-built document. Parse errors are surfaced via `doc.Warnings()` and never panic.
- `verifyContains(s, substr, label)` and `verifyExtractText`/`verifyToMarkdown` helpers run ~18 round-trip checks: extracted text contains `bold text`, `italic text`, `the project repo`, `ExtractText` (table contents), and `Text Extraction & Markdown Demo` (title); markdown contains `**bold text**`, `*italic text*`, `| ExtractText | Done | 0.6.0 |` (pipe table row), the document title, etc.; imported-roundtrip markdown contains `**created**`, `*markdown*`, `Appended Section`. Note the `makeGradientPNG` helper here builds a 50×50 image (smaller than Example 04's 100×100) — the aspect-locked 1.5×1.5-inch display size is set via `run.SetImageWidth(1.5).SetImageHeight(1.5)`.