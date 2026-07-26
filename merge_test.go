package wordingo

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/fabiomarini/wordingo/internal/wml"
)

func TestMerge_Body(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("Hello {{name}}, your {{item}} is ready")
	doc.Merge(map[string]string{"name": "Alice", "item": "report"}, nil)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(content, "Hello Alice, your report is ready") {
		t.Errorf("body merge: got %s", content)
	}
}

func TestMerge_TableCell(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	_, err = doc.AddTable([][]string{{"{{key}}", "{{value}}"}})
	if err != nil {
		t.Fatal(err)
	}
	doc.Merge(map[string]string{"key": "Name", "value": "42"}, nil)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(content, "Name") {
		t.Errorf("table cell merge missing Name")
	}
	if !strings.Contains(content, "42") {
		t.Errorf("table cell merge missing 42")
	}
}

func TestMerge_HeaderFooter(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	h := doc.AddHeader(HeaderDefault)
	h.AddParagraph("{{title}}")
	f := doc.AddFooter(FooterDefault)
	f.AddParagraph("Page {{num}}")

	doc.Merge(map[string]string{"title": "Report", "num": "1"}, nil)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	// Find the header part in the zip
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	headerFound := false
	footerFound := false
	for _, zf := range zr.File {
		if strings.Contains(zf.Name, "header") {
			rc, err := zf.Open()
			if err != nil {
				continue
			}
			var hdrBuf bytes.Buffer
			hdrBuf.ReadFrom(rc)
			rc.Close()
			if strings.Contains(hdrBuf.String(), "Report") {
				headerFound = true
			}
		}
		if strings.Contains(zf.Name, "footer") {
			rc, err := zf.Open()
			if err != nil {
				continue
			}
			var ftrBuf bytes.Buffer
			ftrBuf.ReadFrom(rc)
			rc.Close()
			if strings.Contains(ftrBuf.String(), "1") {
				footerFound = true
			}
		}
	}
	if !headerFound {
		t.Error("header part missing merged title")
	}
	if !footerFound {
		t.Error("footer part missing merged page number")
	}
}

func TestMerge_SplitRun(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("")
	p.ct.R = []*wml.CT_R{
		{RPr: &wml.CT_RPr{B: &wml.CT_OnOff{Val: &[]bool{true}[0]}}, T: &wml.CT_Text{Value: "Hello {{"}},
		{T: &wml.CT_Text{Value: "na"}},
		{RPr: &wml.CT_RPr{I: &wml.CT_OnOff{Val: &[]bool{true}[0]}}, T: &wml.CT_Text{Value: "me}}!"}},
	}
	doc.Merge(map[string]string{"name": "Alice"}, nil)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	if !strings.Contains(content, "Hello Alice!") {
		t.Errorf("split-run merge: got %s", content)
	}
	if p.ct.R[1].T.Value != "" {
		t.Errorf("expected run 1 text to be empty, got %q", p.ct.R[1].T.Value)
	}
	if p.ct.R[2].T.Value != "" {
		t.Errorf("expected run 2 text to be empty, got %q", p.ct.R[2].T.Value)
	}
}

func TestMerge_SplitRunWithNonTextRuns(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("")
	p.ct.R = []*wml.CT_R{
		{T: &wml.CT_Text{Value: "hello {{"}},
		{Br: &wml.CT_Br{}},
		{T: &wml.CT_Text{Value: "key}} world"}},
	}
	doc.Merge(map[string]string{"key": "test"}, nil)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	if !strings.Contains(content, "hello test world") {
		t.Errorf("split-run with br merge: got %s", content)
	}
	if !strings.Contains(content, "<w:br") {
		t.Error("br element not preserved")
	}
	if p.ct.R[0].T.Value != "hello test world" {
		t.Errorf("expected run 0 text 'hello test world', got %q", p.ct.R[0].T.Value)
	}
	if p.ct.R[2].T.Value != "" {
		t.Errorf("expected run 2 text to be empty, got %q", p.ct.R[2].T.Value)
	}
}

func TestMerge_MissingKey(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("Hello {{name}}")
	doc.Merge(map[string]string{"name": "Alice", "unused_key": "val"}, nil)

	warnings := doc.Warnings()
	foundUnused := false
	foundName := false
	for _, w := range warnings {
		if strings.Contains(w, "unused_key") {
			foundUnused = true
		}
		if strings.Contains(w, "name") {
			foundName = true
		}
	}
	if !foundUnused {
		t.Errorf("expected warning about unused_key, got %v", warnings)
	}
	if foundName {
		t.Errorf("unexpected warning about name (used key): %v", warnings)
	}
}

func TestMerge_NilOpts(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("{{a}}")
	h := doc.AddHeader(HeaderDefault)
	h.AddParagraph("{{b}}")
	doc.Merge(map[string]string{"a": "A", "b": "B"}, nil)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(content, "A") {
		t.Errorf("nil opts body merge: got %s", content)
	}
}

func TestMerge_ScopedParts(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("{{a}}")
	h := doc.AddHeader(HeaderDefault)
	h.AddParagraph("{{b}}")

	doc.Merge(map[string]string{"a": "A", "b": "B"}, &MergeOpts{
		ScopedParts: ScopedParts{Body: true, Tables: false, Headers: false, Footers: false},
	})

	warnings := doc.Warnings()
	foundB := false
	for _, w := range warnings {
		if strings.Contains(w, "b") {
			foundB = true
		}
	}
	if !foundB {
		t.Errorf("expected warning about 'b' (headers not scanned), got %v", warnings)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(content, "A") {
		t.Errorf("scoped body merge: got %s", content)
	}
}

func TestMerge_KeyNotFoundInDocument(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("no placeholders here")
	doc.Merge(map[string]string{"missing": "val"}, nil)

	warnings := doc.Warnings()
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "missing") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning about missing key, got %v", warnings)
	}
}
