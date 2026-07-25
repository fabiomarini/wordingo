package wordingo

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// diffParts compares two unzipped part sets per part (D-04 pattern).
// Manifests ([Content_Types].xml and .rels parts) are excluded from
// byte comparison; unmodified parts are expected to be byte-identical.
func diffParts(t *testing.T, a, b []byte) {
	t.Helper()
	manifest := func(name string) bool {
		return name == "[Content_Types].xml" || strings.Contains(name, "/_rels/") || name == "_rels/.rels"
	}
	partsOf := func(data []byte) map[string][]byte {
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			t.Fatal(err)
		}
		m := make(map[string][]byte, len(zr.File))
		for _, f := range zr.File {
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			payload, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatal(err)
			}
			m[f.Name] = payload
		}
		return m
	}
	pa, pb := partsOf(a), partsOf(b)
	for name, payload := range pa {
		if manifest(name) {
			continue
		}
		other, ok := pb[name]
		if !ok {
			t.Errorf("part %s missing from saved package", name)
			continue
		}
		if !bytes.Equal(payload, other) {
			t.Errorf("part %s payload differs (%d vs %d bytes)", name, len(payload), len(other))
		}
	}
	for name := range pb {
		if _, ok := pa[name]; !ok {
			t.Errorf("part %s unexpected in saved package", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Fixture generation helpers
// ---------------------------------------------------------------------------

// fixtureDir is the on-disk location for real-producer fixtures (D-08).
var fixtureDir = filepath.Join("testdata", "roundtrip")

// addZipEntry writes one entry to a deterministically-timed ZIP writer.
func addZipEntry(t *testing.T, zw *zip.Writer, name string, payload []byte) {
	t.Helper()
	h := &zip.FileHeader{Name: name, Method: zip.Deflate}
	h.Modified = time.Date(2020, 3, 4, 5, 6, 7, 0, time.UTC)
	w, err := zw.CreateHeader(h)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(payload); err != nil {
		t.Fatal(err)
	}
}

// synthFixture declares the shared constant parts for synthetic fixtures.
const (
	synthContentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`

	synthRootRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	synthDocRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

	synthStyles = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:style w:type="paragraph" w:styleId="Normal"/>
<w:style w:type="paragraph" w:styleId="Heading1"/>
<w:style w:type="paragraph" w:styleId="Heading2"/>
</w:styles>`
)

// buildSyntheticFixture assembles a complete .docx from the given
// document body XML payload and optional extra parts.
func buildSyntheticFixture(t *testing.T, docBodyXML string, extras map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	addZipEntry(t, zw, "[Content_Types].xml", []byte(synthContentTypes))
	addZipEntry(t, zw, "_rels/.rels", []byte(synthRootRels))
	addZipEntry(t, zw, "word/document.xml", []byte(docBodyXML))
	addZipEntry(t, zw, "word/_rels/document.xml.rels", []byte(synthDocRels))
	addZipEntry(t, zw, "word/styles.xml", []byte(synthStyles))

	for name, payload := range extras {
		addZipEntry(t, zw, name, payload)
	}

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// encodeDocumentXML marshals a *wml.CT_Document to canonical XML bytes.
func encodeDocumentXML(t *testing.T, doc *wml.CT_Document) []byte {
	t.Helper()
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(doc); err != nil {
		t.Fatal(err)
	}
	if err := enc.Flush(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// fixtureSingleParaXML returns a document.xml body containing one
// paragraph with text "Hello World".
func fixtureSingleParaXML(t *testing.T) []byte {
	t.Helper()
	hello := "Hello World"
	doc := &wml.CT_Document{
		Body: &wml.CT_Body{
			P: []*wml.CT_P{
				{
					R: []*wml.CT_R{
						{T: &wml.CT_Text{Value: hello}},
					},
				},
			},
			SectPr: defaultSectPr(),
		},
	}
	return encodeDocumentXML(t, doc)
}

// fixtureMultiHeadingXML returns a document.xml body with three
// paragraphs: Normal, Heading1, Heading2.
func fixtureMultiHeadingXML(t *testing.T) []byte {
	t.Helper()
	valNormal := "Normal"
	valH1 := "Heading1"
	valH2 := "Heading2"
	text1 := "First paragraph"
	text2 := "Second heading"
	text3 := "Third heading"
	doc := &wml.CT_Document{
		Body: &wml.CT_Body{
			P: []*wml.CT_P{
				{
					PPr: &wml.CT_PPr{PStyle: &wml.CT_PStyle{Val: &valNormal}},
					R:   []*wml.CT_R{{T: &wml.CT_Text{Value: text1}}},
				},
				{
					PPr: &wml.CT_PPr{PStyle: &wml.CT_PStyle{Val: &valH1}},
					R:   []*wml.CT_R{{T: &wml.CT_Text{Value: text2}}},
				},
				{
					PPr: &wml.CT_PPr{PStyle: &wml.CT_PStyle{Val: &valH2}},
					R:   []*wml.CT_R{{T: &wml.CT_Text{Value: text3}}},
				},
			},
			SectPr: defaultSectPr(),
		},
	}
	return encodeDocumentXML(t, doc)
}

// generateFixtures writes all on-disk fixture files to testdata/roundtrip/.
func generateFixtures(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(fixtureDir, 0755); err != nil {
		t.Fatal(err)
	}

	// blank.docx — use Create()
	genFile(t, "blank.docx", func() []byte {
		doc, err := Create()
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if _, err := doc.WriteTo(&buf); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	})

	// single-paragraph.docx
	genFile(t, "single-paragraph.docx", func() []byte {
		return buildSyntheticFixture(t, string(fixtureSingleParaXML(t)), nil)
	})

	// multi-heading.docx
	genFile(t, "multi-heading.docx", func() []byte {
		return buildSyntheticFixture(t, string(fixtureMultiHeadingXML(t)), nil)
	})

	// header-footer.docx — includes header1.xml and footer1.xml parts
	genFile(t, "header-footer.docx", func() []byte {
		hfContentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
<Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>
</Types>`

		hfDocRels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header1.xml"/>
</Relationships>`

		docXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
            xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<w:body>
<w:p><w:r><w:t>Header doc</w:t></w:r></w:p>
<w:sectPr>
<w:headerReference w:type="default" r:id="rId2"/>
<w:pgSz w:w="12240" w:h="15840"/>
<w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440"/>
</w:sectPr>
</w:body>
</w:document>`

		headerXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:hdr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:p><w:r><w:t>Header</w:t></w:r></w:p>
</w:hdr>`

		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)

		addZipEntry(t, zw, "[Content_Types].xml", []byte(hfContentTypes))
		addZipEntry(t, zw, "_rels/.rels", []byte(synthRootRels))
		addZipEntry(t, zw, "word/document.xml", []byte(docXML))
		addZipEntry(t, zw, "word/_rels/document.xml.rels", []byte(hfDocRels))
		addZipEntry(t, zw, "word/styles.xml", []byte(synthStyles))
		addZipEntry(t, zw, "word/header1.xml", []byte(headerXML))

		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	})
}

// genFile writes a fixture to disk if it does not already exist.
func genFile(t *testing.T, name string, fn func() []byte) {
	t.Helper()
	path := filepath.Join(fixtureDir, name)
	data := fn()
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// ---------------------------------------------------------------------------
// Test: fixture generation
// ---------------------------------------------------------------------------

func TestGenerateFixtures(t *testing.T) {
	generateFixtures(t)
}

// ---------------------------------------------------------------------------
// Round-trip tests — per-part byte diff
// ---------------------------------------------------------------------------

func TestRoundTrip_Blank(t *testing.T) {
	generateFixtures(t)
	data, err := os.ReadFile(filepath.Join(fixtureDir, "blank.docx"))
	if err != nil {
		t.Fatal(err)
	}

	doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	// Exercise Document.Paragraphs()
	paras := doc.Paragraphs()
	// Blank doc has exactly 1 paragraph (Normal style)
	if len(paras) != 1 {
		t.Errorf("expected 1 paragraph, got %d", len(paras))
	}
	if len(paras) > 0 && paras[0].Style() != "Normal" {
		t.Errorf("expected style Normal, got %q", paras[0].Style())
	}

	// Save and per-part diff
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	diffParts(t, data, buf.Bytes())
}

func TestRoundTrip_SingleParagraph(t *testing.T) {
	generateFixtures(t)
	data, err := os.ReadFile(filepath.Join(fixtureDir, "single-paragraph.docx"))
	if err != nil {
		t.Fatal(err)
	}

	doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	paras := doc.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("expected 1 paragraph, got %d", len(paras))
	}
	if paras[0].Text() != "Hello World" {
		t.Errorf("expected text %q, got %q", "Hello World", paras[0].Text())
	}
	if paras[0].Style() != "" {
		t.Errorf("expected empty style, got %q", paras[0].Style())
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	diffParts(t, data, buf.Bytes())
}

func TestRoundTrip_MultiHeading(t *testing.T) {
	generateFixtures(t)
	data, err := os.ReadFile(filepath.Join(fixtureDir, "multi-heading.docx"))
	if err != nil {
		t.Fatal(err)
	}

	doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	paras := doc.Paragraphs()
	if len(paras) != 3 {
		t.Fatalf("expected 3 paragraphs, got %d", len(paras))
	}

	styleWant := []string{"Normal", "Heading1", "Heading2"}
	textWant := []string{"First paragraph", "Second heading", "Third heading"}
	for i, p := range paras {
		if p.Style() != styleWant[i] {
			t.Errorf("para[%d] Style = %q, want %q", i, p.Style(), styleWant[i])
		}
		if p.Text() != textWant[i] {
			t.Errorf("para[%d] Text = %q, want %q", i, p.Text(), textWant[i])
		}
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	diffParts(t, data, buf.Bytes())
}

func TestRoundTrip_HeaderFooter(t *testing.T) {
	generateFixtures(t)
	data, err := os.ReadFile(filepath.Join(fixtureDir, "header-footer.docx"))
	if err != nil {
		t.Fatal(err)
	}

	doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	// Body eagerly parsed
	paras := doc.Paragraphs()
	if len(paras) != 1 {
		t.Errorf("expected 1 paragraph, got %d", len(paras))
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	diffParts(t, data, buf.Bytes())
}

// ---------------------------------------------------------------------------
// Lazy loading test
// ---------------------------------------------------------------------------

func TestLazyLoading(t *testing.T) {
	generateFixtures(t)
	data, err := os.ReadFile(filepath.Join(fixtureDir, "header-footer.docx"))
	if err != nil {
		t.Fatal(err)
	}

	doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	// Document body parsed eagerly — CT_Document non-nil
	if doc.doc == nil || doc.doc.Body == nil {
		t.Fatal("body should be eagerly parsed")
	}

	// header part should exist but NOT be marked modified
	headerPart, ok := doc.pkg.Parts["word/header1.xml"]
	if !ok {
		t.Fatal("expected header1.xml part in fixture")
	}
	if headerPart.IsModified() {
		t.Error("header part should not be marked modified")
	}

	// Ensure supporting style parts are not modified either
	for _, name := range []string{"word/styles.xml"} {
		part, ok := doc.pkg.Parts[name]
		if !ok {
			continue
		}
		if part.IsModified() {
			t.Errorf("supporting part %s should not be marked modified", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Synthetic inline round-trip tests (no on-disk files needed)
// ---------------------------------------------------------------------------

func TestRoundTrip_Synthetic_Blank(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	paras := doc.Paragraphs()
	if len(paras) != 1 {
		t.Errorf("expected 1 paragraph, got %d", len(paras))
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	// Re-open and verify round-trip
	doc2, err := OpenReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc2.Paragraphs()) != 1 {
		t.Errorf("re-opened: expected 1 paragraph, got %d", len(doc2.Paragraphs()))
	}
}

// ---------------------------------------------------------------------------
// Additional tests
// ---------------------------------------------------------------------------

func TestParagraphs_NilBody(t *testing.T) {
	// Document with nil body should not panic
	doc := &Document{}
	paras := doc.Paragraphs()
	if paras != nil {
		t.Error("expected nil paragraphs for nil body document")
	}
}

func TestOpen_MissingDocumentXML(t *testing.T) {
	t.Skip("opc.Open requires a valid package; testing at Document level is covered by acceptance")
}

func TestParagraph_StyleNilChain(t *testing.T) {
	p := &Paragraph{ct: &wml.CT_P{}}
	if s := p.Style(); s != "" {
		t.Errorf("expected empty style, got %q", s)
	}
}

func TestParagraph_TextEmptyRuns(t *testing.T) {
	p := &Paragraph{ct: &wml.CT_P{R: []*wml.CT_R{{}, {T: &wml.CT_Text{Value: "ok"}}}}}
	if s := p.Text(); s != "ok" {
		t.Errorf("expected 'ok', got %q", s)
	}
}

func TestParagraph_X(t *testing.T) {
	ct := &wml.CT_P{}
	p := &Paragraph{ct: ct}
	if p.X() != ct {
		t.Error("X() should return underlying CT_P")
	}
}

func TestClose(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Close(); err != nil {
		t.Errorf("Close() error: %v", err)
	}
	if doc.pkg != nil {
		t.Error("pkg should be nil after Close")
	}
	if doc.doc != nil {
		t.Error("doc should be nil after Close")
	}
}

func TestCloseThenParagraphs(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.Close()
	// Paragraphs should return nil without panic
	if paras := doc.Paragraphs(); paras != nil {
		t.Error("expected nil paragraphs after Close")
	}
}

func TestSave_File(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "save-test.docx")
	if err := doc.Save(path); err != nil {
		t.Fatalf("Save(path) error: %v", err)
	}

	// Verify file exists and is valid
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 100 {
		t.Errorf("saved file too small: %d bytes", len(data))
	}

	// Re-open via opc.Open
	pkg, err := opc.Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("opc.Open after Save: %v", err)
	}
	if len(pkg.Warnings()) != 0 {
		t.Errorf("unexpected warnings: %v", pkg.Warnings())
	}
}

func TestRoundTrip_SyntheticSinglePara(t *testing.T) {
	data := buildSyntheticFixture(t, string(fixtureSingleParaXML(t)), nil)
	doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Paragraphs()) != 1 {
		t.Fatalf("expected 1 paragraph, got %d", len(doc.Paragraphs()))
	}
	if doc.Paragraphs()[0].Text() != "Hello World" {
		t.Errorf("text = %q, want %q", doc.Paragraphs()[0].Text(), "Hello World")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	diffParts(t, data, buf.Bytes())
}

func TestRoundTrip_SyntheticMultiHeading(t *testing.T) {
	data := buildSyntheticFixture(t, string(fixtureMultiHeadingXML(t)), nil)
	doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	paras := doc.Paragraphs()
	if len(paras) != 3 {
		t.Fatalf("expected 3 paragraphs, got %d", len(paras))
	}

	want := []struct{ style, text string }{
		{"Normal", "First paragraph"},
		{"Heading1", "Second heading"},
		{"Heading2", "Third heading"},
	}
	for i, w := range want {
		if paras[i].Style() != w.style {
			t.Errorf("para[%d] Style = %q, want %q", i, paras[i].Style(), w.style)
		}
		if paras[i].Text() != w.text {
			t.Errorf("para[%d] Text = %q, want %q", i, paras[i].Text(), w.text)
		}
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	diffParts(t, data, buf.Bytes())
}

func TestRoundTrip_SyntheticHeaderFooter(t *testing.T) {
	// Verify that header-footer fixture is syntactically round-trippable
	generateFixtures(t)
	data, err := os.ReadFile(filepath.Join(fixtureDir, "header-footer.docx"))
	if err != nil {
		t.Fatal(err)
	}

	doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	_ = doc.Paragraphs() // exercise eager parse

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	diffParts(t, data, buf.Bytes())
}

// TestRoundTrip_AllFixtures runs on-disk round-trip tests for all
// fixtures in testdata/roundtrip/.
func TestRoundTrip_AllFixtures(t *testing.T) {
	generateFixtures(t)
	matches, err := filepath.Glob(filepath.Join(fixtureDir, "*.docx"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatal("no fixtures found in testdata/roundtrip/")
	}
	for _, path := range matches {
		path := path
		t.Run(strings.TrimSuffix(filepath.Base(path), ".docx"), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
			if err != nil {
				t.Fatal(err)
			}
			_ = doc.Paragraphs() // exercise eager parse

			var buf bytes.Buffer
			if _, err := doc.WriteTo(&buf); err != nil {
				t.Fatal(err)
			}
			diffParts(t, data, buf.Bytes())
		})
	}
}
