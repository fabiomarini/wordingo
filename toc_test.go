package wordingo

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fabiomarini/wordingo/internal/wml"
)

// saveToBuf serializes the document and returns the bytes.
func saveToBuf(t *testing.T, d *Document) []byte {
	t.Helper()
	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	return buf.Bytes()
}

// zipParts returns the saved package parts as a name->content map.
func zipParts(t *testing.T, data []byte) map[string][]byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	parts := make(map[string][]byte, len(zr.File))
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", f.Name, err)
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("read %s: %v", f.Name, err)
		}
		parts[f.Name] = b
	}
	return parts
}

// assertWellFormed parses raw XML bytes strictly.
func assertWellFormed(t *testing.T, name string, data []byte) {
	t.Helper()
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		if _, err := dec.Token(); err != nil {
			if err == io.EOF {
				return
			}
			t.Fatalf("%s is not well-formed XML: %v", name, err)
		}
	}
}

// newHeadingsDoc returns a fresh document with a Title plus headings
// and body text mixed at levels 1..3.
func newHeadingsDoc(t *testing.T) *Document {
	t.Helper()
	doc, err := Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	doc.AddParagraph("Report").SetStyle("Title")
	doc.AddParagraph("Introduction").SetStyle("Heading1")
	doc.AddParagraph("Some intro text.")
	doc.AddParagraph("Background").SetStyle("Heading2")
	doc.AddParagraph("More text.")
	doc.AddParagraph("Deep Dive").SetStyle("Heading3")
	doc.AddParagraph("Wrap-up").SetStyle("Heading1")
	return doc
}

func TestHeadings(t *testing.T) {
	doc := newHeadingsDoc(t)
	hs := doc.Headings()
	want := []Heading{
		{Level: 1, Text: "Introduction", Style: "Heading1"},
		{Level: 2, Text: "Background", Style: "Heading2"},
		{Level: 3, Text: "Deep Dive", Style: "Heading3"},
		{Level: 1, Text: "Wrap-up", Style: "Heading1"},
	}
	if len(hs) != len(want) {
		t.Fatalf("got %d headings, want %d: %+v", len(hs), len(want), hs)
	}
	for i := range want {
		if hs[i] != want[i] {
			t.Errorf("heading[%d] = %+v, want %+v", i, hs[i], want[i])
		}
	}
}

func TestHeadingsOutlineLevel(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	// A paragraph styled with a custom style that carries an explicit
	// outline level is still a heading (Word's \u TOC switch).
	lvl := int64(1)
	custom := doc.AddParagraph("Custom Heading").SetStyle("MyHeading")
	custom.X().PPr.OutlineLvl = &wml.CT_OutlineLvl{Val: &lvl}
	doc.AddParagraph("plain text")

	hs := doc.Headings()
	if len(hs) != 1 {
		t.Fatalf("got %d headings, want 1: %+v", len(hs), hs)
	}
	if hs[0].Level != 2 || hs[0].Text != "Custom Heading" || hs[0].Style != "MyHeading" {
		t.Errorf("unexpected heading: %+v", hs[0])
	}
}

func TestHeadingsSkipsEmptyText(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("").SetStyle("Heading1")
	doc.AddParagraph("   ").SetStyle("Heading2")
	doc.AddParagraph("Real").SetStyle("Heading1")
	hs := doc.Headings()
	if len(hs) != 1 || hs[0].Text != "Real" {
		t.Fatalf("got %+v, want only the non-empty heading", hs)
	}
}

func TestAddTableOfContentsStructure(t *testing.T) {
	doc := newHeadingsDoc(t)
	toc, err := doc.AddTableOfContents(nil)
	if err != nil {
		t.Fatalf("AddTableOfContents: %v", err)
	}
	if toc.Title() == nil || toc.Title().Text() != "Table of Contents" {
		t.Errorf("TOC title = %q, want %q", toc.Title().Text(), "Table of Contents")
	}

	paras := doc.Paragraphs()
	if got := len(paras); got != 13 {
		t.Fatalf("got %d paragraphs, want 13 (8 body + 5 TOC)", got)
	}
	// Create() starts with one blank Normal paragraph, so the body is:
	// 0 blank, 1 Title, 2 Introduction, 3 text, 4 Background, 5 text,
	// 6 Deep Dive, 7 Wrap-up. The TOC section follows at 8..12.
	title := paras[8]
	entries := paras[9:13]

	// Title paragraph: field begin (dirty), TOC instruction, separate.
	titleRuns := title.X().R
	if len(titleRuns) != 4 {
		t.Fatalf("title paragraph has %d runs, want 4", len(titleRuns))
	}
	fb := titleRuns[0].FldChar
	if fb == nil || fb.Type == nil || *fb.Type != "begin" || fb.Dirty == nil || *fb.Dirty != "true" {
		t.Errorf("title run[0] = %+v, want fldChar begin dirty=true", fb)
	}
	instr := titleRuns[1].InstrText
	if instr == nil || instr.Value != ` TOC \o "1-3" \h \z \u ` {
		t.Errorf("title run[1] instr = %q, want %q", instrTextOf(titleRuns[1]), ` TOC \o "1-3" \h \z \u `)
	}
	if sep := titleRuns[2].FldChar; sep == nil || sep.Type == nil || *sep.Type != "separate" {
		t.Errorf("title run[2] = %+v, want fldChar separate", sep)
	}
	if tt := titleRuns[3].T; tt == nil || tt.Value != "Table of Contents" {
		t.Errorf("title run[3] text = %+v, want the cached title", tt)
	}

	// Entries: 4 headings, all within 3 levels.
	if len(entries) != 4 {
		t.Fatalf("got %d TOC entries, want 4", len(entries))
	}
	wantTexts := []string{"Introduction", "Background", "Deep Dive", "Wrap-up"}
	wantAnchors := []string{"_Toc00000000", "_Toc00000001", "_Toc00000002", "_Toc00000003"}
	wantLevels := []int{1, 2, 3, 1} // entry indentation follows the heading level
	for i, e := range entries {
		ep := e.X()
		if got := entryText(ep); got != wantTexts[i] {
			t.Errorf("entry[%d] text = %q, want %q", i, got, wantTexts[i])
		}
		if len(ep.Hyperlink) != 1 || ep.Hyperlink[0].Anchor == nil || *ep.Hyperlink[0].Anchor != wantAnchors[i] {
			t.Errorf("entry[%d] hyperlink anchor = %+v, want %q", i, ep.Hyperlink, wantAnchors[i])
		}
		// The cached result must display the heading text first, then
		// the tab and the PAGEREF page number — never a leading "0".
		hlRuns := ep.Hyperlink[0].R
		if len(hlRuns) < 7 || hlRuns[0].T == nil || hlRuns[0].T.Value != wantTexts[i] {
			t.Errorf("entry[%d] first hyperlink run = %+v, want heading text first", i, hlRuns[0])
		}
		if hlRuns[1].Tab == nil {
			t.Errorf("entry[%d] second hyperlink run missing tab", i)
		}
		if len(hlRuns) >= 4 && hlRuns[2].FldChar == nil {
			t.Errorf("entry[%d] PAGEREF field not inside hyperlink", i)
		}
		if len(ep.R) != 0 {
			t.Errorf("entry[%d] has %d direct runs; all entry content must live in the hyperlink", i, len(ep.R))
		}
		// Right tab with dot leader; left indent grows with level.
		if ep.PPr == nil || ep.PPr.Tabs == nil || len(ep.PPr.Tabs.Tab) != 1 {
			t.Fatalf("entry[%d] missing right tab", i)
		}
		tab := ep.PPr.Tabs.Tab[0]
		if tab.Val == nil || *tab.Val != "right" || tab.Leader == nil || *tab.Leader != "dot" || tab.Pos == nil || *tab.Pos != 9360 {
			t.Errorf("entry[%d] tab = %+v, want right dot tab at 9360", i, tab)
		}
		if ep.PPr.Ind == nil || ep.PPr.Ind.Left == nil || *ep.PPr.Ind.Left != int64(wantLevels[i]-1)*226 {
			t.Errorf("entry[%d] ind = %+v, want left=%d", i, ep.PPr.Ind, int64(wantLevels[i]-1)*226)
		}
		// PAGEREF field present.
		if !strings.Contains(entryXML(t, ep), "PAGEREF") {
			t.Errorf("entry[%d] missing PAGEREF instruction", i)
		}
	}

	// Last entry closes the TOC field: the closing character is the
	// final run inside the last entry's hyperlink, after the PAGEREF end.
	for i := 0; i < 4; i++ {
		hl := entries[i].X().Hyperlink[0]
		ends := 0
		for _, r := range hl.R {
			if r.FldChar != nil && r.FldChar.Type != nil && *r.FldChar.Type == "end" {
				ends++
			}
		}
		want := 1 // the PAGEREF field's own end
		if i == 3 {
			want = 2 // plus the closing character of the TOC field
		}
		if ends != want {
			t.Errorf("entry[%d] has %d end field chars, want %d", i, ends, want)
		}
	}

	// Headings carry matching bookmarks.
	headingParas := []*wml.CT_P{paras[2].X(), paras[4].X(), paras[6].X(), paras[7].X()}
	for i, hp := range headingParas {
		if len(hp.BookmarkStart) != 1 || hp.BookmarkStart[0].Name == nil || *hp.BookmarkStart[0].Name != wantAnchors[i] {
			t.Errorf("heading[%d] bookmark = %+v, want %q", i, hp.BookmarkStart, wantAnchors[i])
		}
		if len(hp.BookmarkEnd) != 1 {
			t.Errorf("heading[%d] missing bookmark end", i)
		}
	}

	// Settings patched for update-on-open.
	parts := zipParts(t, saveToBuf(t, doc))
	settings := string(parts["word/settings.xml"])
	if !strings.Contains(settings, `w:updateFields w:val="true"`) {
		t.Error("settings.xml missing w:updateFields w:val=\"true\"")
	}
}

// entryText returns the heading-text portion of a TOC entry paragraph
// (the text run before the tab; the PAGEREF field's cached "0" is
// excluded).
func entryText(p *wml.CT_P) string {
	for _, hl := range p.Hyperlink {
		for _, r := range hl.R {
			if r.FldChar != nil {
				return "" // field content reached without finding text
			}
			if r.T != nil {
				return r.T.Value
			}
		}
	}
	return ""
}

func entryXML(t *testing.T, p *wml.CT_P) string {
	t.Helper()
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		t.Fatalf("encode entry: %v", err)
	}
	return buf.String()
}

func instrTextOf(r *wml.CT_R) string {
	if r == nil || r.InstrText == nil {
		return ""
	}
	return r.InstrText.Value
}

func TestAddTableOfContentsCustomOptions(t *testing.T) {
	doc := newHeadingsDoc(t)
	updateOff := false
	_, err := doc.AddTableOfContents(&TOCOptions{
		Title:        "Contents",
		Levels:       1,
		UpdateOnOpen: &updateOff,
	})
	if err != nil {
		t.Fatalf("AddTableOfContents: %v", err)
	}
	paras := doc.Paragraphs()
	// Only Heading1 paragraphs become entries (2 of them).
	entries := 0
	for _, p := range paras {
		if strings.Contains(entryXML(t, p.X()), "PAGEREF") {
			entries++
		}
	}
	if entries != 2 {
		t.Errorf("got %d entries, want 2 (levels clamped to 1)", entries)
	}
	title := paras[8]
	if title.Text() != "Contents" {
		t.Errorf("title = %q, want %q", title.Text(), "Contents")
	}
	if instr := instrTextOf(title.X().R[1]); instr != ` TOC \o "1-1" \h \z \u ` {
		t.Errorf("instruction = %q, want levels 1-1", instr)
	}

	// UpdateOnOpen=false must leave settings.xml untouched.
	parts := zipParts(t, saveToBuf(t, doc))
	if settings := string(parts["word/settings.xml"]); strings.Contains(settings, "updateFields") {
		t.Error("settings.xml patched despite UpdateOnOpen=false")
	}
}

func TestInsertTableOfContentsBefore(t *testing.T) {
	doc := newHeadingsDoc(t)
	target := doc.Paragraphs()[2] // the "Introduction" heading
	toc, err := doc.InsertTableOfContentsBefore(target, nil)
	if err != nil {
		t.Fatalf("InsertTableOfContentsBefore: %v", err)
	}
	if toc == nil {
		t.Fatal("returned nil TOC")
	}
	paras := doc.Paragraphs()
	// TOC must sit right before the target heading. Layout after insert:
	// 0 blank Normal, 1 Title, 2 TOC title, 3..6 entries, 7 Introduction.
	if paras[2].Text() != "Table of Contents" {
		t.Errorf("paras[2] = %q, want TOC title", paras[2].Text())
	}
	if paras[7].Text() != "Introduction" {
		t.Errorf("paras[7] = %q, want Introduction (target moved after TOC)", paras[7].Text())
	}
	// Body order preserved through the body iterator.
	elems := doc.Body()
	if len(elems) != len(paras) {
		t.Errorf("Body() returned %d elements, want %d", len(elems), len(paras))
	}
	if elems[2].Type != ElementParagraph || elems[2].Para.Text() != "Table of Contents" {
		t.Errorf("Body()[2] = %+v, want the TOC title", elems[2])
	}
}

func TestInsertTableOfContentsBeforeMissingTarget(t *testing.T) {
	doc := newHeadingsDoc(t)
	other, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	foreign := other.AddParagraph("elsewhere")

	if _, err := doc.InsertTableOfContentsBefore(nil, nil); err != ErrTOCTargetNotFound {
		t.Errorf("nil target: got %v, want ErrTOCTargetNotFound", err)
	}
	if _, err := doc.InsertTableOfContentsBefore(foreign, nil); err != ErrTOCTargetNotFound {
		t.Errorf("foreign target: got %v, want ErrTOCTargetNotFound", err)
	}
}

func TestAddTableOfContentsEmptyHeadings(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("No headings here.")
	if _, err := doc.AddTableOfContents(nil); err != nil {
		t.Fatalf("AddTableOfContents: %v", err)
	}
	paras := doc.Paragraphs()
	if len(paras) != 5 { // blank Normal + body + title + hint + field end
		t.Fatalf("got %d paragraphs, want 5", len(paras))
	}
	hint := paras[3].Text()
	if !strings.Contains(hint, "No headings found") {
		t.Errorf("hint text = %q, want a no-headings hint", hint)
	}
	// The field must still be closed.
	lastRuns := paras[4].X().R
	if last := lastRuns[len(lastRuns)-1].FldChar; last == nil || last.Type == nil || *last.Type != "end" {
		t.Errorf("missing field end paragraph: %+v", lastRuns)
	}
}

func TestTOCBookmarkUniquenessAcrossCalls(t *testing.T) {
	doc := newHeadingsDoc(t)
	if _, err := doc.AddTableOfContents(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := doc.AddTableOfContents(nil); err != nil {
		t.Fatal(err)
	}
	// Bookmark ids and names must be unique across both TOCs.
	seenNames := map[string]bool{}
	seenIDs := map[int64]bool{}
	for _, p := range doc.Paragraphs() {
		for _, bs := range p.X().BookmarkStart {
			if bs.Name == nil || bs.ID == nil {
				t.Fatalf("bookmark missing name/id: %+v", bs)
			}
			if seenNames[*bs.Name] {
				t.Errorf("duplicate bookmark name %q", *bs.Name)
			}
			if seenIDs[*bs.ID] {
				t.Errorf("duplicate bookmark id %d", *bs.ID)
			}
			seenNames[*bs.Name] = true
			seenIDs[*bs.ID] = true
		}
	}
	if len(seenNames) != 8 {
		t.Errorf("got %d unique bookmarks, want 8", len(seenNames))
	}
}

func TestTOCBookmarkIDsStartAboveExisting(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("Existing").SetStyle("Heading1")
	// Pre-existing bookmark via escape hatch (simulating an opened doc).
	id := int64(42)
	name := "existing_mark"
	p.X().BookmarkStart = append(p.X().BookmarkStart, &wml.CT_BookmarkStart{ID: &id, Name: &name})
	p.X().BookmarkEnd = append(p.X().BookmarkEnd, &wml.CT_BookmarkEnd{ID: &id})

	if _, err := doc.AddTableOfContents(nil); err != nil {
		t.Fatal(err)
	}
	// New bookmark must not reuse id 42 nor the name.
	found := false
	for _, bp := range doc.Paragraphs() {
		for _, bs := range bp.X().BookmarkStart {
			if bs.Name != nil && *bs.Name == "existing_mark" {
				continue
			}
			if bs.ID != nil && *bs.ID == 42 {
				t.Errorf("new bookmark reused id 42")
			}
			if bs.Name != nil && *bs.Name == "_Toc00000042" {
				t.Errorf("new bookmark reused the derived name for id 42")
			}
			found = true
		}
	}
	if !found {
		t.Error("no new bookmark found")
	}
}

func TestTOCRoundTrip(t *testing.T) {
	doc := newHeadingsDoc(t)
	if _, err := doc.AddTableOfContents(nil); err != nil {
		t.Fatal(err)
	}
	data := saveToBuf(t, doc)

	// Strict well-formedness of the re-encoded parts.
	parts := zipParts(t, data)
	assertWellFormed(t, "word/document.xml", parts["word/document.xml"])
	assertWellFormed(t, "word/settings.xml", parts["word/settings.xml"])

	// Reopen through the library and check the structure survives.
	re, err := OpenReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer re.Close()
	if hs := re.Headings(); len(hs) != 4 {
		t.Errorf("reopened Headings() = %+v, want 4", hs)
	}
	paras := re.Paragraphs()
	if len(paras) != 13 {
		t.Fatalf("reopened paragraphs = %d, want 13", len(paras))
	}
	if !strings.Contains(instrTextOf(paras[8].X().R[1]), "TOC") {
		t.Error("TOC instruction lost on round trip")
	}
	if len(paras[2].X().BookmarkStart) != 1 {
		t.Error("heading bookmark lost on round trip")
	}
	// Warnings: TOC must not introduce unknown-style noise.
	if ws := re.Warnings(); len(ws) != 0 {
		t.Errorf("unexpected warnings: %v", ws)
	}
}

func TestTOCOnOpenedFile(t *testing.T) {
	src := filepath.Join("testdata", "word", "01-blank.docx")
	doc, err := Open(src)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()
	doc.AddParagraph("Alpha").SetStyle("Heading1")
	doc.AddParagraph("Beta").SetStyle("Heading2")
	if _, err := doc.AddTableOfContents(&TOCOptions{Levels: 2}); err != nil {
		t.Fatal(err)
	}

	// Same-path save must be rejected before truncating the source.
	if err := doc.Save(src); err == nil {
		t.Fatal("Save over the open source file should fail")
	}

	out := filepath.Join(t.TempDir(), "out.docx")
	if err := doc.Save(out); err != nil {
		t.Fatalf("Save: %v", err)
	}
	re, err := Open(out)
	if err != nil {
		t.Fatalf("reopen saved: %v", err)
	}
	defer re.Close()
	if hs := re.Headings(); len(hs) != 2 || hs[0].Text != "Alpha" || hs[1].Text != "Beta" {
		t.Errorf("reopened headings = %+v", hs)
	}
	data := saveToBuf(t, re)
	parts := zipParts(t, data)
	assertWellFormed(t, "word/document.xml", parts["word/document.xml"])
	if !strings.Contains(string(parts["word/settings.xml"]), "updateFields") {
		t.Error("saved settings.xml missing updateFields")
	}
}

func TestTOCNoStyleWarnings(t *testing.T) {
	doc := newHeadingsDoc(t)
	if _, err := doc.AddTableOfContents(nil); err != nil {
		t.Fatal(err)
	}
	saveToBuf(t, doc) // serializeBody runs checkStyleNames
	if ws := doc.Warnings(); len(ws) != 0 {
		t.Errorf("TOC introduced warnings: %v", ws)
	}
}

func TestTOCPreservesTableElementOrder(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	section := doc.AddParagraph("Section").SetStyle("Heading1")
	if _, err := doc.AddTable([][]string{{"a", "b"}, {"c", "d"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := doc.InsertTableOfContentsBefore(section, nil); err != nil {
		t.Fatal(err)
	}
	// Table must still come after the TOC in document order, and the
	// entry paragraphs must sit between the title and the heading.
	elems := doc.Body()
	if len(elems) != 5 { // blank Normal + title + entry + Section + table
		t.Fatalf("got %d elements, want 5", len(elems))
	}
	if elems[1].Type != ElementParagraph || elems[1].Para.Text() != "Table of Contents" {
		t.Errorf("elems[1] = %+v, want the TOC title", elems[1])
	}
	if elems[2].Type != ElementParagraph {
		t.Errorf("elems[2] = %+v, want the TOC entry paragraph", elems[2])
	}
	if elems[3].Type != ElementParagraph || elems[3].Para.Text() != "Section" {
		t.Errorf("elems[3] = %+v, want the Section heading", elems[3])
	}
	if elems[4].Type != ElementTable {
		t.Errorf("elems[4] = %+v, want the table", elems[4])
	}
}

func TestOpenKeepsSourceFileForLazyParts(t *testing.T) {
	src := filepath.Join("testdata", "word", "01-blank.docx")
	doc, err := Open(src)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	out := filepath.Join(t.TempDir(), "copy.docx")
	if err := doc.Save(out); err != nil {
		t.Fatalf("Save unmodified copy: %v", err)
	}
	doc.Close()
	// The copy must contain the unmodified parts.
	a, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	pa, pb := zipParts(t, a), zipParts(t, b)
	for name, payload := range pa {
		if name == "[Content_Types].xml" || strings.Contains(name, "/_rels/") || name == "_rels/.rels" {
			continue
		}
		if !bytes.Equal(payload, pb[name]) {
			t.Errorf("part %s differs in untouched copy", name)
		}
	}
}
