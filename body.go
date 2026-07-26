package wordingo

import "github.com/fabiomarini/wordingo/internal/wml"

type BodyElementType int

const (
	ElementParagraph BodyElementType = iota
	ElementTable
)

type BodyElement struct {
	Type  BodyElementType
	Para  *Paragraph
	Table *TableBuilder
}

type BodyParagraph struct {
	P *Paragraph
}

func (bp BodyParagraph) X() *wml.CT_P {
	return bp.P.X()
}

type BodyTable struct {
	T *TableBuilder
}

func (bt BodyTable) X() *wml.CT_Tbl {
	return bt.T.X()
}

func (d *Document) Body() []BodyElement {
	if d == nil {
		panic("wordingo: Body called on nil Document")
	}
	if d.doc == nil || d.doc.Body == nil {
		return nil
	}

	pi := 0
	ti := 0
	var elems []BodyElement

	for pi < len(d.doc.Body.P) || ti < len(d.doc.Body.Tbl) {
		if pi < len(d.doc.Body.P) {
			elems = append(elems, BodyElement{
				Type: ElementParagraph,
				Para: &Paragraph{ct: d.doc.Body.P[pi], doc: d},
			})
			pi++
		} else if ti < len(d.doc.Body.Tbl) {
			elems = append(elems, BodyElement{
				Type: ElementTable,
				Table: &TableBuilder{ct: d.doc.Body.Tbl[ti], doc: d},
			})
			ti++
		}
	}
	return elems
}
