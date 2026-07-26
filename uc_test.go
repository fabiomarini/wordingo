package wordingo

import (
	"bytes"
	"strings"
	"testing"
)

func TestUC1_TemplateMerge(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("Hello {{name}}, today is {{date}}")
	doc.AddParagraph("Subject: {{title}}")
	_, err = doc.AddTable([][]string{{"Value: {{value}}"}})
	if err != nil {
		t.Fatal(err)
	}

	doc.Merge(map[string]string{
		"name":  "Alice",
		"date":  "2026-07-26",
		"title": "Report",
		"value": "42",
	}, nil)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reopened, err := OpenReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}

	texts := reopened.Paragraphs()
	if len(texts) < 3 {
		t.Fatalf("expected >=3 paragraphs (empty + 2 content), got %d", len(texts))
	}
	foundAlice := false
	foundReport := false
	for _, p := range texts {
		if strings.Contains(p.Text(), "Alice") {
			foundAlice = true
		}
		if strings.Contains(p.Text(), "Report") {
			foundReport = true
		}
	}
	if !foundAlice {
		t.Errorf("Alice not found in paragraphs")
	}
	if !foundReport {
		t.Errorf("Report not found in paragraphs")
	}

	warnings := doc.Warnings()
	for _, w := range warnings {
		if strings.Contains(w, "not found") {
			t.Errorf("unexpected warning on happy path: %s", w)
		}
	}
}

func TestUC2_MergeWithMissingKeys(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("Hello {{present}}")
	doc.Merge(map[string]string{"present": "here", "missing": "gone"}, nil)

	warnings := doc.Warnings()
	foundMissing := false
	foundPresent := false
	for _, w := range warnings {
		if strings.Contains(w, "missing") {
			foundMissing = true
		}
		if strings.Contains(w, "present") {
			foundPresent = true
		}
	}
	if !foundMissing {
		t.Errorf("expected warning for 'missing', got %v", warnings)
	}
	if foundPresent {
		t.Errorf("unexpected warning for 'present': %v", warnings)
	}
}

func TestUC3_EditOperations(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	paras := doc.Paragraphs()
	bodyPara := doc.AddParagraph("Body {{name}}")
	doc.Merge(map[string]string{"name": "Alice"}, nil)

	firstPara := doc.InsertBefore(paras[0], "Preamble")
	if firstPara == nil {
		t.Fatal("InsertBefore returned nil")
	}

	doc.AddParagraph("Appendix")

	delPara := doc.AddParagraph("ToDelete")
	doc.DeleteParagraph(delPara)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reopened, err := OpenReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}

	allText := reopened.Paragraphs()
	if len(allText) < 3 {
		t.Fatalf("expected at least 3 paragraphs, got %d", len(allText))
	}

	foundPreamble := false
	foundAlice := false
	for _, p := range allText {
		if strings.Contains(p.Text(), "Preamble") {
			foundPreamble = true
		}
		if strings.Contains(p.Text(), "Alice") {
			foundAlice = true
		}
		if strings.Contains(p.Text(), "ToDelete") {
			t.Errorf("ToDelete should not appear")
		}
	}
	if !foundPreamble {
		t.Errorf("Preamble not found in paragraphs")
	}
	if !foundAlice {
		t.Errorf("Alice should be present after merge")
	}

	_ = bodyPara
}

func TestUC4_TableEditAndMerge(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	tb, err := doc.AddTable([][]string{{"{{a}}"}, {"{{b}}"}, {"{{c}}"}})
	if err != nil {
		t.Fatal(err)
	}

	doc.Merge(map[string]string{"a": "A", "b": "B", "c": "C"}, nil)

	err = tb.DeleteRow(1)
	if err != nil {
		t.Fatalf("DeleteRow: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reopened, err := OpenReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}

	tables := reopened.Tables()
	if len(tables) == 0 {
		t.Fatal("no tables found")
	}

	docContent := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(docContent, ">A<") {
		t.Errorf("expected A in output")
	}
	if !strings.Contains(docContent, ">C<") {
		t.Errorf("expected C in output")
	}
	if strings.Contains(docContent, ">B<") {
		t.Errorf("B should be deleted")
	}
}
