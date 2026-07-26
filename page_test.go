package wordingo

import (
	"bytes"
	"strings"
	"testing"
)

func TestPageSetup_SetOrientation_Landscape(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	// Initial: Letter portrait (W=12240, H=15840, W<H)
	doc.SetOrientation(OrientationLandscape)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// After landscape swap: W should be 15840, H should be 12240
	if !strings.Contains(content, `w:w="15840"`) {
		t.Error("document.xml missing landscape pgSz w=15840 (swapped)")
	}
	if !strings.Contains(content, `w:h="12240"`) {
		t.Error("document.xml missing landscape pgSz h=12240 (swapped)")
	}
}

func TestPageSetup_SetOrientation_LandscapeToPortrait(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	// Set landscape then back to portrait — should restore W<H
	doc.SetOrientation(OrientationLandscape)
	doc.SetOrientation(OrientationPortrait)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// After portrait restore: W=12240, H=15840
	if !strings.Contains(content, `w:w="12240"`) {
		t.Error("document.xml missing portrait pgSz w=12240")
	}
	if !strings.Contains(content, `w:h="15840"`) {
		t.Error("document.xml missing portrait pgSz h=15840")
	}
}

func TestPageSetup_SetPaperSize_A4(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	doc.SetPaperSize(PaperA4W, PaperA4H)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// A4: W=11906, H=16838
	if !strings.Contains(content, `w:w="11906"`) {
		t.Error("document.xml missing A4 pgSz w=11906")
	}
	if !strings.Contains(content, `w:h="16838"`) {
		t.Error("document.xml missing A4 pgSz h=16838")
	}
}

func TestPageSetup_SetPaperSize_Legal(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	doc.SetPaperSize(PaperLegalW, PaperLegalH)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// Legal: W=12240, H=20160
	if !strings.Contains(content, `w:w="12240"`) {
		t.Error("document.xml missing Legal pgSz w=12240")
	}
	if !strings.Contains(content, `w:h="20160"`) {
		t.Error("document.xml missing Legal pgSz h=20160")
	}
}

func TestPageSetup_SetMargins(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	// 2 inches = 2880 twips
	doc.SetMargins(2880, 1440, 2880, 1440)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	if !strings.Contains(content, `w:top="2880"`) {
		t.Error("document.xml missing pgMar top=2880")
	}
	if !strings.Contains(content, `w:right="1440"`) {
		t.Error("document.xml missing pgMar right=1440")
	}
	if !strings.Contains(content, `w:bottom="2880"`) {
		t.Error("document.xml missing pgMar bottom=2880")
	}
	if !strings.Contains(content, `w:left="1440"`) {
		t.Error("document.xml missing pgMar left=1440")
	}
}

func TestPageSetup_Chaining(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	// Set paper size BEFORE orientation to get A4 landscape
	doc.SetPaperSize(PaperA4W, PaperA4H)
	d := doc.SetOrientation(OrientationLandscape).SetMargins(1440, 1440, 1440, 1440)
	if d != doc {
		t.Error("chaining should return the same Document")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// A4 landscape: W=16838, H=11906 (swapped)
	if !strings.Contains(content, `w:w="16838"`) {
		t.Error("document.xml missing landscaped A4 pgSz w=16838")
	}
	if !strings.Contains(content, `w:h="11906"`) {
		t.Error("document.xml missing landscaped A4 pgSz h=11906")
	}
}

func TestPageSetup_AddPageBreak(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	doc.AddParagraph("First page")
	doc.AddPageBreak()
	doc.AddParagraph("Second page")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// Verify PageBreakBefore element exists
	if !strings.Contains(content, "w:pageBreakBefore") {
		t.Error("document.xml missing pageBreakBefore on page break paragraph")
	}

	// Verify all text present
	if !strings.Contains(content, "First page") {
		t.Error("document.xml missing 'First page' text")
	}
	if !strings.Contains(content, "Second page") {
		t.Error("document.xml missing 'Second page' text")
	}
}

func TestPageSetup_SetPageBreakBefore(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	p := doc.AddParagraph("This starts on new page")
	p.SetPageBreakBefore(true)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	if !strings.Contains(content, "w:pageBreakBefore") {
		t.Error("document.xml missing pageBreakBefore after SetPageBreakBefore(true)")
	}
}

func TestPageSetup_SetPageBreakBefore_Clear(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	p := doc.AddParagraph("No break")
	p.SetPageBreakBefore(true)
	p.SetPageBreakBefore(false)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	if strings.Contains(content, "w:pageBreakBefore") {
		t.Error("document.xml should NOT have pageBreakBefore after SetPageBreakBefore(false)")
	}
}

func TestPageSetup_SectionWrapper(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	s := doc.Section()
	if s == nil {
		t.Fatal("Section() returned nil")
	}

	// Set paper size BEFORE orientation to get Legal landscape
	s.SetPaperSize(PaperLegalW, PaperLegalH)
	s.SetOrientation(OrientationLandscape)
	s.SetMargins(720, 720, 720, 720)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// Legal landscape: W=20160, H=12240 (swapped)
	if !strings.Contains(content, `w:w="20160"`) {
		t.Error("document.xml missing Section-set pgSz w=20160")
	}
	if !strings.Contains(content, `w:h="12240"`) {
		t.Error("document.xml missing Section-set pgSz h=12240")
	}
	if !strings.Contains(content, `w:top="720"`) {
		t.Error("document.xml missing Section-set pgMar top=720")
	}
}

func TestPageSetup_SectionXEscapeHatch(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	s := doc.Section()
	ct := s.X()
	if ct == nil {
		t.Fatal("Section.X() returned nil")
	}
}
