# Example 02: Text and Styles

## What it demonstrates

Paragraph styles plus inline run formatting on a blank document. Builds a short "article" with `Title`, `Subtitle`, `Heading1`-`Heading3`-style paragraphs and demonstrations of `SetBold`, `SetItalic`, `SetColor`.

## How to run

```
cd examples/02-text-and-styles
go run main.go
```

## Output

- `output.docx` — a multi-paragraph document with styled headings and inline-formatted runs.

## Code walkthrough

- `doc.AddParagraph("The Art of Go").SetStyle("Title")` — sets the paragraph style ID to `"Title"`, matching the default `Title` style in the blank document.
- `p := doc.AddParagraph("Go offers a clean syntax...")` — returns a `*Paragraph` so you can append runs to an existing paragraph instead of starting a fresh one.
- `p.AddRun(" This sentence is bold.").SetBold(true)` — appends a new run with the given text and chains run formatting; every `Run` setter returns `*Run` for chaining.
- `r := doc.AddParagraph("")` followed by `r.AddRun("Compiled to native code — ").SetColor("666666")` then `r.AddRun("lightning fast execution").SetBold(true).SetColor("2E75B6")` — multiple runs in one paragraph with distinct formatting; `SetColor` validates a 6-hex-digit value and warns on anything else.

