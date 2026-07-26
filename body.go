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

func (d *Document) insertBodyOrderAtPIndex(pIdx int) {
	if d.doc.Body == nil {
		return
	}
	order := d.doc.Body.ElemOrder
	if len(order) == 0 {
		d.doc.Body.ElemOrder = make([]wml.BodyElemType, len(d.doc.Body.P))
		for i := range d.doc.Body.ElemOrder {
			d.doc.Body.ElemOrder[i] = wml.BodyP
		}
		return
	}
	// Find the position in ElemOrder that corresponds to the pIdx-th paragraph
	pCount := 0
	for i, e := range order {
		if e == wml.BodyP {
			if pCount == pIdx {
				d.doc.Body.ElemOrder = append(order, 0)
				copy(d.doc.Body.ElemOrder[i+1:], d.doc.Body.ElemOrder[i:])
				d.doc.Body.ElemOrder[i] = wml.BodyP
				return
			}
			pCount++
		}
	}
	// Append at end
	d.doc.Body.ElemOrder = append(d.doc.Body.ElemOrder, wml.BodyP)
}

// syncBodyOrder ensures ElemOrder is populated for all elements in P and Tbl.
// Called after opening/creating a document when ElemOrder may be nil.
func (d *Document) syncBodyOrder() {
	if d.doc == nil || d.doc.Body == nil {
		return
	}
	if len(d.doc.Body.ElemOrder) > 0 {
		return
	}
	// Count existing elements
	total := len(d.doc.Body.P) + len(d.doc.Body.Tbl)
	if total == 0 {
		return
	}
	d.doc.Body.ElemOrder = make([]wml.BodyElemType, total)
	pCount := len(d.doc.Body.P)
	for i := 0; i < pCount; i++ {
		d.doc.Body.ElemOrder[i] = wml.BodyP
	}
	for i := pCount; i < total; i++ {
		d.doc.Body.ElemOrder[i] = wml.BodyTbl
	}
}

func (d *Document) deleteBodyOrderAtPIndex(pIdx int) {
	if d.doc.Body == nil {
		return
	}
	order := d.doc.Body.ElemOrder
	if len(order) == 0 {
		return
	}
	pCount := 0
	for i, e := range order {
		if e == wml.BodyP {
			if pCount == pIdx {
				d.doc.Body.ElemOrder = append(order[:i], order[i+1:]...)
				return
			}
			pCount++
		}
	}
}

func (d *Document) Body() []BodyElement {
	if d == nil {
		panic("wordingo: Body called on nil Document")
	}
	if d.doc == nil || d.doc.Body == nil {
		return nil
	}

	var elems []BodyElement
	order := d.doc.Body.ElemOrder
	if len(order) > 0 {
		pIdx := 0
		tblIdx := 0
		for _, elemType := range order {
			switch elemType {
			case wml.BodyP:
				if pIdx < len(d.doc.Body.P) {
					elems = append(elems, BodyElement{
						Type: ElementParagraph,
						Para: &Paragraph{ct: d.doc.Body.P[pIdx], doc: d},
					})
					pIdx++
				}
			case wml.BodyTbl:
				if tblIdx < len(d.doc.Body.Tbl) {
					elems = append(elems, BodyElement{
						Type: ElementTable,
						Table: &TableBuilder{ct: d.doc.Body.Tbl[tblIdx], doc: d},
					})
					tblIdx++
				}
			}
		}
	} else {
		for _, p := range d.doc.Body.P {
			elems = append(elems, BodyElement{
				Type: ElementParagraph,
				Para: &Paragraph{ct: p, doc: d},
			})
		}
		for _, tbl := range d.doc.Body.Tbl {
			elems = append(elems, BodyElement{
				Type: ElementTable,
				Table: &TableBuilder{ct: tbl, doc: d},
			})
		}
	}
	return elems
}
