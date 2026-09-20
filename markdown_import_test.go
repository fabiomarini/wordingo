package wordingo

import (
	"bytes"
	"strings"
	"testing"
)

// Table cells must keep inline markdown formatting through the import:
// AddTable alone writes plain runs, so **bold** used to land in the docx
// verbatim.
func TestMarkdownImport_TableCellFormatting(t *testing.T) {
	md := "| Oggetto | Tipo | Ruolo |\n" +
		"| --- | --- | --- |\n" +
		"| **PY59SG01** | Application | Interfaccia **utente** |\n" +
		"| **NY59SG02** | NER | stato `910` |\n"

	doc, err := CreateFromMarkdown(md)
	if err != nil {
		t.Fatalf("CreateFromMarkdown: %v", err)
	}
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	doc.Close()

	back, err := OpenReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer back.Close()

	out, err := back.ToMarkdown(nil)
	if err != nil {
		t.Fatalf("ToMarkdown: %v", err)
	}
	for _, want := range []string{"**PY59SG01**", "**utente**", "**NY59SG02**"} {
		if !strings.Contains(out, want) {
			t.Errorf("round-trip lost cell formatting %q; got:\n%s", want, out)
		}
	}
}

// Paragraph formatting must keep working next to the table-cell path.
func TestMarkdownImport_ParagraphBoldStillWorks(t *testing.T) {
	doc, err := CreateFromMarkdown("Test **bold** text.\n\n| a | b |\n| --- | --- |\n| **c** | d |\n")
	if err != nil {
		t.Fatalf("CreateFromMarkdown: %v", err)
	}
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	doc.Close()

	back, err := OpenReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer back.Close()

	out, err := back.ToMarkdown(nil)
	if err != nil {
		t.Fatalf("ToMarkdown: %v", err)
	}
	if !strings.Contains(out, "**bold**") {
		t.Errorf("paragraph bold lost; got:\n%s", out)
	}
	if !strings.Contains(out, "**c**") {
		t.Errorf("table cell bold lost; got:\n%s", out)
	}
}
