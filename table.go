package wordingo

import (
	"fmt"

	"github.com/fabiomarini/wordingo/internal/wml"
)

// TableBuilder provides a fluent API for building complex tables.
type TableBuilder struct {
	ct  *wml.CT_Tbl
	doc *Document
}

// RowBuilder provides a fluent API for building table rows.
type RowBuilder struct {
	ct  *wml.CT_Tr
	doc *Document
}

// CellBuilder provides a fluent API for building table cells.
type CellBuilder struct {
	ct  *wml.CT_Tc
	doc *Document
}

// ---- Table builder ----

// X returns the underlying CT_Tbl for escape-hatch access.
func (tb *TableBuilder) X() *wml.CT_Tbl {
	if tb == nil {
		panic("wordingo: X called on nil TableBuilder")
	}
	return tb.ct
}

// SetTableStyle sets the table style by name (D-03).
func (tb *TableBuilder) SetTableStyle(name string) *TableBuilder {
	if tb == nil {
		panic("wordingo: SetTableStyle called on nil TableBuilder")
	}
	if tb.ct.TblPr == nil {
		tb.ct.TblPr = &wml.CT_TblPr{}
	}
	tb.ct.TblPr.TblStyle = &wml.CT_TblStyle{Val: &name}
	tb.doc.dirty = true
	return tb
}

// SetWidth sets the table width (D-04). wType is "dxa", "pct", or "auto".
func (tb *TableBuilder) SetWidth(w int64, wType string) *TableBuilder {
	if tb == nil {
		panic("wordingo: SetWidth called on nil TableBuilder")
	}
	if tb.ct.TblPr == nil {
		tb.ct.TblPr = &wml.CT_TblPr{}
	}
	tb.ct.TblPr.TblW = &wml.CT_TblW{W: &w, Type: &wType}
	tb.doc.dirty = true
	return tb
}

// SetBorders sets table-level borders.
func (tb *TableBuilder) SetBorders(b *TableBorders) *TableBuilder {
	if tb == nil {
		panic("wordingo: SetBorders called on nil TableBuilder")
	}
	if b == nil {
		return tb
	}
	if tb.ct.TblPr == nil {
		tb.ct.TblPr = &wml.CT_TblPr{}
	}
	tb.ct.TblPr.Borders = toTblBorders(b)
	tb.doc.dirty = true
	return tb
}

// SetShading sets table-level shading.
func (tb *TableBuilder) SetShading(val, fill string) *TableBuilder {
	if tb == nil {
		panic("wordingo: SetShading called on nil TableBuilder")
	}
	if tb.ct.TblPr == nil {
		tb.ct.TblPr = &wml.CT_TblPr{}
	}
	tb.ct.TblPr.Shd = &wml.CT_Shd{Val: &val, Fill: &fill}
	tb.doc.dirty = true
	return tb
}

// Row returns the RowBuilder for the row at index idx. Grows the row
// slice if idx is beyond current length.
func (tb *TableBuilder) Row(idx int) *RowBuilder {
	if tb == nil {
		panic("wordingo: Row called on nil TableBuilder")
	}
	for len(tb.ct.Tr) <= idx {
		tb.ct.Tr = append(tb.ct.Tr, &wml.CT_Tr{Tc: []*wml.CT_Tc{}})
	}
	return &RowBuilder{ct: tb.ct.Tr[idx], doc: tb.doc}
}

// ---- Row builder ----

// X returns the underlying CT_Tr for escape-hatch access.
func (rb *RowBuilder) X() *wml.CT_Tr {
	if rb == nil {
		panic("wordingo: X called on nil RowBuilder")
	}
	return rb.ct
}

// SetBorders sets borders on every cell in this row.
func (rb *RowBuilder) SetBorders(b *TableBorders) *RowBuilder {
	if rb == nil {
		panic("wordingo: SetBorders called on nil RowBuilder")
	}
	if b == nil {
		return rb
	}
	for _, tc := range rb.ct.Tc {
		if tc.TcPr == nil {
			tc.TcPr = &wml.CT_TcPr{}
		}
		tc.TcPr.Borders = toTcBorders(b)
	}
	rb.doc.dirty = true
	return rb
}

// Cell returns the CellBuilder for the cell at index idx. Grows the cell
// slice if idx is beyond current length.
func (rb *RowBuilder) Cell(idx int) *CellBuilder {
	if rb == nil {
		panic("wordingo: Cell called on nil RowBuilder")
	}
	for len(rb.ct.Tc) <= idx {
		rb.ct.Tc = append(rb.ct.Tc, &wml.CT_Tc{P: []*wml.CT_P{{}}})
	}
	return &CellBuilder{ct: rb.ct.Tc[idx], doc: rb.doc}
}

// ---- Cell builder ----

// X returns the underlying CT_Tc for escape-hatch access.
func (cb *CellBuilder) X() *wml.CT_Tc {
	if cb == nil {
		panic("wordingo: X called on nil CellBuilder")
	}
	return cb.ct
}

// SetText sets the text content of the first paragraph in the cell.
func (cb *CellBuilder) SetText(text string) *CellBuilder {
	if cb == nil {
		panic("wordingo: SetText called on nil CellBuilder")
	}
	if len(cb.ct.P) == 0 {
		cb.ct.P = []*wml.CT_P{{}}
	}
	p := cb.ct.P[0]
	p.R = []*wml.CT_R{{T: &wml.CT_Text{Value: text}}}
	cb.doc.dirty = true
	return cb
}

// SetShading sets cell-level shading (D-02 per-cell).
func (cb *CellBuilder) SetShading(val, fill string) *CellBuilder {
	if cb == nil {
		panic("wordingo: SetShading called on nil CellBuilder")
	}
	if cb.ct.TcPr == nil {
		cb.ct.TcPr = &wml.CT_TcPr{}
	}
	cb.ct.TcPr.Shd = &wml.CT_Shd{Val: &val, Fill: &fill}
	cb.doc.dirty = true
	return cb
}

// SetBold sets bold on the first run of the first paragraph.
func (cb *CellBuilder) SetBold(b bool) *CellBuilder {
	if cb == nil {
		panic("wordingo: SetBold called on nil CellBuilder")
	}
	if len(cb.ct.P) == 0 {
		cb.ct.P = []*wml.CT_P{{}}
	}
	p := cb.ct.P[0]
	if len(p.R) == 0 {
		p.R = []*wml.CT_R{{}}
	}
	if p.R[0].RPr == nil {
		p.R[0].RPr = &wml.CT_RPr{}
	}
	p.R[0].RPr.B = &wml.CT_OnOff{Val: &b}
	cb.doc.dirty = true
	return cb
}

// SetWidth sets the cell width.
func (cb *CellBuilder) SetWidth(w int64, wType string) *CellBuilder {
	if cb == nil {
		panic("wordingo: SetWidth called on nil CellBuilder")
	}
	if cb.ct.TcPr == nil {
		cb.ct.TcPr = &wml.CT_TcPr{}
	}
	cb.ct.TcPr.TcW = &wml.CT_TblW{W: &w, Type: &wType}
	cb.doc.dirty = true
	return cb
}

// MergeRight sets GridSpan to merge this cell with the cells to the right.
func (cb *CellBuilder) MergeRight() *CellBuilder {
	if cb == nil {
		panic("wordingo: MergeRight called on nil CellBuilder")
	}
	if cb.ct.TcPr == nil {
		cb.ct.TcPr = &wml.CT_TcPr{}
	}
	v := int64(2) // default: merge 2 cells
	cb.ct.TcPr.GridSpan = &wml.CT_GridSpan{Val: &v}
	cb.doc.dirty = true
	return cb
}

// MergeDown sets VMerge to "restart" on this cell (start of vertical merge).
func (cb *CellBuilder) MergeDown() *CellBuilder {
	if cb == nil {
		panic("wordingo: MergeDown called on nil CellBuilder")
	}
	if cb.ct.TcPr == nil {
		cb.ct.TcPr = &wml.CT_TcPr{}
	}
	restart := "restart"
	cb.ct.TcPr.VMerge = &wml.CT_VMerge{Val: &restart}
	cb.doc.dirty = true
	return cb
}

// ---- Document table methods ----

// AddTable creates a simple grid table from string data and appends it
// to the document. Returns the TableBuilder for further customization.
// Tables are appended after all paragraphs (v1 limitation per D-24).
func (d *Document) AddTable(data [][]string) (*TableBuilder, error) {
	if d == nil {
		panic("wordingo: AddTable called on nil Document")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("wordingo: AddTable: empty data")
	}

	cols := 0
	for _, row := range data {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if cols == 0 {
		return nil, fmt.Errorf("wordingo: AddTable: no columns")
	}

	ct := &wml.CT_Tbl{
		TblPr:   &wml.CT_TblPr{},
		TblGrid: &wml.CT_TblGrid{},
	}

	// Build grid columns
	for range cols {
		ct.TblGrid.GridCol = append(ct.TblGrid.GridCol, &wml.CT_GridCol{})
	}

	// Build rows
	for _, rowData := range data {
		tr := &wml.CT_Tr{}
		for ci := 0; ci < cols; ci++ {
			tc := &wml.CT_Tc{P: []*wml.CT_P{{}}}
			if ci < len(rowData) && rowData[ci] != "" {
				tc.P[0].R = []*wml.CT_R{{T: &wml.CT_Text{Value: rowData[ci]}}}
			}
			tr.Tc = append(tr.Tc, tc)
		}
		ct.Tr = append(ct.Tr, tr)
	}

	if d.doc.Body == nil {
		d.doc.Body = &wml.CT_Body{SectPr: defaultSectPr()}
	}
	d.doc.Body.Tbl = append(d.doc.Body.Tbl, ct)
	d.dirty = true
	return &TableBuilder{ct: ct, doc: d}, nil
}

// AddTableBuilder returns a TableBuilder for building a complex table.
// The table is appended to the document body. Tables are appended after
// all paragraphs (v1 limitation per D-24).
func (d *Document) AddTableBuilder() *TableBuilder {
	if d == nil {
		panic("wordingo: AddTableBuilder called on nil Document")
	}
	ct := &wml.CT_Tbl{
		TblPr:   &wml.CT_TblPr{},
		TblGrid: &wml.CT_TblGrid{},
	}
	if d.doc.Body == nil {
		d.doc.Body = &wml.CT_Body{SectPr: defaultSectPr()}
	}
	d.doc.Body.Tbl = append(d.doc.Body.Tbl, ct)
	d.dirty = true
	return &TableBuilder{ct: ct, doc: d}
}

// Tables returns the document's body tables.
func (d *Document) Tables() []*TableBuilder {
	if d.doc == nil || d.doc.Body == nil {
		return nil
	}
	tbs := make([]*TableBuilder, len(d.doc.Body.Tbl))
	for i, t := range d.doc.Body.Tbl {
		tbs[i] = &TableBuilder{ct: t, doc: d}
	}
	return tbs
}

// ---- Helper functions ----

// toTblBorders converts a TableBorders value to the WML type.
func toTblBorders(b *TableBorders) *wml.CT_TblBorders {
	tb := &wml.CT_TblBorders{}
	if b.Top != nil {
		tb.Top = toTblBorder(b.Top)
	}
	if b.Bottom != nil {
		tb.Bottom = toTblBorder(b.Bottom)
	}
	if b.Left != nil {
		tb.Left = toTblBorder(b.Left)
	}
	if b.Right != nil {
		tb.Right = toTblBorder(b.Right)
	}
	if b.InsideH != nil {
		tb.InsideH = toTblBorder(b.InsideH)
	}
	if b.InsideV != nil {
		tb.InsideV = toTblBorder(b.InsideV)
	}
	return tb
}

// toTcBorders converts a TableBorders value to the WML cell borders type.
func toTcBorders(b *TableBorders) *wml.CT_TcBorders {
	tb := &wml.CT_TcBorders{}
	if b.Top != nil {
		tb.Top = toTblBorder(b.Top)
	}
	if b.Bottom != nil {
		tb.Bottom = toTblBorder(b.Bottom)
	}
	if b.Left != nil {
		tb.Left = toTblBorder(b.Left)
	}
	if b.Right != nil {
		tb.Right = toTblBorder(b.Right)
	}
	return tb
}

// toTblBorder converts a BorderDef to the WML border type.
func toTblBorder(b *BorderDef) *wml.CT_TblBorder {
	style := b.Style
	size := b.Size
	color := b.Color
	return &wml.CT_TblBorder{
		Val:   &style,
		Sz:    &size,
		Color: &color,
	}
}
