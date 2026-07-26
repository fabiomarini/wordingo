package wordingo

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/fabiomarini/wordingo/internal/wml"
)

func TestSetStyleParagraph(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("styled")
	p.SetStyle("Title")
	if p.ct.PPr == nil || p.ct.PPr.PStyle == nil || p.ct.PPr.PStyle.Val == nil {
		t.Fatal("PStyle not set")
	}
	if *p.ct.PPr.PStyle.Val != "Title" {
		t.Errorf("PStyle.Val = %q, want Title", *p.ct.PPr.PStyle.Val)
	}
}

func TestSetStyleRun(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("test")
	r := p.AddRun("emphasized")
	r.SetStyle("Emphasis")
	if r.ct.RPr == nil || r.ct.RPr.RStyle == nil || r.ct.RPr.RStyle.Val == nil {
		t.Fatal("RStyle not set")
	}
	if *r.ct.RPr.RStyle.Val != "Emphasis" {
		t.Errorf("RStyle.Val = %q, want Emphasis", *r.ct.RPr.RStyle.Val)
	}
}

func TestSetStyleClearParagraph(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("styled")
	p.SetStyle("Title")
	p.SetStyle("")
	if p.ct.PPr.PStyle != nil {
		t.Error("SetStyle('') should clear PStyle")
	}
}

func TestSetStyleClearRun(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("test")
	r := p.AddRun("emphasized")
	r.SetStyle("Emphasis")
	r.SetStyle("")
	if r.ct.RPr.RStyle != nil {
		t.Error("SetStyle('') should clear RStyle")
	}
}

func TestSetStyleOnExistingParagraph(t *testing.T) {
	data := buildSyntheticFixture(t, string(fixtureMultiHeadingXML(t)), nil)
	doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	paras := doc.Paragraphs()
	if len(paras) < 1 {
		t.Fatal("expected at least 1 paragraph")
	}

	paras[0].SetStyle("Heading1")
	if paras[0].Style() != "Heading1" {
		t.Errorf("Style = %q, want Heading1", paras[0].Style())
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	doc2, err := OpenReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if doc2.Paragraphs()[0].Style() != "Heading1" {
		t.Errorf("re-opened: Style = %q, want Heading1", doc2.Paragraphs()[0].Style())
	}
}

func TestStyleWarningOnSave(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("bad")
	p.SetStyle("NonExistentStyle")

	r := p.AddRun("run")
	r.SetStyle("MissingRunStyle")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	w := doc.Warnings()
	foundUnknown := false
	for _, msg := range w {
		if strings.Contains(msg, "unknown style") {
			foundUnknown = true
			break
		}
	}
	if !foundUnknown {
		t.Errorf("expected unknown style warnings, got: %v", w)
	}
}

func TestStyleWarningXMLIllegal(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("bad")
	p.SetStyle("Bad\x00Style")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	w := doc.Warnings()
	found := false
	for _, msg := range w {
		if strings.Contains(msg, "XML-illegal") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected XML-illegal warning, got: %v", w)
	}
}

func TestStyleNilReceiverPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil *Paragraph.SetStyle")
		}
	}()
	(*Paragraph)(nil).SetStyle("Title")
}

func TestStylePartsUntouched(t *testing.T) {
	data := buildSyntheticFixture(t, string(fixtureSingleParaXML(t)), nil)
	doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	paras := doc.Paragraphs()
	paras[0].SetStyle("Heading1")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	pa := partPayloads(t, data)
	pb := partPayloads(t, buf.Bytes())

	styleNames := []string{"word/styles.xml"}
	for _, name := range styleNames {
		orig, ok := pa[name]
		if !ok {
			continue
		}
		saved, ok := pb[name]
		if !ok {
			t.Errorf("style part %q missing from output", name)
			continue
		}
		if !bytes.Equal(orig, saved) {
			t.Errorf("style part %q differs after SetStyle", name)
		}
	}
}

func TestQualRunX(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("test")
	r := p.AddRun("x")
	ct := r.X()
	if ct == nil {
		t.Fatal("Run.X() returned nil")
	}
	if _, ok := interface{}(ct).(*wml.CT_R); !ok {
		t.Error("Run.X() should return *wml.CT_R")
	}
}

func TestQualParagraphX(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("test")
	ct := p.X()
	if ct == nil {
		t.Fatal("Paragraph.X() returned nil")
	}
}

func TestQualDocumentX(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	pkg := doc.X()
	if pkg == nil {
		t.Fatal("Document.X() returned nil")
	}
}

func TestFullContentDoc(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	p1 := doc.AddParagraph("Title")
	p1.SetStyle("Title")
	p1.SetAlignment(AlignmentCenter)

	p2 := doc.AddParagraph("Body text")
	p2.AddRun("bold text").SetBold(true)
	p2.AddRun(" red").SetColor("FF0000").SetSize(12)
	p2.SetSpacing(&ParSpacing{Before: 120, After: 60})
	p2.SetIndent(&ParIndent{Left: 720})

	p3 := doc.AddParagraph("Styled run")
	p3.AddRun("emphasis").SetStyle("Emphasis")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	doc2, err := OpenReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}

	paras := doc2.Paragraphs()
	if len(paras) != 4 {
		t.Fatalf("expected 4 paragraphs, got %d", len(paras))
	}
	if paras[1].Text() != "Title" {
		t.Errorf("para[1] text = %q, want Title", paras[1].Text())
	}
	if paras[2].Text() != "Body textbold text red" {
		t.Errorf("para[2] text = %q, want 'Body textbold text red'", paras[2].Text())
	}
	if paras[3].Text() != "Styled runemphasis" {
		t.Errorf("para[3] text = %q, want 'Styled runemphasis'", paras[3].Text())
	}

	w := doc.Warnings()
	if len(w) > 0 {
		nonStyleWarnings := 0
		for _, msg := range w {
			if !strings.Contains(msg, "unknown style") {
				nonStyleWarnings++
			}
		}
		if nonStyleWarnings > 0 {
			t.Errorf("unexpected warnings: %v", w)
		}
	}
}

func TestFromTemplateWithStyle(t *testing.T) {
	fixture := buildTemplateFixture(t, 3, false)
	doc, err := FromTemplateReader(bytes.NewReader(fixture), int64(len(fixture)))
	if err != nil {
		t.Fatalf("FromTemplateReader: %v", err)
	}

	p := doc.AddParagraph("Template styled")
	p.SetStyle("Heading1")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	doc2, err := OpenReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	paras := doc2.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("expected 1 paragraph, got %d", len(paras))
	}
	if paras[0].Style() != "Heading1" {
		t.Errorf("Style = %q, want Heading1", paras[0].Style())
	}
}

func TestSetStyleNormalisesEmpty(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("test")
	p.SetStyle("Normal")
	if p.Style() != "Normal" {
		t.Errorf("Style = %q, want Normal", p.Style())
	}

	p.SetStyle("")
	if p.Style() != "" {
		t.Errorf("Style = %q after clear, want empty", p.Style())
	}
}

func TestStyleRunOnOpenedDoc(t *testing.T) {
	data := buildSyntheticFixture(t, string(fixtureSingleParaXML(t)), nil)
	doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	paras := doc.Paragraphs()
	r := paras[0].AddRun(" new")
	r.SetStyle("Emphasis")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	pb := partPayloads(t, buf.Bytes())
	docXML, ok := pb["word/document.xml"]
	if !ok {
		t.Fatal("word/document.xml missing from output")
	}
	if !strings.Contains(string(docXML), "Emphasis") {
		t.Error("rStyle reference missing from document.xml")
	}
}

func TestStyleValidation(t *testing.T) {
	data := buildSyntheticFixture(t, string(fixtureMultiHeadingXML(t)), nil)
	doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	doc.Paragraphs()[0].SetStyle("Heading1")
	doc.AddParagraph("extra").SetStyle("NonExistent")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	w := doc.Warnings()
	foundUnknown := false
	for _, msg := range w {
		if msg == "wordingo: unknown style \"NonExistent\" referenced by paragraph 3" {
			foundUnknown = true
		}
	}
	if !foundUnknown {
		t.Errorf("expected warning for unknown style, got: %v", w)
	}
}

func TestWarningNonExistentStyleValidXML(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("test").SetStyle("NonExistent")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", f.Name, err)
		}
		payload, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("read %s: %v", f.Name, err)
		}
		if f.Name == "word/document.xml" {
			if !strings.Contains(string(payload), "NonExistent") {
				t.Error("document.xml should contain the style reference even if unknown")
			}
		}
	}
}
