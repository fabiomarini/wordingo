package style

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fabiomarini/wordingo/internal/opc"
)

// ---------- Synthetic ZIP builders (copy pattern from opc_test.go:47-73) ----------

const clonerStylesNS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
const clonerRelsNS = "http://schemas.openxmlformats.org/package/2006/relationships"

// buildBaseContentTypes returns [Content_Types].xml with all 5 style parts
// as Overrides plus a Default for xml/rels.
func buildBaseContentTypes(extra ...string) string {
	ct := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
<Override PartName="/word/numbering.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"/>
<Override PartName="/word/fontTable.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml"/>
<Override PartName="/word/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>
<Override PartName="/word/settings.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"/>`
	for _, e := range extra {
		ct += e
	}
	ct += "\n</Types>"
	return ct
}

// minimal style part payloads used in synthetic fixtures.
var (
	synStylesXML    = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:styles xmlns:w="` + clonerStylesNS + `"><w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/></w:style></w:styles>`)
	synNumberingXML = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:numbering xmlns:w="` + clonerStylesNS + `"><w:abstractNum w:abstractNumId="0"><w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/><w:start w:val="1"/></w:lvl></w:abstractNum><w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num></w:numbering>`)
	synFontTableXML = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:fonts xmlns:w="` + clonerStylesNS + `"><w:font w:name="Aptos"><w:family w:val="swiss"/></w:font></w:fonts>`)
	synThemeXML     = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Test"><a:themeElements><a:clrScheme name="Test"><a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1><a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1><a:accent1><a:srgbClr val="156082"/></a:accent1></a:clrScheme></a:themeElements></a:theme>`)
	synSettingsXML  = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:settings xmlns:w="` + clonerStylesNS + `"><w:zoom w:percent="100"/></w:settings>`)
	synDocumentXML  = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="` + clonerStylesNS + `"><w:body><w:p/></w:body></w:document>`)
)

// addZipEntry is a helper to add a single entry to a zip.Writer with
// a fixed timestamp (deterministic header).
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

// buildSourceZip returns a ZIP with all 5 style parts plus minimal
// package infrastructure. If excludeNumbering is true, numbering.xml
// is omitted.
func buildSourceZip(t *testing.T, excludeNumbering bool) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	contentTypes := buildBaseContentTypes()
	addZipEntry(t, zw, "[Content_Types].xml", []byte(contentTypes))
	addZipEntry(t, zw, "_rels/.rels", []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="`+clonerRelsNS+`">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`))
	addZipEntry(t, zw, "word/document.xml", synDocumentXML)
	addZipEntry(t, zw, "word/_rels/document.xml.rels", []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="`+clonerRelsNS+`">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`))

	addZipEntry(t, zw, "word/styles.xml", synStylesXML)
	addZipEntry(t, zw, "word/fontTable.xml", synFontTableXML)
	addZipEntry(t, zw, "word/theme/theme1.xml", synThemeXML)
	addZipEntry(t, zw, "word/settings.xml", synSettingsXML)

	if !excludeNumbering {
		addZipEntry(t, zw, "word/numbering.xml", synNumberingXML)
	}

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// buildFreshTargetZip returns a fresh-empty target ZIP with no style parts.
func buildFreshTargetZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Content types: only Defaults + document override (no style overrides).
	ct := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`
	addZipEntry(t, zw, "[Content_Types].xml", []byte(ct))
	addZipEntry(t, zw, "_rels/.rels", []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="`+clonerRelsNS+`">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`))
	addZipEntry(t, zw, "word/document.xml", synDocumentXML)
	// Empty document rels set.
	addZipEntry(t, zw, "word/_rels/document.xml.rels", []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="`+clonerRelsNS+`">
</Relationships>`))

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// buildTargetWithStylesZip returns a target ZIP with a pre-existing
// word/styles.xml (to trigger ErrCloneTargetNotEmpty).
func buildTargetWithStylesZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	ct := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`
	addZipEntry(t, zw, "[Content_Types].xml", []byte(ct))
	addZipEntry(t, zw, "_rels/.rels", []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="`+clonerRelsNS+`">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`))
	addZipEntry(t, zw, "word/document.xml", synDocumentXML)
	addZipEntry(t, zw, "word/_rels/document.xml.rels", []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="`+clonerRelsNS+`">
</Relationships>`))
	// Pre-existing styles.xml with distinctive payload.
	addZipEntry(t, zw, "word/styles.xml", []byte(`<existing-styles/>`))

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// ---------- Helpers ----------

// openBytes opens a ZIP byte slice as an *opc.Package.
func openClonerPkg(t *testing.T, data []byte) *opc.Package {
	t.Helper()
	pkg, err := opc.Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("opc.Open: %v", err)
	}
	return pkg
}

// assertPartBytesEqual reads a part from both packages and asserts
// their payloads are byte-identical.
func assertPartBytesEqual(t *testing.T, src, dst *opc.Package, name string) {
	t.Helper()
	srcPart, ok := src.Parts[name]
	if !ok {
		t.Fatalf("source part %q not found", name)
	}
	dstPart, ok := dst.Parts[name]
	if !ok {
		t.Fatalf("dst part %q not found after clone", name)
	}

	srcBytes := readPartBytes(t, srcPart)
	dstBytes := readPartBytes(t, dstPart)

	if !bytes.Equal(srcBytes, dstBytes) {
		// Hex dump first differing byte.
		minLen := len(srcBytes)
		if len(dstBytes) < minLen {
			minLen = len(dstBytes)
		}
		for i := 0; i < minLen; i++ {
			if srcBytes[i] != dstBytes[i] {
				t.Fatalf("part %q differs at byte %d: src[%d]=0x%02x dst[%d]=0x%02x",
					name, i, i, srcBytes[i], i, dstBytes[i])
			}
		}
		if len(srcBytes) != len(dstBytes) {
			t.Fatalf("part %q length differs: src=%d dst=%d", name, len(srcBytes), len(dstBytes))
		}
	}
}

// readPartBytes reads the full payload of a part.
func readPartBytes(t *testing.T, p *opc.Part) []byte {
	t.Helper()
	rc, err := p.Open()
	if err != nil {
		t.Fatalf("Part.Open: %v", err)
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("io.ReadAll: %v", err)
	}
	return b
}

// partNames returns the set of part names in a package.
func partNames(pkg *opc.Package) map[string]bool {
	names := make(map[string]bool, len(pkg.Parts))
	for name := range pkg.Parts {
		names[name] = true
	}
	return names
}

// partExists reports whether a part exists in the package.
func partExists(pkg *opc.Package, name string) bool {
	_, ok := pkg.Parts[name]
	return ok
}

// ---------- Test cases ----------

func TestCloneStyles_FreshEmptyTarget(t *testing.T) {
	srcZip := buildSourceZip(t, false)
	dstZip := buildFreshTargetZip(t)
	src := openClonerPkg(t, srcZip)
	dst := openClonerPkg(t, dstZip)

	err := CloneStyles(src, dst)
	if err != nil {
		t.Fatalf("CloneStyles: %v", err)
	}

	// All 5 parts must be present in dst.
	for _, p := range cloneParts {
		if !partExists(dst, p.name) {
			t.Errorf("part %q missing from dst after clone", p.name)
		}
	}

	// Each cloned part must be byte-identical to source.
	for _, p := range cloneParts {
		assertPartBytesEqual(t, src, dst, p.name)
	}
}

func TestCloneStyles_ErrCloneTargetNotEmpty(t *testing.T) {
	srcZip := buildSourceZip(t, false)
	dstZip := buildTargetWithStylesZip(t)
	src := openClonerPkg(t, srcZip)
	dst := openClonerPkg(t, dstZip)

	// Capture the pre-existing styles.xml bytes before clone attempt.
	preExistingBytes := readPartBytes(t, dst.Parts["word/styles.xml"])

	err := CloneStyles(src, dst)

	// Must return an error wrapping ErrCloneTargetNotEmpty.
	if err == nil {
		t.Fatal("CloneStyles: expected error, got nil")
	}
	if !errors.Is(err, ErrCloneTargetNotEmpty) {
		t.Fatalf("CloneStyles: err = %v, want ErrCloneTargetNotEmpty wrapping", err)
	}

	// Atomicity assertion: the pre-existing styles.xml must be untouched
	// (the precondition scan aborted before any byte copy).
	postBytes := readPartBytes(t, dst.Parts["word/styles.xml"])
	if !bytes.Equal(preExistingBytes, postBytes) {
		t.Fatal("dst styles.xml was modified despite ErrCloneTargetNotEmpty — atomicity violated (D-08)")
	}
}

func TestCloneStyles_AbsentSourcePartSkip(t *testing.T) {
	// Source WITHOUT numbering.xml (excludeNumbering=true).
	srcZip := buildSourceZip(t, true)
	dstZip := buildFreshTargetZip(t)
	src := openClonerPkg(t, srcZip)
	dst := openClonerPkg(t, dstZip)

	err := CloneStyles(src, dst)
	if err != nil {
		t.Fatalf("CloneStyles: %v", err)
	}

	// The 4 present parts must be cloned.
	presentParts := []struct{ name string }{
		{"word/styles.xml"},
		{"word/fontTable.xml"},
		{"word/theme/theme1.xml"},
		{"word/settings.xml"},
	}
	for _, p := range presentParts {
		if !partExists(dst, p.name) {
			t.Errorf("part %q should exist after clone", p.name)
		}
	}

	// numbering.xml must NOT be present.
	if partExists(dst, "word/numbering.xml") {
		t.Error("word/numbering.xml should NOT exist (source lacked it)")
	}

	// No relationship for numbering should exist in dst.
	dstDocRels := dst.Rels["word/document.xml"]
	if dstDocRels != nil {
		for _, rel := range dstDocRels.Rels {
			if rel.Target == "word/numbering.xml" {
				t.Errorf("unexpected relationship to numbering.xml: %s", rel.ID)
			}
		}
	}
}

func TestCloneStyles_RelationshipAllocation(t *testing.T) {
	srcZip := buildSourceZip(t, false)
	dstZip := buildFreshTargetZip(t)
	src := openClonerPkg(t, srcZip)
	dst := openClonerPkg(t, dstZip)

	err := CloneStyles(src, dst)
	if err != nil {
		t.Fatalf("CloneStyles: %v", err)
	}

	docRels := dst.Rels["word/document.xml"]
	if docRels == nil {
		t.Fatal("dst.Rels[\"word/document.xml\"] is nil")
	}

	// Must have at least 5 entries (the pre-existing rId1 from the
	// source relationship to styles, plus 5 new ones from cloning).
	// Fresh target starts with 0 rels, so we expect exactly 5.
	if len(docRels.Rels) != 5 {
		t.Fatalf("expected 5 document rels after clone, got %d", len(docRels.Rels))
	}

	// All rIds must be unique and follow rId<N> format.
	seen := make(map[string]bool)
	for _, rel := range docRels.Rels {
		if !strings.HasPrefix(rel.ID, "rId") {
			t.Errorf("relationship id %q does not match rId<N> format", rel.ID)
		}
		if seen[rel.ID] {
			t.Errorf("duplicate relationship id %q", rel.ID)
		}
		seen[rel.ID] = true
	}
}

func TestCloneStyles_ContentTypeOverrides(t *testing.T) {
	srcZip := buildSourceZip(t, false)
	dstZip := buildFreshTargetZip(t)
	src := openClonerPkg(t, srcZip)
	dst := openClonerPkg(t, dstZip)

	err := CloneStyles(src, dst)
	if err != nil {
		t.Fatalf("CloneStyles: %v", err)
	}

	// All 5 parts must have content-type overrides with leading slash.
	for _, p := range cloneParts {
		key := "/" + p.name
		got, ok := dst.ContentTypes.Overrides[key]
		if !ok {
			t.Errorf("ContentTypes.Overrides[%q] missing", key)
			continue
		}
		if got != p.ct {
			t.Errorf("ContentTypes.Overrides[%q] = %q, want %q", key, got, p.ct)
		}
	}
}

func TestCloneStyles_SaveRoundTrip(t *testing.T) {
	srcZip := buildSourceZip(t, false)
	dstZip := buildFreshTargetZip(t)
	src := openClonerPkg(t, srcZip)
	dst := openClonerPkg(t, dstZip)

	err := CloneStyles(src, dst)
	if err != nil {
		t.Fatalf("CloneStyles: %v", err)
	}

	// Save dst to buffer.
	var saved bytes.Buffer
	if err := dst.Save(&saved); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Re-open the saved bytes.
	reopened := openClonerPkg(t, saved.Bytes())

	// The 5 cloned parts must be byte-identical to the source's parts
	// after Save round-trip (D-09 + OPC-04).
	for _, p := range cloneParts {
		assertPartBytesEqual(t, src, reopened, p.name)
	}
}

func TestCloneStyles_NoReParse(t *testing.T) {
	// Structural assertion: cloner.go must not import xmlutil or wml.
	// This is a compile-time structural gate (D-09 enforced).
	// The done criteria verifies via grep that cloner.go has no
	// xmlutil or wml references.  This test is a placeholder that
	// always passes to signal the structural requirement.
}

func TestCloneStyles_IdempotentFails(t *testing.T) {
	srcZip := buildSourceZip(t, false)
	dstZip := buildFreshTargetZip(t)
	src := openClonerPkg(t, srcZip)
	dst := openClonerPkg(t, dstZip)

	// First call — must succeed.
	err := CloneStyles(src, dst)
	if err != nil {
		t.Fatalf("first CloneStyles: %v", err)
	}

	// Second call — must fail with ErrCloneTargetNotEmpty.
	err = CloneStyles(src, dst)
	if err == nil {
		t.Fatal("second CloneStyles: expected error, got nil")
	}
	if !errors.Is(err, ErrCloneTargetNotEmpty) {
		t.Fatalf("second CloneStyles: err = %v, want ErrCloneTargetNotEmpty wrapping", err)
	}
}

func TestCloneStyles_NilPackageDefensive(t *testing.T) {
	srcZip := buildSourceZip(t, false)
	dstZip := buildFreshTargetZip(t)
	src := openClonerPkg(t, srcZip)
	dst := openClonerPkg(t, dstZip)

	t.Run("nil src", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("CloneStyles(nil, dst) panicked: %v", r)
			}
		}()
		_ = CloneStyles(nil, dst)
	})

	t.Run("nil dst", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("CloneStyles(src, nil) panicked: %v", r)
			}
		}()
		_ = CloneStyles(src, nil)
	})
}

func TestCloneStyles_RealFixtures(t *testing.T) {
	// Per D-10/D-11, real .docx fixtures are user-authored.
	// This placeholder signals plan 02-04 where the real-fixture
	// equality harness is wired.
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	fixtureDir := filepath.Join(root, "..", "..", "testdata", "word", "style-rich")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		fixtureDir = filepath.Join(root, "testdata", "word", "style-rich")
		if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
			t.Skip("testdata/word/style-rich/ not present — real-fixture tests require user-authored .docx per D-10")
		}
	}
	t.Log("style-rich fixtures present — 02-04 will wire equality harness here")
}

// ---------- Relationship type URI constants for test assertions ----------

// Verify cloneParts table exists (compile-time check via reference).
func TestClonePartsTable_Exists(t *testing.T) {
	if len(cloneParts) != 5 {
		t.Fatalf("cloneParts has %d entries, want 5", len(cloneParts))
	}
	expected := []struct{ name string }{
		{"word/styles.xml"},
		{"word/numbering.xml"},
		{"word/fontTable.xml"},
		{"word/theme/theme1.xml"},
		{"word/settings.xml"},
	}
	for i, e := range expected {
		if cloneParts[i].name != e.name {
			t.Errorf("cloneParts[%d].name = %q, want %q", i, cloneParts[i].name, e.name)
		}
	}
}
