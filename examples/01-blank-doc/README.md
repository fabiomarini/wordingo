# Example 01: Blank Doc

## What it demonstrates

The absolute minimum wordingo program: create a blank .docx, add one paragraph, save to disk. Exercises `wordingo.Create()`, `Document.AddParagraph`, `Document.Save`, and `Document.Close` — the smallest end-to-end pipeline that produces a valid OOXML package.

## How to run

```
cd examples/01-blank-doc
go run main.go
```

## Output

- `output.docx` — a single-paragraph blank document with the default Letter page size and 1-inch margins.

## Code walkthrough

- `doc, err := wordingo.Create()` — builds a blank OPC package with default `styles.xml`, `theme1.xml`, `fontTable.xml`, `settings.xml`, `webSettings.xml`, and one section (Letter, 1-inch margins). Returns `(*Document, error)`.
- `defer doc.Close()` — releases the package and document references; the `Document` is not usable after `Close`.
- `doc.AddParagraph("Hello, World!")` — appends a `w:p` to the body with one run carrying the text; returns `*Paragraph` (ignored here).
- `doc.Save(out)` — calls `WriteTo` into a file at the given path; serializes the OPC package to a ZIP `.docx`.
- `fi, _ := os.Stat(out); fmt.Printf("Created %s (%d bytes)\n", out, fi.Size())` — reports the on-disk byte size of the produced `.docx`.