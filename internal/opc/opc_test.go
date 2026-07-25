package opc

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
)

// --- synthetic fixture builders -------------------------------------

const synContentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`

const synRootRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

const synDocRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="https://example.invalid/evil.png" TargetMode="External"/>
</Relationships>`

const synDocument = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p/></w:body></w:document>`

const synStyles = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"></w:styles>`

// synCustomXML is the unmodeled-part payload used to prove byte-exact
// pass-through (OPC-04).
var synCustomXML = []byte("<?xml version=\"1.0\"?>\n<custom>\x00\x01 binary-ish blob \xf0\x9f\x93\x84</custom>")

// buildSyntheticZip writes a minimal in-memory .docx.
func buildSyntheticZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name string, payload []byte) {
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
	add("[Content_Types].xml", []byte(synContentTypes))
	add("_rels/.rels", []byte(synRootRels))
	add("word/document.xml", []byte(synDocument))
	add("word/_rels/document.xml.rels", []byte(synDocRels))
	add("word/styles.xml", []byte(synStyles))
	add("customXml/item1.xml", synCustomXML)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func openBytes(t *testing.T, b []byte) *Package {
	t.Helper()
	pkg, err := Open(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return pkg
}

// HasFixtures reports whether the user-authored fixture corpus (D-02)
// is present; per-producer tests skip otherwise.
func HasFixtures(t *testing.T) bool {
	t.Helper()
	root := filepath.Join("..", "..", "testdata")
	_, err := os.Stat(filepath.Join(root, "word", "blank.docx"))
	if err != nil {
		t.Skip("testdata/ fixture corpus not present (D-02 pending) — synthetic path active")
		return false
	}
	return true
}

// DiffParts compares two unzipped part sets per part (D-04 harness).
// Modeled manifests ([Content_Types].xml and .rels parts) are
// re-serialized canonically on save, so they are excluded from the
// byte comparison; OPC-04 byte-identity covers unmodeled parts.
func DiffParts(t *testing.T, a, b []byte) {
	t.Helper()
	manifest := func(name string) bool {
		return name == "[Content_Types].xml" || isRelsPath(name)
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

// --- Task 1: open-path tests ----------------------------------------

func TestOpen(t *testing.T) {
	pkg := openBytes(t, buildSyntheticZip(t))

	wantParts := []string{
		"[Content_Types].xml", "_rels/.rels", "word/document.xml",
		"word/_rels/document.xml.rels", "word/styles.xml", "customXml/item1.xml",
	}
	for _, name := range wantParts {
		if _, ok := pkg.Parts[name]; !ok {
			t.Errorf("part %s not enumerated", name)
		}
	}
	if got := pkg.ContentTypes.TypeFor("word/document.xml"); !strings.Contains(got, "document.main") {
		t.Errorf("TypeFor(document.xml) = %q", got)
	}
	if got := pkg.ContentTypes.TypeFor("foo.xml"); got != "application/xml" {
		t.Errorf("TypeFor(foo.xml) = %q, want application/xml default", got)
	}
	root, ok := pkg.Rels[""]
	if !ok || len(root.Rels) != 1 {
		t.Fatalf("root rels = %+v", root)
	}
	docRels := pkg.Rels["word/document.xml"]
	if docRels == nil || len(docRels.Rels) != 2 {
		t.Fatalf("document rels = %+v", docRels)
	}
	if pkg.Conformance != Transitional {
		t.Errorf("Conformance = %v, want Transitional", pkg.Conformance)
	}
	// External rel surfaced as warning, never fetched.
	found := false
	for _, w := range pkg.Warnings() {
		if strings.Contains(w, "external relationship") {
			found = true
		}
	}
	if !found {
		t.Errorf("no external-relationship warning: %v", pkg.Warnings())
	}
}

func TestOpenStrictConformance(t *testing.T) {
	strictDoc := strings.Replace(synDocument,
		"http://schemas.openxmlformats.org/wordprocessingml/2006/main",
		"http://purl.oclc.org/ooxml/wordprocessingml/main", 1)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name, payload string) {
		w, _ := zw.Create(name)
		w.Write([]byte(payload))
	}
	add("[Content_Types].xml", synContentTypes)
	add("_rels/.rels", synRootRels)
	add("word/document.xml", strictDoc)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	pkg := openBytes(t, buf.Bytes())
	if pkg.Conformance != Strict {
		t.Errorf("Conformance = %v, want Strict", pkg.Conformance)
	}
}

func TestSafety(t *testing.T) {
	build := func(t *testing.T, entries map[string][]byte, extra int) []byte {
		t.Helper()
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		for name, payload := range entries {
			w, err := zw.Create(name)
			if err != nil {
				t.Fatal(err)
			}
			w.Write(payload)
		}
		for i := 0; i < extra; i++ {
			w, err := zw.Create("filler/part" + strings.Repeat("x", 1) + itoa(i) + ".xml")
			if err != nil {
				t.Fatal(err)
			}
			w.Write([]byte("<x/>"))
		}
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}

	t.Run("path traversal", func(t *testing.T) {
		b := build(t, map[string][]byte{"../evil": []byte("x")}, 0)
		_, err := Open(bytes.NewReader(b), int64(len(b)))
		if !errors.Is(err, ErrUnsafePath) {
			t.Fatalf("err = %v, want ErrUnsafePath", err)
		}
	})

	t.Run("backslash", func(t *testing.T) {
		b := build(t, map[string][]byte{`word\evil.xml`: []byte("x")}, 0)
		_, err := Open(bytes.NewReader(b), int64(len(b)))
		if !errors.Is(err, ErrUnsafePath) {
			t.Fatalf("err = %v, want ErrUnsafePath", err)
		}
	})

	t.Run("too many parts", func(t *testing.T) {
		b := build(t, nil, MaxParts+1)
		_, err := Open(bytes.NewReader(b), int64(len(b)))
		if !errors.Is(err, ErrTooManyParts) {
			t.Fatalf("err = %v, want ErrTooManyParts", err)
		}
	})

	t.Run("decompression limit", func(t *testing.T) {
		// An entry declaring > MaxPartBytes uncompressed must be
		// rejected from the central directory alone. Hand-craft via
		// FileHeader with a Store method and a forged size would need
		// raw ZIP editing; instead stream MaxPartBytes+1 of zeros —
		// the declared size trips the cap before any decompress.
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		w, err := zw.Create("big.bin")
		if err != nil {
			t.Fatal(err)
		}
		zeros := make([]byte, 1<<20)
		for i := int64(0); i < (MaxPartBytes>>20)+1; i++ {
			if _, err := w.Write(zeros); err != nil {
				t.Fatal(err)
			}
		}
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
		_, err = Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		if !errors.Is(err, ErrDecompressionLimit) {
			t.Fatalf("err = %v, want ErrDecompressionLimit", err)
		}
	})
}

func TestRelationshipsNextRID(t *testing.T) {
	rs := &Relationships{Rels: []Relationship{
		{ID: "rId1", Type: "t", Target: "a.xml"},
		{ID: "rId2", Type: "t", Target: "b.xml"},
		{ID: "rId3", Type: "t", Target: "c.xml"},
	}}
	// Parse-time scan establishes the high-water mark above rId3;
	// delete rId3, then allocate 100 ids: none may collide or reuse 3.
	rs.rescan()
	rs.Rels = rs.Rels[:2]
	seen := map[string]bool{"rId1": true, "rId2": true}
	for i := 0; i < 100; i++ {
		id := rs.NextRID()
		if seen[id] {
			t.Fatalf("NextRID reissued %s", id)
		}
		if id == "rId3" {
			t.Fatalf("NextRID reused deleted rId3")
		}
		seen[id] = true
		rs.Rels = append(rs.Rels, Relationship{ID: id, Type: "t", Target: "x.xml"})
	}
}

func itoa(i int) string {
	// small local helper to avoid strconv import churn in table builds
	if i == 0 {
		return "0"
	}
	var b [8]byte
	p := len(b)
	for i > 0 {
		p--
		b[p] = byte('0' + i%10)
		i /= 10
	}
	return string(b[p:])
}

// --- Task 2: save-path tests ----------------------------------------

func TestCanonicalOrder(t *testing.T) {
	shuffled := []string{
		"customXml/item1.xml",
		"word/styles.xml",
		"word/_rels/document.xml.rels",
		"docProps/core.xml",
		"word/document.xml",
		"word/_rels/styles.xml.rels",
		"_rels/.rels",
		"word/theme/theme1.xml",
		"[Content_Types].xml",
	}
	got := CanonicalOrder(shuffled)
	want := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"word/document.xml",
		"word/_rels/document.xml.rels",
		"word/styles.xml",
		"word/_rels/styles.xml.rels",
		"word/theme/theme1.xml",
		"customXml/item1.xml",
		"docProps/core.xml",
	}
	if len(got) != len(want) {
		t.Fatalf("CanonicalOrder len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("CanonicalOrder[%d] = %q, want %q\nfull: %v", i, got[i], want[i], got)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	src := buildSyntheticZip(t)
	pkg := openBytes(t, src)

	var out bytes.Buffer
	if err := pkg.Save(&out); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// Per-part byte diff (D-04) — unmodeled customXml part must be
	// byte-identical via raw pass-through (OPC-04).
	DiffParts(t, src, out.Bytes())

	// Saved package starts with [Content_Types].xml then _rels/.rels.
	zr, err := zip.NewReader(bytes.NewReader(out.Bytes()), int64(out.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if zr.File[0].Name != "[Content_Types].xml" || zr.File[1].Name != "_rels/.rels" {
		t.Errorf("first entries = %q, %q", zr.File[0].Name, zr.File[1].Name)
	}
}

func TestRoundTripHostileFixture(t *testing.T) {
	if !HasFixtures(t) {
		return
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "hostile", "customxml-glossary.docx"))
	if err != nil {
		t.Skip("hostile fixture not present")
	}
	pkg, err := Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Open hostile fixture: %v", err)
	}
	var out bytes.Buffer
	if err := pkg.Save(&out); err != nil {
		t.Fatalf("Save hostile fixture: %v", err)
	}
	DiffParts(t, data, out.Bytes())
}

func TestSave(t *testing.T) {
	t.Run("dangling target fails closed", func(t *testing.T) {
		pkg := openBytes(t, buildSyntheticZip(t))
		// Corrupt the graph: point root rel at a nonexistent part.
		pkg.Rels[""].Rels[0].Target = "word/missing.xml"
		var out bytes.Buffer
		err := pkg.Save(&out)
		if !errors.Is(err, ErrInvalidPackage) {
			t.Fatalf("err = %v, want ErrInvalidPackage", err)
		}
		if out.Len() != 0 {
			t.Errorf("wrote %d bytes despite validation failure", out.Len())
		}
	})

	t.Run("deterministic", func(t *testing.T) {
		pkg := openBytes(t, buildSyntheticZip(t))
		var a, b bytes.Buffer
		if err := pkg.Save(&a); err != nil {
			t.Fatal(err)
		}
		if err := pkg.Save(&b); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a.Bytes(), b.Bytes()) {
			t.Error("two consecutive Saves produced different bytes")
		}
	})

	t.Run("modified part serializes, untouched stay raw", func(t *testing.T) {
		src := buildSyntheticZip(t)
		pkg := openBytes(t, src)
		newDoc := []byte(strings.Replace(synDocument, "<w:p/>", "<w:p><w:r><w:t>hi</w:t></w:r></w:p>", 1))
		pkg.MarkModified("word/document.xml", newDoc)
		var out bytes.Buffer
		if err := pkg.Save(&out); err != nil {
			t.Fatal(err)
		}
		zr, err := zip.NewReader(bytes.NewReader(out.Bytes()), int64(out.Len()))
		if err != nil {
			t.Fatal(err)
		}
		payloads := map[string][]byte{}
		for _, f := range zr.File {
			rc, _ := f.Open()
			b, _ := io.ReadAll(rc)
			rc.Close()
			payloads[f.Name] = b
		}
		if !bytes.Equal(payloads["word/document.xml"], newDoc) {
			t.Error("modified part payload mismatch")
		}
		if !bytes.Equal(payloads["customXml/item1.xml"], synCustomXML) {
			t.Error("untouched unmodeled part not byte-identical")
		}
		if !bytes.Equal(payloads["word/styles.xml"], []byte(synStyles)) {
			t.Error("untouched modeled part not byte-identical")
		}
	})
}
