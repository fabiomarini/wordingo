package wordingo

import (
	"testing"
)

func TestBody_BodyAccessor(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	// Create() adds one empty Normal paragraph already
	doc.AddParagraph("P1")
	_, err = doc.AddTable([][]string{{"R1C1"}, {"R2C1"}})
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("P2")

	elems := doc.Body()
	if len(elems) != 4 {
		t.Fatalf("expected 4 body elements (1 empty + P1 + table + P2), got %d", len(elems))
	}

	if elems[1].Type != ElementParagraph {
		t.Errorf("element 1 should be paragraph, got type %d", elems[1].Type)
	}
	if elems[2].Type != ElementParagraph {
		t.Errorf("element 2 should be paragraph, got type %d", elems[2].Type)
	}
	if elems[3].Type != ElementTable {
		t.Errorf("element 3 should be table, got type %d", elems[3].Type)
	}
	if elems[3].Table == nil {
		t.Fatal("element 3 Table is nil")
	}
}

func TestBody_BackwardCompat(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	// Create() has 1 empty paragraph + our content
	doc.AddParagraph("P")
	_, err = doc.AddTable([][]string{{"T1"}})
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("P2")

	paras := doc.Paragraphs()
	if len(paras) != 3 {
		t.Errorf("expected 3 paragraphs (1 empty + P + P2), got %d", len(paras))
	}
	if paras[1].Text() != "P" {
		t.Errorf("paragraph 1 text: got %q", paras[1].Text())
	}

	tables := doc.Tables()
	if len(tables) != 1 {
		t.Errorf("expected 1 table, got %d", len(tables))
	}
}

func TestBody_NilBody(t *testing.T) {
	d := &Document{}
	elems := d.Body()
	if elems != nil {
		t.Errorf("expected nil for nil body, got %v", elems)
	}
}

func TestBody_XEscapeHatch(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("x")
	tb, err := doc.AddTable([][]string{{"y"}})
	if err != nil {
		t.Fatal(err)
	}

	bp := BodyParagraph{P: p}
	bt := BodyTable{T: tb}

	if bp.X() != p.X() {
		t.Error("BodyParagraph.X() should return same CT_P")
	}
	if bt.X() != tb.X() {
		t.Error("BodyTable.X() should return same CT_Tbl")
	}
}
