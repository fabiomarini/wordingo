# Example 09: Merge and Edit

## What it demonstrates

Phase 6 editing surface: template merge (`{{placeholder}}` replacement in body, table cells, headers, footers), pointer-identity `InsertBefore`/`InsertAfter`/`DeleteParagraph`, run `SetText`/`ReplaceText`, `TableBuilder.DeleteRow` with bounds safety, scoped merge (`MergeOpts.ScopedParts`), `Body()` ordered iteration, and the `Warnings()` surface for unused merge keys.

## How to run

```
cd examples/09-merge-and-edit
go run main.go
```

## Output

- `output.docx` — the fully merged and edited document: header/footer placeholders filled, table rows merged and row 2 deleted, paragraphs inserted around the intro line, runs' text replaced, and a second scoped merge updating only the headers/footers.

## Code walkthrough

- `h := doc.AddHeader(wordingo.HeaderDefault); h.AddParagraph("Report: {{report_name}} — {{date}}")` — header paragraph with `{{key}}` placeholders that `Merge` will later substitute.
- `doc.Merge(map[string]string{...}, nil)` — replaces every `{{key}}` in body paragraphs, table cells, header paragraphs, and footer paragraphs (split-run safe: placeholders spanning multiple runs are detected via a char-offset map and rewritten). `nil` opts defaults to scanning all four scopes. Unused keys surface as warnings via `doc.Warnings()`.
- `doc.InsertBefore(introPara, "Dear Alice,")` and `doc.InsertAfter(introPara, "Please find the details below.")` — both locate `introPara` by **pointer identity** (`p.ct == target.ct`) within the body's paragraphs and splice a new paragraph in place; not-found targets record a warning and return `nil` instead of mutating.
- `p.AddRun(" (updated)").SetText(" (revised)")` and `p.AddRun(" Welcome!").SetText(" Welcome to the team!")` — `Run.SetText` overwrites the run's `w:t`; `Run.ReplaceText` does a `strings.ReplaceAll` on the existing text.
- `doc.DeleteParagraph(toDelete)` — removes `toDelete` from the body by pointer identity, preserving the `ElemOrder` mapping so subsequent `Body()` iteration stays consistent.
- `for _, el := range doc.Body() { switch el.Type { case wordingo.ElementParagraph: ...; case wordingo.ElementTable: ... } }` — iterate the body in true document order, mixing paragraphs and tables via the `BodyElement` union (`Type` plus `Para`/`Table`). Counts paragraphs and tables to summarise in a follow-up paragraph.
- `tables := doc.Tables(); tables[0].DeleteRow(2)` — `TableBuilder.DeleteRow` returns an error on out-of-bounds (it does **not** panic); the example `log.Fatalf`s on the returned error.
- `doc.Merge(map[string]string{"page_num": "2"}, &wordingo.MergeOpts{ScopedParts: wordingo.ScopedParts{Headers: true, Footers: true}})` — second merge touches only header/footer paragraphs, leaving body/table content untouched.
- `doc.Warnings()` at the end — non-fatal issues from both merges (e.g., keys not present in the scoped scope become "not found" warnings).