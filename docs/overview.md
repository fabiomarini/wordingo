# wordingo

Pure Go library for creating and editing Microsoft Word (.docx) documents. Zero external dependencies.

## What it is

wordingo builds Word documents from scratch, opens existing ones for editing, converts between Word and Markdown, and handles template-based document generation — all in pure Go with no C bindings, no system calls to Word, and no external binaries.

## Why use it?

- **No external dependencies.** Only the Go standard library. No XML libraries, no Office interop, no Docker containers.
- **Faithful round-trips.** Open a Word document, make targeted edits, and save — everything you didn't touch stays identical to the original.
- **Template-driven.** Clone a template's styles, fonts, colors, and numbering into a fresh document, or keep the template's existing content and extend it.
- **Opens clean in Word.** Every output file opens in Microsoft Word without repair prompts.
- **Escape hatch.** If the high-level API doesn't expose a particular XML attribute, the `X()` methods give you direct access to the underlying XML structs.

## Key features

| Feature | Description |
|---------|-------------|
| Create & open | Blank documents or existing .docx files |
| Templates | Clone styles/fonts/colors from any template |
| Styled text | Bold, italic, underline, color, highlights, font, size |
| Paragraph styles | Named styles (Title, Heading1–9, Subtitle, Quote, etc.) |
| Tables | Quick grid or full builder with borders, shading, merged cells |
| Images | Embed PNG/JPEG with auto DPI detection |
| Headers & footers | Default, first-page, even-page variants |
| Lists | Bulleted, ordered, nested up to 9 levels |
| Hyperlinks | Clickable links with styled text |
| Page setup | Portrait/landscape, Letter/A4/Legal, custom margins |
| Template merge | Replace `{{placeholders}}` in body, tables, headers, footers |
| Edit operations | Insert/delete paragraphs by reference, replace text, delete rows |
| Text extraction | Pull plain text from body/tables/headers/footers |
| Markdown export | .docx → GitHub-flavored Markdown |
| Markdown import | Markdown → .docx (headings, lists, tables, code blocks) |

## Design principles

- **Stdlib only.** `go.mod` has no `require` directives.
- **Spec-driven.** Built on the ISO OOXML standard. Unknown XML elements are preserved verbatim, never lossy-parsed.
- **Fidelity first.** Templates are cloned byte-for-byte — styles, numbering, fonts, theme, settings are never re-encoded.
- **Escape hatch over bloat.** The `X()` methods expose raw XML structs for anything the public API doesn't cover, keeping the high-level surface focused.

## When to use a template

- **`FromTemplate`** — copies the template's formatting into an empty document. You write all the content.
- **`OpenTemplate`** — copies the formatting AND keeps the template's existing paragraphs and tables. You can append, edit, or delete.
- **`Open`** — opens any .docx for editing. No style cloning. Paragraphs and tables are preserved as-is.
