package wordingo

import (
	"bytes"
	"strings"
	"testing"
)

func TestTable_AddTable(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	data := [][]string{
		{"A1", "B1", "C1"},
		{"A2", "B2", "C2"},
	}
	tb, err := doc.AddTable(data)
	if err != nil {
		t.Fatalf("AddTable() error: %v", err)
	}
	if tb == nil {
		t.Fatal("AddTable returned nil")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// Verify tbl element exists
	if !strings.Contains(content, "<w:tbl") {
		t.Error("document.xml missing w:tbl element")
	}

	// Verify all cell text content
	for _, cell := range []string{"A1", "B1", "C1", "A2", "B2", "C2"} {
		if !strings.Contains(content, cell) {
			t.Errorf("document.xml missing cell text %q", cell)
		}
	}

	// Verify TblGrid with 3 gridCol
	if !strings.Contains(content, "<w:tblGrid") {
		t.Error("document.xml missing w:tblGrid")
	}

	// Verify 2 rows (tr)
	if got := strings.Count(content, "<w:tr"); got != 2 {
		t.Errorf("expected 2 rows, got %d", got)
	}
}

func TestTable_AddTable_EmptyDataError(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	_, err = doc.AddTable([][]string{})
	if err == nil {
		t.Error("AddTable with empty data should return error")
	}
}

func TestTable_AddTableBuilder(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	tb := doc.AddTableBuilder()
	if tb == nil {
		t.Fatal("AddTableBuilder returned nil")
	}

	tb.SetTableStyle("TableGrid")
	tb.SetWidth(9000, "dxa")

	// Build header row
	tb.Row(0).Cell(0).SetText("Name").SetBold(true)
	tb.Row(0).Cell(1).SetText("Value").SetBold(true)

	// Build data row
	tb.Row(1).Cell(0).SetText("Item 1")
	tb.Row(1).Cell(1).SetText("42")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// Verify table style
	if !strings.Contains(content, "TableGrid") {
		t.Error("document.xml missing TableGrid style")
	}

	// Verify cell text
	if !strings.Contains(content, "Name") {
		t.Error("document.xml missing 'Name' cell")
	}
	if !strings.Contains(content, "Value") {
		t.Error("document.xml missing 'Value' cell")
	}
	if !strings.Contains(content, "Item 1") {
		t.Error("document.xml missing 'Item 1' cell")
	}
	if !strings.Contains(content, "42") {
		t.Error("document.xml missing '42' cell")
	}

	// Verify bold on header row cells
	if !strings.Contains(content, `<w:b`) {
		t.Error("document.xml missing bold formatting on header cells")
	}

	// Verify tblW with width
	if !strings.Contains(content, `w:w="9000"`) {
		t.Error("document.xml missing table width w=9000")
	}
}

func TestTable_SetBorders(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	tb := doc.AddTableBuilder()
	tb.SetBorders(&TableBorders{
		Top:    &BorderDef{Style: "single", Size: 4, Color: "000000"},
		Bottom: &BorderDef{Style: "single", Size: 4, Color: "000000"},
		Left:   &BorderDef{Style: "single", Size: 4, Color: "000000"},
		Right:  &BorderDef{Style: "single", Size: 4, Color: "000000"},
	})
	tb.Row(0).Cell(0).SetText("Border cell")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// Verify border elements
	if !strings.Contains(content, `<w:tblBorders`) {
		t.Error("document.xml missing tblBorders")
	}
	if !strings.Contains(content, `w:val="single"`) {
		t.Error("document.xml missing border style 'single'")
	}
}

func TestTable_CellShading(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	tb := doc.AddTableBuilder()
	tb.Row(0).Cell(0).SetText("Shaded").SetShading("clear", "D9E2F3")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	if !strings.Contains(content, `w:fill="D9E2F3"`) {
		t.Error("document.xml missing cell shading fill D9E2F3")
	}
	if !strings.Contains(content, `w:val="clear"`) {
		t.Error("document.xml missing cell shading val clear")
	}
}

func TestTable_CellWidth(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	tb := doc.AddTableBuilder()
	tb.Row(0).Cell(0).SetText("Wide").SetWidth(5000, "dxa")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	if !strings.Contains(content, `w:w="5000"`) {
		t.Error("document.xml missing cell width w=5000")
	}
}

func TestTable_MergeRight(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	tb := doc.AddTableBuilder()
	tb.Row(0).Cell(0).SetText("Merged").MergeRight()

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	if !strings.Contains(content, `w:gridSpan`) {
		t.Error("document.xml missing gridSpan for merged cell")
	}
	if !strings.Contains(content, `w:val="2"`) {
		t.Error("document.xml missing gridSpan val=2")
	}
}

func TestTable_MergeDown(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	tb := doc.AddTableBuilder()
	tb.Row(0).Cell(0).SetText("Top").MergeDown()

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	if !strings.Contains(content, `w:vMerge`) {
		t.Error("document.xml missing vMerge for merge-down cell")
	}
	if !strings.Contains(content, `w:val="restart"`) {
		t.Error("document.xml missing vMerge val=restart")
	}
}

func TestTable_XEscapeHatch(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	tb := doc.AddTableBuilder()
	tb.Row(0).Cell(0).SetText("Test")

	ct := tb.X()
	if ct == nil {
		t.Fatal("TableBuilder.X() returned nil")
	}
	if len(ct.Tr) == 0 {
		t.Error("X() table has no rows")
	}
}

func TestTable_RowBorders(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	tb := doc.AddTableBuilder()
	tb.Row(0).Cell(0).SetText("A")
	tb.Row(0).SetBorders(&TableBorders{
		Top:    &BorderDef{Style: "single", Size: 4, Color: "FF0000"},
		Bottom: &BorderDef{Style: "single", Size: 4, Color: "0000FF"},
	})

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	content := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	if !strings.Contains(content, `w:color="FF0000"`) {
		t.Error("document.xml missing row top border color FF0000")
	}
	if !strings.Contains(content, `w:color="0000FF"`) {
		t.Error("document.xml missing row bottom border color 0000FF")
	}
}

func TestTable_TablesAccessor(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	doc.AddTableBuilder()
	tbs := doc.Tables()
	if len(tbs) != 1 {
		t.Errorf("Tables() returned %d builders, want 1", len(tbs))
	}
	if tbs[0] == nil {
		t.Error("Tables()[0] is nil")
	}
}
