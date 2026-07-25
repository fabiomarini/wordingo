package wordingo

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// ---------------------------------------------------------------------------
// Fixture builder helpers
// ---------------------------------------------------------------------------

// buildTemplateFixture builds a synthetic .docx template in memory.
// If bodyParagraphs > 0, the document.xml will contain that many
// paragraphs with distinguishable text. Style parts come from the
// default embedded assets (defaultStyles, etc.).
func buildTemplateFixture(t *testing.T, bodyParagraphs int, withHdrFtr bool) []byte {
	t.Helper()
	pkg := newBlankPackage()

	if bodyParagraphs > 0 || withHdrFtr {
		doc := &wml.CT_Document{
			Body: &wml.CT_Body{
				SectPr: defaultSectPr(),
			},
		}

		if bodyParagraphs > 0 {
			doc.Body.P = make([]*wml.CT_P, bodyParagraphs)
			for i := 0; i < bodyParagraphs; i++ {
				text := fmt.Sprintf("Paragraph %d text content", i+1)
				doc.Body.P[i] = &wml.CT_P{
					R: []*wml.CT_R{
						{T: &wml.CT_Text{Value: text}},
					},
				}
			}
		}

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
		pkg.MarkModified("word/document.xml", buf.Bytes())
	}

	var out bytes.Buffer
	if err := pkg.Save(&out); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

// partPayloads returns all part payloads from a .docx byte slice,
// keyed by part name.
func partPayloads(t *testing.T, data []byte) map[string][]byte {
	t.Helper()
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

// isManifest returns true for ZIP entries that are package manifests
// (content types or relationship parts).
func isManifest(name string) bool {
	return name == "[Content_Types].xml" || strings.HasSuffix(name, ".rels")
}

// assertPartsMatchExcept checks that all non-manifest parts in data
// have byte-identical counterparts in other, except for the
// specified part names.
func assertPartsMatchExcept(t *testing.T, data, other []byte, except ...string) {
	t.Helper()
	exceptSet := make(map[string]bool, len(except))
	for _, e := range except {
		exceptSet[e] = true
	}

	pa := partPayloads(t, data)
	pb := partPayloads(t, other)

	for name, payload := range pa {
		if isManifest(name) || exceptSet[name] {
			continue
		}
		otherPayload, ok := pb[name]
		if !ok {
			t.Errorf("part %q missing from output", name)
			continue
		}
		if !bytes.Equal(payload, otherPayload) {
			t.Errorf("part %q differs between input and output (%d vs %d bytes)",
				name, len(payload), len(otherPayload))
		}
	}
	for name := range pb {
		if _, ok := pa[name]; !ok && !isManifest(name) {
			t.Errorf("part %q unexpected in output", name)
		}
	}
}

var templateStylePartNames = []string{
	"word/styles.xml",
	"word/numbering.xml",
	"word/fontTable.xml",
	"word/theme/theme1.xml",
	"word/settings.xml",
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestFromTemplate_EmptyBody(t *testing.T) {
	fixture := buildTemplateFixture(t, 3, false)

	doc, err := FromTemplateReader(bytes.NewReader(fixture), int64(len(fixture)))
	if err != nil {
		t.Fatalf("FromTemplateReader: %v", err)
	}

	// Body must be empty (CREATE-03)
	if paras := doc.Paragraphs(); len(paras) != 0 {
		t.Errorf("FromTemplate: got %d paragraphs, want 0", len(paras))
	}

	// Style parts that exist in the source must be present in output.
	// numbering.xml is absent from blank-doc source so CloneStyles skips it.
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	savedPayloads := partPayloads(t, buf.Bytes())
	for _, name := range templateStylePartNames {
		if name == "word/numbering.xml" {
			continue // not in blank-doc source
		}
		if _, ok := savedPayloads[name]; !ok {
			t.Errorf("style part %q missing from FromTemplate output", name)
		}
	}

	// Re-open and verify style parts present
	reopened, err := OpenReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("re-open FromTemplate output: %v", err)
	}
	if len(reopened.Paragraphs()) != 0 {
		t.Error("re-opened FromTemplate document should have empty body")
	}
}

func TestFromTemplate_RoundTrip(t *testing.T) {
	fixture := buildTemplateFixture(t, 3, false)

	doc, err := FromTemplateReader(bytes.NewReader(fixture), int64(len(fixture)))
	if err != nil {
		t.Fatalf("FromTemplateReader: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	// Style parts should be byte-identical after round-trip (STYLE-ROUNDTRIP-01).
	// Body (word/document.xml) will differ — that's the point of FromTemplate.
	// webSettings.xml exists in original (blank-doc default) but not in output
	// (CloneStyles only copies the 5 style parts).
	assertPartsMatchExcept(t, fixture, buf.Bytes(), "word/document.xml", "word/webSettings.xml")
}

func TestOpenTemplate_BodyPreserved(t *testing.T) {
	fixture := buildTemplateFixture(t, 3, false)

	doc, err := OpenTemplateReader(bytes.NewReader(fixture), int64(len(fixture)))
	if err != nil {
		t.Fatalf("OpenTemplateReader: %v", err)
	}

	// Body must preserve original paragraphs (CREATE-04)
	paras := doc.Paragraphs()
	if len(paras) != 3 {
		t.Fatalf("OpenTemplate: got %d paragraphs, want 3", len(paras))
	}

	// Verify paragraph text content
	expectedTexts := []string{
		"Paragraph 1 text content",
		"Paragraph 2 text content",
		"Paragraph 3 text content",
	}
	for i, p := range paras {
		if p.Text() != expectedTexts[i] {
			t.Errorf("para[%d] Text = %q, want %q", i, p.Text(), expectedTexts[i])
		}
	}

	// Verify no header/footer references in sectPr (Pitfall 4)
	if doc.doc.Body.SectPr != nil {
		if len(doc.doc.Body.SectPr.HdrFtrRef) > 0 {
			t.Error("OpenTemplate body has header references in sectPr — Pitfall 4 violation")
		}
		if len(doc.doc.Body.SectPr.FtrRef) > 0 {
			t.Error("OpenTemplate body has footer references in sectPr — Pitfall 4 violation")
		}
	}
}

func TestOpenTemplate_RoundTrip(t *testing.T) {
	fixture := buildTemplateFixture(t, 3, false)

	doc, err := OpenTemplateReader(bytes.NewReader(fixture), int64(len(fixture)))
	if err != nil {
		t.Fatalf("OpenTemplateReader: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	// Style parts should be byte-identical after round-trip.
	// Body will differ because it's re-encoded with fresh sectPr.
	// webSettings.xml exists in original (blank-doc default) but not in output.
	assertPartsMatchExcept(t, fixture, buf.Bytes(), "word/document.xml", "word/webSettings.xml")
}

func TestFromTemplate_And_OpenTemplate_ReaderVariants(t *testing.T) {
	fixture := buildTemplateFixture(t, 2, false)
	dir := t.TempDir()

	// Write fixture to temp file
	tmplPath := filepath.Join(dir, "template.docx")
	if err := os.WriteFile(tmplPath, fixture, 0644); err != nil {
		t.Fatal(err)
	}

	// FromTemplate(path) — path-based convenience
	fromTmpl, err := FromTemplate(tmplPath)
	if err != nil {
		t.Fatalf("FromTemplate(path): %v", err)
	}
	if len(fromTmpl.Paragraphs()) != 0 {
		t.Error("FromTemplate(path) should return empty body")
	}

	// OpenTemplate(path) — path-based convenience
	openTmpl, err := OpenTemplate(tmplPath)
	if err != nil {
		t.Fatalf("OpenTemplate(path): %v", err)
	}
	if len(openTmpl.Paragraphs()) != 2 {
		t.Errorf("OpenTemplate(path): got %d paragraphs, want 2", len(openTmpl.Paragraphs()))
	}
}

func TestTemplateFromBlankDoc(t *testing.T) {
	// Create a blank document as the "template"
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	var blankBuf bytes.Buffer
	if _, err := doc.WriteTo(&blankBuf); err != nil {
		t.Fatal(err)
	}

	// FromTemplate on blank doc — should work (CloneStyles skips absent parts)
	tmpl, err := FromTemplateReader(bytes.NewReader(blankBuf.Bytes()), int64(blankBuf.Len()))
	if err != nil {
		t.Fatalf("FromTemplateReader on blank doc: %v", err)
	}
	if len(tmpl.Paragraphs()) != 0 {
		t.Error("FromTemplate on blank doc should clear body")
	}
}

func TestBuildEmptyBodyXML_Valid(t *testing.T) {
	xml := buildEmptyBodyXML()

	// Verify it's valid by wrapping in a minimal package and opening
	pkgBytes := mockPackageAround(t, xml)
	doc, err := OpenReader(bytes.NewReader(pkgBytes), int64(len(pkgBytes)))
	if err != nil {
		t.Fatalf("OpenReader on buildEmptyBodyXML package: %v", err)
	}
	if doc.doc == nil || doc.doc.Body == nil {
		t.Fatal("buildEmptyBodyXML: body is nil")
	}
	// Verify no paragraphs
	if len(doc.doc.Body.P) != 0 {
		t.Errorf("buildEmptyBodyXML: got %d paragraphs, want 0", len(doc.doc.Body.P))
	}
	// Verify sectPr present and correct
	if doc.doc.Body.SectPr == nil {
		t.Error("buildEmptyBodyXML: sectPr is nil")
	} else {
		if doc.doc.Body.SectPr.PgSz == nil {
			t.Error("buildEmptyBodyXML: PgSz is nil")
		} else {
			if doc.doc.Body.SectPr.PgSz.W == nil || *doc.doc.Body.SectPr.PgSz.W != 12240 {
				t.Error("buildEmptyBodyXML: pgSz W != 12240")
			}
			if doc.doc.Body.SectPr.PgSz.H == nil || *doc.doc.Body.SectPr.PgSz.H != 15840 {
				t.Error("buildEmptyBodyXML: pgSz H != 15840")
			}
		}
		if doc.doc.Body.SectPr.PgMar == nil {
			t.Error("buildEmptyBodyXML: PgMar is nil")
		}
		// Verify no header/footer references (Pitfall 4)
		if len(doc.doc.Body.SectPr.HdrFtrRef) > 0 {
			t.Error("buildEmptyBodyXML: sectPr has header references")
		}
		if len(doc.doc.Body.SectPr.FtrRef) > 0 {
			t.Error("buildEmptyBodyXML: sectPr has footer references")
		}
	}
}

// mockPackageAround wraps an XML payload in a minimal .docx ZIP so it
// can be opened by OpenReader.
func mockPackageAround(t *testing.T, docXML []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	ctXML := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`)
	addZipEntry(t, zw, "[Content_Types].xml", ctXML)
	addZipEntry(t, zw, "_rels/.rels", []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`))
	addZipEntry(t, zw, "word/document.xml", docXML)
	addZipEntry(t, zw, "word/_rels/document.xml.rels", []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
</Relationships>`))

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
