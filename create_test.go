package wordingo

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fabiomarini/wordingo/internal/opc"
)

func TestCreate(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error: %v", err)
	}

	// --- Entry order ---
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}

	if len(zr.File) != 9 {
		t.Fatalf("got %d entries, want 9", len(zr.File))
	}

	if zr.File[0].Name != "[Content_Types].xml" {
		t.Errorf("entry[0] = %q, want [Content_Types].xml", zr.File[0].Name)
	}
	if zr.File[1].Name != "_rels/.rels" {
		t.Errorf("entry[1] = %q, want _rels/.rels", zr.File[1].Name)
	}

	// --- Part set exactly 9, no extras ---
	expectedParts := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"word/document.xml",
		"word/_rels/document.xml.rels",
		"word/styles.xml",
		"word/settings.xml",
		"word/webSettings.xml",
		"word/fontTable.xml",
		"word/theme/theme1.xml",
	}
	partSet := make(map[string]bool)
	for _, f := range zr.File {
		partSet[f.Name] = true
	}
	for _, exp := range expectedParts {
		if !partSet[exp] {
			t.Errorf("missing part: %s", exp)
		}
	}
	if len(partSet) != 9 {
		t.Errorf("part set has %d entries, want 9", len(partSet))
	}
	// No docProps
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "docProps/") {
			t.Errorf("unexpected part: %s (should be omitted)", f.Name)
		}
	}

	// --- Styles content markers ---
	stylesContent := readZipEntry(t, zr, "word/styles.xml")
	for _, marker := range []string{"heading 1", "heading 9", "Title"} {
		if !strings.Contains(stylesContent, marker) {
			t.Errorf("styles.xml missing marker: %q", marker)
		}
	}
	if !strings.Contains(stylesContent, `styleId="`) {
		t.Error("styles.xml missing any styleId attribute")
	}

	// --- Document sectPr attributes ---
	docContent := readZipEntry(t, zr, "word/document.xml")
	if !strings.Contains(docContent, `w:w="12240"`) {
		t.Error("document.xml missing w:w=12240 in sectPr")
	}
	if !strings.Contains(docContent, `w:h="15840"`) {
		t.Error("document.xml missing w:h=15840 in sectPr")
	}
	if !strings.Contains(docContent, `w:top="1440"`) {
		t.Error("document.xml missing w:top=1440 in pgMar")
	}

	// --- Canonical prefix on document.xml root ---
	if !strings.Contains(docContent, `xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"`) {
		t.Error("document.xml missing canonical w namespace declaration")
	}

	// --- Round-trip through opc.Open ---
	pkg, err := opc.Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("opc.Open round-trip: %v", err)
	}
	if len(pkg.Warnings()) > 0 {
		t.Errorf("round-trip has warnings: %v", pkg.Warnings())
	}
}

func readZipEntry(t *testing.T, zr *zip.Reader, name string) string {
	t.Helper()
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

func TestSaveFile(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	path := filepath.Join(t.TempDir(), "savefile.docx")
	if err := doc.SaveFile(path); err != nil {
		t.Fatalf("SaveFile: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("os.Open: %v", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	pkg, err := opc.Open(f, stat.Size())
	if err != nil {
		t.Fatalf("opc.Open: %v", err)
	}
	if len(pkg.Warnings()) != 0 {
		t.Errorf("unexpected warnings: %v", pkg.Warnings())
	}
	if pkg.Conformance != opc.Transitional {
		t.Errorf("Conformance = %v, want Transitional", pkg.Conformance)
	}
}

func TestXEscapeHatch(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	pkg := doc.X()
	if pkg == nil {
		t.Fatal("X() returned nil")
	}
	if pkg.ContentTypes == nil {
		t.Error("pkg.ContentTypes is nil")
	}
	if len(pkg.Parts) == 0 {
		t.Error("pkg.Parts is empty")
	}
	if pkg.Conformance != opc.Transitional {
		t.Errorf("Conformance = %v, want Transitional", pkg.Conformance)
	}
}
