package wordingo

import (
	"bytes"
	"strings"
	"testing"
)

func TestEdit_InsertBefore(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	second := doc.AddParagraph("Second")
	first := doc.InsertBefore(second, "First")
	if first == nil {
		t.Fatal("InsertBefore returned nil")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	p1 := strings.Index(content, "First")
	p2 := strings.Index(content, "Second")
	if p1 == -1 {
		t.Error("expected First in output")
	}
	if p2 == -1 {
		t.Error("expected Second in output")
	}
	if p1 > p2 {
		t.Error("First should appear before Second")
	}
}

func TestEdit_InsertAfter(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	first := doc.AddParagraph("First")
	second := doc.InsertAfter(first, "Second")
	if second == nil {
		t.Fatal("InsertAfter returned nil")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	p1 := strings.Index(content, "First")
	p2 := strings.Index(content, "Second")
	if p1 == -1 {
		t.Error("expected First in output")
	}
	if p2 == -1 {
		t.Error("expected Second in output")
	}
	if p1 > p2 {
		t.Error("First should appear before Second")
	}
}

func TestEdit_DeleteParagraph(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p1 := doc.AddParagraph("P1")
	p2 := doc.AddParagraph("P2")
	p3 := doc.AddParagraph("P3")
	doc.DeleteParagraph(p2)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(content, "P1") {
		t.Error("expected P1 in output")
	}
	if !strings.Contains(content, "P3") {
		t.Error("expected P3 in output")
	}
	if strings.Contains(content, "P2") {
		t.Error("P2 should be deleted")
	}
	_ = p1
	_ = p3
}

func TestEdit_DeleteParagraphNotFound(t *testing.T) {
	docA, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	docB, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := docB.AddParagraph("orphan")
	docA.DeleteParagraph(p)

	warnings := docA.Warnings()
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "not found") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning for not-found paragraph, got %v", warnings)
	}
}

func TestEdit_DeleteRow(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	tb, err := doc.AddTable([][]string{{"A"}, {"B"}, {"C"}})
	if err != nil {
		t.Fatal(err)
	}
	err = tb.DeleteRow(1)
	if err != nil {
		t.Fatalf("DeleteRow error: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(content, "A") {
		t.Error("expected A in output")
	}
	if !strings.Contains(content, "C") {
		t.Error("expected C in output")
	}
	if strings.Contains(content, "B") {
		t.Error("B row should be deleted")
	}
	if strings.Count(content, "<w:tr") != 2 {
		t.Errorf("expected 2 rows, got content with B still present")
	}
}

func TestEdit_DeleteRowOutOfBounds(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	tb, err := doc.AddTable([][]string{{"only"}})
	if err != nil {
		t.Fatal(err)
	}
	err = tb.DeleteRow(5)
	if err == nil {
		t.Fatal("expected error for out-of-bounds DeleteRow")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Errorf("error should mention out of range, got: %v", err)
	}
}

func TestEdit_SetText(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("old")
	if len(p.ct.R) == 0 {
		t.Fatal("no runs in paragraph")
	}
	run := &Run{ct: p.ct.R[0], para: p}
	run.SetText("new")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(content, ">new<") {
		t.Errorf("expected 'new' in output, got %s", content)
	}
	if strings.Contains(content, ">old<") {
		t.Errorf("should not contain 'old' after SetText, got %s", content)
	}
}

func TestEdit_ReplaceText(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("hello world")
	if len(p.ct.R) == 0 {
		t.Fatal("no runs in paragraph")
	}
	run := &Run{ct: p.ct.R[0], para: p}
	run.ReplaceText("world", "there")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(content, "hello there") {
		t.Errorf("expected 'hello there' in output, got %s", content)
	}
}

func TestEdit_ReplaceTextNoMatch(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("hello")
	if len(p.ct.R) == 0 {
		t.Fatal("no runs")
	}
	run := &Run{ct: p.ct.R[0], para: p}
	run.ReplaceText("xyz", "abc")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(content, "hello") {
		t.Errorf("expected 'hello' unchanged, got %s", content)
	}
}

func TestEdit_StyleRoundTrip(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	first := doc.AddParagraph("original")

	var before bytes.Buffer
	if _, err := doc.WriteTo(&before); err != nil {
		t.Fatal(err)
	}

	paras := doc.Paragraphs()
	doc.InsertBefore(paras[0], "preamble")
	doc.DeleteParagraph(first)
	doc.AddParagraph("appendix")

	var after bytes.Buffer
	if _, err := doc.WriteTo(&after); err != nil {
		t.Fatal(err)
	}

	// Only check non-body parts for byte-identity (STYLE-ROUNDTRIP)
	styleBefore := readZipEntryFromBuf(t, before.Bytes(), "word/styles.xml")
	styleAfter := readZipEntryFromBuf(t, after.Bytes(), "word/styles.xml")
	if styleBefore != styleAfter {
		t.Error("styles.xml changed after edit operations")
	}
}
