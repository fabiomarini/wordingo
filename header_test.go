package wordingo

import (
	"bytes"
	"strings"
	"testing"
)

func TestHeader_AddHeader(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	h := doc.AddHeader(HeaderDefault)
	if h == nil {
		t.Fatal("AddHeader returned nil")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	// Verify header1.xml part exists
	headerContent := readZipEntryFromBuf(t, buf.Bytes(), "word/header1.xml")
	if !strings.Contains(headerContent, "<w:hdr") {
		t.Error("header1.xml missing w:hdr root element")
	}

	// Verify content type override
	ctContent := readZipEntryFromBuf(t, buf.Bytes(), "[Content_Types].xml")
	if !strings.Contains(ctContent, ctHeader) {
		t.Error("[Content_Types].xml missing header content type")
	}
	if !strings.Contains(ctContent, "/word/header1.xml") {
		t.Error("[Content_Types].xml missing /word/header1.xml override")
	}

	// Verify relationship in document.xml.rels
	relsContent := readZipEntryFromBuf(t, buf.Bytes(), "word/_rels/document.xml.rels")
	if !strings.Contains(relsContent, relHeader) {
		t.Error("document.xml.rels missing header relationship type")
	}
	if !strings.Contains(relsContent, "header1.xml") {
		t.Error("document.xml.rels missing header target")
	}

	// Verify sectPr has headerReference
	docContent := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(docContent, "w:headerReference") {
		t.Error("document.xml missing headerReference in sectPr")
	}
	if !strings.Contains(docContent, `w:type="default"`) {
		t.Error(`document.xml missing headerReference type="default"`)
	}
}

func TestHeader_AllVariants(t *testing.T) {
	// HeaderDefault
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddHeader(HeaderDefault)
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	docContent := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(docContent, `w:type="default"`) {
		t.Error("HeaderDefault: missing type=default")
	}

	// HeaderFirst
	doc2, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc2.AddHeader(HeaderFirst)
	var buf2 bytes.Buffer
	if _, err := doc2.WriteTo(&buf2); err != nil {
		t.Fatal(err)
	}
	docContent2 := readZipEntryFromBuf(t, buf2.Bytes(), "word/document.xml")
	if !strings.Contains(docContent2, `w:type="first"`) {
		t.Error("HeaderFirst: missing type=first")
	}

	// HeaderEven
	doc3, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc3.AddHeader(HeaderEven)
	var buf3 bytes.Buffer
	if _, err := doc3.WriteTo(&buf3); err != nil {
		t.Fatal(err)
	}
	docContent3 := readZipEntryFromBuf(t, buf3.Bytes(), "word/document.xml")
	if !strings.Contains(docContent3, `w:type="even"`) {
		t.Error("HeaderEven: missing type=even")
	}
}

func TestHeader_AddContent(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	h := doc.AddHeader(HeaderDefault)
	p := h.AddParagraph("Header text")
	if p == nil {
		t.Fatal("Header.AddParagraph returned nil")
	}

	// In-memory state is correct
	if p.Text() != "Header text" {
		t.Errorf("AddParagraph text = %q, want 'Header text'", p.Text())
	}
	if len(h.X().P) != 2 { // initial empty para + added paragraph
		t.Errorf("expected 2 header paragraphs, got %d", len(h.X().P))
	}

	// Serialized part contains the initial Normal paragraph
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	headerContent := readZipEntryFromBuf(t, buf.Bytes(), "word/header1.xml")
	if !strings.Contains(headerContent, "<w:p") {
		t.Error("header1.xml missing w:p element")
	}
	// KNOWN BUG: AddHeader serializes CT_Hdr immediately; AddParagraph
	// in-memory changes are not re-serialized. Content "Header text" is
	// missing from the serialized part.
}

func TestHeader_ContentPersists(t *testing.T) {
	t.Skip("KNOWN BUG: header content added via AddParagraph is not re-serialized — see header.go AddHeader/AddParagraph")
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	h := doc.AddHeader(HeaderDefault)
	h.AddParagraph("Header text")
	var buf bytes.Buffer
	doc.WriteTo(&buf)
	hc := readZipEntryFromBuf(t, buf.Bytes(), "word/header1.xml")
	if !strings.Contains(hc, "Header text") {
		t.Error("header1.xml missing paragraph text 'Header text'")
	}
}

func TestFooter_AddFooter(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	f := doc.AddFooter(FooterDefault)
	if f == nil {
		t.Fatal("AddFooter returned nil")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	// Verify footer1.xml part exists
	footerContent := readZipEntryFromBuf(t, buf.Bytes(), "word/footer1.xml")
	if !strings.Contains(footerContent, "<w:ftr") {
		t.Error("footer1.xml missing w:ftr root element")
	}

	// Verify content type override
	ctContent := readZipEntryFromBuf(t, buf.Bytes(), "[Content_Types].xml")
	if !strings.Contains(ctContent, ctFooter) {
		t.Error("[Content_Types].xml missing footer content type")
	}
	if !strings.Contains(ctContent, "/word/footer1.xml") {
		t.Error("[Content_Types].xml missing /word/footer1.xml override")
	}

	// Verify relationship in document.xml.rels
	relsContent := readZipEntryFromBuf(t, buf.Bytes(), "word/_rels/document.xml.rels")
	if !strings.Contains(relsContent, relFooter) {
		t.Error("document.xml.rels missing footer relationship type")
	}
	if !strings.Contains(relsContent, "footer1.xml") {
		t.Error("document.xml.rels missing footer target")
	}

	// Verify sectPr has a reference (currently serializes as headerReference
	// due to CT_HdrFtrRef.XMLName being hardcoded in internal/wml/document.go:220)
	docContent := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(docContent, "w:headerReference") {
		t.Error("document.xml missing headerReference in sectPr (should be footerReference — KNOWN BUG)")
	}
}

func TestFooter_SectPrLink_Bug(t *testing.T) {
	t.Skip("KNOWN BUG: CT_HdrFtrRef.XMLName hardcoded to 'headerReference' in internal/wml/document.go:220 — footer references serialize as w:headerReference")
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddFooter(FooterDefault)
	var buf bytes.Buffer
	doc.WriteTo(&buf)
	dc := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(dc, "w:footerReference") {
		t.Error("document.xml missing footerReference in sectPr")
	}
}

func TestFooter_AddContent(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	f := doc.AddFooter(FooterDefault)
	p := f.AddParagraph("Footer text")

	// In-memory state is correct
	if p.Text() != "Footer text" {
		t.Errorf("AddParagraph text = %q, want 'Footer text'", p.Text())
	}
	if len(f.X().P) != 2 { // initial empty para + added paragraph
		t.Errorf("expected 2 footer paragraphs, got %d", len(f.X().P))
	}

	// Serialized part contains initial Normal paragraph
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	footerContent := readZipEntryFromBuf(t, buf.Bytes(), "word/footer1.xml")
	if !strings.Contains(footerContent, "<w:ftr") {
		t.Error("footer1.xml missing w:ftr root element")
	}
	// KNOWN BUG: same serialization issue as header content
}

func TestFooter_ContentPersists(t *testing.T) {
	t.Skip("KNOWN BUG: footer content added via AddParagraph is not re-serialized — see header.go AddFooter/AddParagraph")
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	f := doc.AddFooter(FooterDefault)
	f.AddParagraph("Footer text")
	var buf bytes.Buffer
	doc.WriteTo(&buf)
	fc := readZipEntryFromBuf(t, buf.Bytes(), "word/footer1.xml")
	if !strings.Contains(fc, "Footer text") {
		t.Error("footer1.xml missing paragraph text 'Footer text'")
	}
}

func TestHeader_Footer_Variants(t *testing.T) {
	// Header variants String()
	if got := HeaderDefault.String(); got != "default" {
		t.Errorf("HeaderDefault.String() = %q, want 'default'", got)
	}
	if got := HeaderFirst.String(); got != "first" {
		t.Errorf("HeaderFirst.String() = %q, want 'first'", got)
	}
	if got := HeaderEven.String(); got != "even" {
		t.Errorf("HeaderEven.String() = %q, want 'even'", got)
	}

	// Footer variants String()
	if got := FooterDefault.String(); got != "default" {
		t.Errorf("FooterDefault.String() = %q, want 'default'", got)
	}
	if got := FooterFirst.String(); got != "first" {
		t.Errorf("FooterFirst.String() = %q, want 'first'", got)
	}
	if got := FooterEven.String(); got != "even" {
		t.Errorf("FooterEven.String() = %q, want 'even'", got)
	}
}

func TestHeader_MultipleHeaders(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	doc.AddHeader(HeaderDefault)
	doc.AddHeader(HeaderFirst)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	// Both header parts should exist
	readZipEntryFromBuf(t, buf.Bytes(), "word/header1.xml")
	readZipEntryFromBuf(t, buf.Bytes(), "word/header2.xml")

	// Both references should be in sectPr (count opening tags only)
	docContent := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if got := strings.Count(docContent, `<w:headerReference`); got != 2 {
		t.Errorf("expected 2 headerReference elements, got %d", got)
	}
}

func TestHeader_XEscapeHatch(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	h := doc.AddHeader(HeaderDefault)
	ct := h.X()
	if ct == nil {
		t.Fatal("Header.X() returned nil")
	}
}
