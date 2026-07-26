# Example 07: Hyperlinks

## What it demonstrates

External hyperlinks inside paragraphs with chained run-style formatting on the returned `*Run`. Shows single-link paragraphs, multi-link paragraphs, and the `SetBold`/`SetColor`/`SetUnderline` chainers applied to hyperlink runs.

## How to run

```
cd examples/07-hyperlinks
go run main.go
```

## Output

- `output.docx` — three paragraphs containing inline hyperlinks to `go.dev`, `pkg.go.dev`, `google.com`, `github.com`, and `stackoverflow.com`, each with run formatting.

## Code walkthrough

- `p := doc.AddParagraph("Visit the ")` — start a paragraph with leading text; the `Paragraph` will hold both runs and hyperlinks.
- `p.AddHyperlink("Go Programming Language", "https://go.dev").SetColor("0563C1").SetUnderline("single")` — `AddHyperlink` creates a new external OPC relationship in `word/_rels/document.xml.rels` with `TargetMode="External"`, appends a `w:hyperlink` element wrapping one run with the display text, and returns a `*Run` supporting the same chainers as a normal run. Each call allocates a fresh `rId`; URIs are not deduplicated.
- `p.AddRun(" website to download the latest version.")` — mix hyperlinks with plain runs in the same paragraph.
- `doc2 := doc.AddParagraph("For documentation, see ")` followed by `doc2.AddHyperlink("pkg.go.dev", "https://pkg.go.dev").SetBold(true).SetColor("0563C1")` — demonstrates `SetBold` on a hyperlink run alongside the more common `SetColor`/`SetUnderline`.
- Multiple hyperlinks in one paragraph: `para.AddHyperlink("Google", ...).SetColor("2E75B6")`, interleaved with `para.AddRun(", ")` separators — wordingo reads and appends `w:hyperlink` and `w:r` siblings in document order.