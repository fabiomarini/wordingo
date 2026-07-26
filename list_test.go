package wordingo

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestList_Ordered(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	lb := doc.AddList(true)
	lb.AddItem("First", 0).AddItem("Second", 0).AddItem("Third", 0)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// Verify paragraphs have NumPr with numId and ilvl
	if !strings.Contains(content, "numPr") {
		t.Error("document.xml missing numPr on list paragraphs")
	}
	if !strings.Contains(content, `ilvl w:val="0"`) {
		t.Error("document.xml missing ilvl val=0")
	}
	if !strings.Contains(content, `numId`) {
		t.Error("document.xml missing numId")
	}

	// Verify all 3 list items appear
	for _, item := range []string{"First", "Second", "Third"} {
		if !strings.Contains(content, item) {
			t.Errorf("document.xml missing list item %q", item)
		}
	}

	// Verify numbering.xml was created with abstractNum + num
	numContent := readZipEntryFromBuf(t, buf.Bytes(), "word/numbering.xml")
	if !strings.Contains(numContent, `<w:abstractNum`) {
		t.Error("numbering.xml missing abstractNum")
	}
	if !strings.Contains(numContent, `<w:num`) {
		t.Error("numbering.xml missing num")
	}
	if !strings.Contains(numContent, `decimal`) {
		t.Error("numbering.xml missing decimal numFmt")
	}
	if !strings.Contains(numContent, `%1.`) {
		t.Error("numbering.xml missing lvlText %1.")
	}

	// Verify 9-level depth — count opening <w:lvl tags (not </w:lvl>)
	if got := strings.Count(numContent, `<w:lvl `); got != 9 {
		t.Errorf("expected 9 lvl elements in numbering.xml, got %d", got)
	}

	// Verify doc rels has numbering relationship
	relsContent := readZipEntryFromBuf(t, buf.Bytes(), "word/_rels/document.xml.rels")
	if !strings.Contains(relsContent, relNumbering) {
		t.Error("document.xml.rels missing numbering relationship")
	}
}

func TestList_Bulleted(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	lb := doc.AddList(false)
	lb.AddItem("Bullet 1", 0).AddItem("Bullet 2", 0)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	numContent := readZipEntryFromBuf(t, buf.Bytes(), "word/numbering.xml")
	if !strings.Contains(numContent, `bullet`) {
		t.Error("numbering.xml missing bullet numFmt")
	}

	// Check bullet chars — lvl 0 uses \u2022 (bullet)
	if !strings.Contains(numContent, "\u2022") {
		t.Error("numbering.xml missing bullet character for level 0")
	}
}

func TestList_MultiLevel(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	lb := doc.AddList(true)
	lb.AddItem("Top", 0).AddItem("Nested", 1).AddItem("Deep nested", 2)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// Check ilvl values in document.xml — uses full element syntax
	if !strings.Contains(content, `ilvl w:val="0"`) {
		t.Error("missing ilvl 0")
	}
	if !strings.Contains(content, `ilvl w:val="1"`) {
		t.Error("missing ilvl 1")
	}
	if !strings.Contains(content, `ilvl w:val="2"`) {
		t.Error("missing ilvl 2")
	}
}

func TestList_AddListFromSlice(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	items := []string{"Apple", "Banana", "Cherry"}
	lb := doc.AddListFromSlice(items, true)
	if lb == nil {
		t.Fatal("AddListFromSlice returned nil")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// All 3 items should be present
	if !strings.Contains(content, "Apple") {
		t.Error("missing Apple text")
	}
	if !strings.Contains(content, "Banana") {
		t.Error("missing Banana text")
	}
	if !strings.Contains(content, "Cherry") {
		t.Error("missing Cherry text")
	}
}

func TestList_AddNumberingDef(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	lb := doc.AddNumberingDef("upperRoman", 1)
	lb.AddItem("Roman I", 0)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	numContent := readZipEntryFromBuf(t, buf.Bytes(), "word/numbering.xml")
	if !strings.Contains(numContent, `upperRoman`) {
		t.Error("numbering.xml missing upperRoman numFmt")
	}
}

func TestList_XEscapeHatch(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	lb := doc.AddList(true)
	lb.AddItem("Only item", 0)
	last := lb.X()
	if last == nil {
		t.Fatal("X() returned nil")
	}
	if len(last.R) == 0 || last.R[0].T == nil || last.R[0].T.Value != "Only item" {
		t.Errorf("X() last paragraph text = %v, want 'Only item'", last.R[0].T)
	}
}

func TestList_TemplateNumberingPreserved(t *testing.T) {
	// Simulate a template with existing numbering: abstractNumId=0, numId=1
	templateNumXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:abstractNum w:abstractNumId="0">
<w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/><w:start w:val="1"/></w:lvl>
</w:abstractNum>
<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>
</w:numbering>`

	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	// Add template numbering to the package
	doc.pkg.MarkModified("word/numbering.xml", []byte(templateNumXML))
	doc.pkg.ContentTypes.Overrides["/word/numbering.xml"] = ctNumbering

	// Ensure rels exist (using Package.Rels rather than &opc.Relationships{})
	doc.pkg.MarkModified("word/document.xml", []byte(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p/></w:body></w:document>`))
	doc.pkg.ContentTypes.Overrides["/word/numbering.xml"] = ctNumbering

	// Now add our list — should use abstractNumId=1, numId=2
	lb := doc.AddList(true)
	lb.AddItem("New list item", 0)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	numContent := readZipEntryFromBuf(t, buf.Bytes(), "word/numbering.xml")

	// Both old and new entries should be present
	if !strings.Contains(numContent, `abstractNumId="0"`) {
		t.Error("template abstractNumId=0 missing")
	}
	if !strings.Contains(numContent, `abstractNumId="1"`) {
		t.Error("auto-generated abstractNumId=1 missing")
	}
	if !strings.Contains(numContent, `numId="1"`) {
		t.Error("template numId=1 missing")
	}
	if !strings.Contains(numContent, `numId="2"`) {
		t.Error("auto-generated numId=2 missing")
	}
}

// ---------- Hyperlink tests ----------

func TestHyperlink_Basic(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("")
	run := p.AddHyperlink("click here", "https://example.com")
	if run == nil {
		t.Fatal("AddHyperlink returned nil")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	docContent := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// Verify CT_Hyperlink element
	if !strings.Contains(docContent, `<w:hyperlink`) {
		t.Error("document.xml missing w:hyperlink")
	}

	// Verify run text inside hyperlink
	if !strings.Contains(docContent, "click here") {
		t.Error("document.xml missing hyperlink display text")
	}

	// Verify relationship with TargetMode=External
	relsContent := readZipEntryFromBuf(t, buf.Bytes(), "word/_rels/document.xml.rels")
	if !strings.Contains(relsContent, `TargetMode="External"`) {
		t.Error("rels missing TargetMode=External")
	}
	if !strings.Contains(relsContent, "https://example.com") {
		t.Error("rels missing hyperlink target URI")
	}
	if !strings.Contains(relsContent, relHyperlink) {
		t.Error("rels missing hyperlink relationship type")
	}
}

func TestHyperlink_FormattingChaining(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("")
	p.AddHyperlink("link", "https://example.com").SetBold(true).SetColor("0563C1")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// Verify bold formatting on the hyperlink run
	if !strings.Contains(content, `<w:b`) {
		t.Error("missing bold formatting on hyperlink run")
	}
	if !strings.Contains(content, `0563C1`) {
		t.Error("missing color 0563C1 on hyperlink run")
	}
}

func TestHyperlink_MultipleCalls(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("")
	p.AddHyperlink("first", "https://a.com")
	p.AddHyperlink("second", "https://b.com")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	docsContent := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if got := strings.Count(docsContent, `<w:hyperlink`); got != 2 {
		t.Errorf("expected 2 hyperlink elements, got %d", got)
	}

	relsContent := readZipEntryFromBuf(t, buf.Bytes(), "word/_rels/document.xml.rels")
	// Each hyperlink gets its own rId — should have at least 2 hyperlink rels
	hyperlinkRelCount := strings.Count(relsContent, relHyperlink)
	if hyperlinkRelCount < 2 {
		t.Errorf("expected at least 2 hyperlink relationships, got %d", hyperlinkRelCount)
	}
}

// ---------- Helper ----------

func readZipEntryFromBuf(t *testing.T, data []byte, name string) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	for _, f := range zr.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open %s: %v", name, err)
			}
			defer rc.Close()
			var buf bytes.Buffer
			_, err = buf.ReadFrom(rc)
			if err != nil {
				t.Fatalf("read %s: %v", name, err)
			}
			return buf.String()
		}
	}
	t.Fatalf("part %s not found", name)
	return ""
}
