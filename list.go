package wordingo

import (
	"bytes"
	"fmt"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

const (
	relNumbering = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering"
	ctNumbering  = "application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"
)

// ListBuilder provides a fluent API for building ordered and bulleted lists.
type ListBuilder struct {
	numID int64
	doc   *Document
	last  *wml.CT_P
}

// X returns the last paragraph added via AddItem for escape-hatch access.
func (lb *ListBuilder) X() *wml.CT_P {
	if lb == nil {
		panic("wordingo: X called on nil ListBuilder")
	}
	return lb.last
}

// AddItem appends a list item at the given nesting level and returns the
// ListBuilder for chaining. level must be in range 0-8 (full 9-level depth
// per D-17). The paragraph is linked to the list's numbering definition
// via NumPr (numId + ilvl).
func (lb *ListBuilder) AddItem(text string, level int) *ListBuilder {
	if lb == nil {
		panic("wordingo: AddItem called on nil ListBuilder")
	}
	if level < 0 {
		level = 0
	}
	if level > 8 {
		level = 8
	}
	ilvl := int64(level)
	ct := &wml.CT_P{
		PPr: &wml.CT_PPr{
			NumPr: &wml.CT_NumPr{
				ILvl:  &wml.CT_ILvl{Val: &ilvl},
				NumId: &wml.CT_NumId{Val: &lb.numID},
			},
		},
		R: []*wml.CT_R{{T: &wml.CT_Text{Value: text}}},
	}
	lb.doc.doc.Body.AppendP(ct)
	lb.last = ct
	lb.doc.dirty = true
	return lb
}

// readOrCreateNumbering reads word/numbering.xml from the package if it
// exists, or returns an empty CT_Numbering for first-use creation.
func (d *Document) readOrCreateNumbering() *wml.CT_Numbering {
	part, ok := d.pkg.Parts["word/numbering.xml"]
	if !ok {
		return &wml.CT_Numbering{}
	}
	rc, err := part.Open()
	if err != nil {
		return &wml.CT_Numbering{}
	}
	defer rc.Close()
	var nb wml.CT_Numbering
	dec := xmlutil.NewSafeDecoder(rc, opc.MaxPartBytes)
	if err := dec.Decode(&nb); err != nil {
		return &wml.CT_Numbering{}
	}
	return &nb
}

// writeNumbering serializes nb and replaces word/numbering.xml in the
// package. Lazily creates the part, content-type override, and
// document-level relationship on first call.
func (d *Document) writeNumbering(nb *wml.CT_Numbering) {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(nb); err != nil {
		d.warn("wordingo: encode numbering.xml: %v", err)
		return
	}
	if err := enc.Flush(); err != nil {
		d.warn("wordingo: flush numbering.xml: %v", err)
		return
	}

	d.pkg.MarkModified("word/numbering.xml", buf.Bytes())

	// Add content-type override on first write.
	if _, ok := d.pkg.ContentTypes.Overrides["/word/numbering.xml"]; !ok {
		d.pkg.ContentTypes.Overrides["/word/numbering.xml"] = ctNumbering
	}

	// Add document-level relationship on first write.
	rels := d.pkg.Rels["word/document.xml"]
	if rels == nil {
		rels = &opc.Relationships{}
		d.pkg.Rels["word/document.xml"] = rels
	}
	hasNumRel := false
	for _, r := range rels.Rels {
		if r.Type == relNumbering {
			hasNumRel = true
			break
		}
	}
	if !hasNumRel {
		rID := rels.NextRID()
		rels.Rels = append(rels.Rels, opc.Relationship{
			ID:     rID,
			Type:   relNumbering,
			Target: "numbering.xml",
		})
	}
}

// findMaxAbstractNumID returns the next available abstractNumId by scanning
// all existing abstractNum entries. Returns 0 if none exist.
func findMaxAbstractNumID(nb *wml.CT_Numbering) int64 {
	var max int64 = -1
	for _, an := range nb.AbstractNum {
		if an.AbstractNumID != nil && *an.AbstractNumID > max {
			max = *an.AbstractNumID
		}
	}
	return max + 1
}

// findMaxNumID returns the next available numId by scanning all existing
// num entries. Returns 1 if none exist — reserving numId 0 for Word's
// built-in ListNumber definition (Pitfall 3).
func findMaxNumID(nb *wml.CT_Numbering) int64 {
	var max int64 = 0 // start at 0 to reserve numId=0 for Word ListNumber
	for _, n := range nb.Num {
		if n.NumID != nil && *n.NumID > max {
			max = *n.NumID
		}
	}
	return max + 1
}

// makeLevels generates 9 CT_Lvl entries (levels 0-8) for either ordered
// (decimal, "%N.") or bulleted (bullet chars per level) numbering.
func makeLevels(ordered bool) []*wml.CT_Lvl {
	bulletChars := []string{
		"\u2022", "\u25E6", "\u25AA", "\u25AB",
		"\u2022", "\u25E6", "\u25AA", "\u25AB",
		"\u2022",
	}
	levels := make([]*wml.CT_Lvl, 9)
	for i := 0; i < 9; i++ {
		ilvl := int64(i)
		startOne := int64(1)
		lvl := &wml.CT_Lvl{
			ILvl:  &ilvl,
			Start: &wml.CT_Start{Val: &startOne},
		}
		if ordered {
			numFmt := "decimal"
			text := fmt.Sprintf("%%%d.", i+1)
			lvl.NumFmt = &wml.CT_NumFmt{Val: &numFmt}
			lvl.LvlText = &wml.CT_LvlText{Val: &text}
		} else {
			numFmt := "bullet"
			lvl.NumFmt = &wml.CT_NumFmt{Val: &numFmt}
			lvl.LvlText = &wml.CT_LvlText{Val: &bulletChars[i]}
		}
		levels[i] = lvl
	}
	return levels
}

// AddList creates a new ordered or bulleted list with auto-generated
// numbering definitions and returns a *ListBuilder for adding items.
// Ordered lists use numFmt=decimal with "%N." level text; bulleted lists
// use numFmt=bullet. Full 9-level depth is configured on the abstract
// numbering definition (levels 0-8).
//
// Numbering definitions are merged with any existing numbering.xml content
// from the package (template numbering is preserved). Auto-generated
// abstractNumId and numId values are computed by scanning all existing
// entries to avoid collisions (Pitfall 3).
func (d *Document) AddList(ordered bool) *ListBuilder {
	if d == nil {
		panic("wordingo: AddList called on nil Document")
	}
	nb := d.readOrCreateNumbering()

	nextAbsID := findMaxAbstractNumID(nb)
	nextNumID := findMaxNumID(nb)

	absNum := &wml.CT_AbstractNum{
		AbstractNumID: &nextAbsID,
		Lvl:           makeLevels(ordered),
	}
	nb.AbstractNum = append(nb.AbstractNum, absNum)

	num := &wml.CT_Num{
		NumID:         &nextNumID,
		AbstractNumID: &wml.CT_AbstractNumID{Val: &nextAbsID},
	}
	nb.Num = append(nb.Num, num)

	d.writeNumbering(nb)
	d.dirty = true
	return &ListBuilder{numID: nextNumID, doc: d}
}

// AddListFromSlice creates an ordered or bulleted list from a string
// slice. Each string becomes a level-0 list item. Returns the
// ListBuilder for further customization.
func (d *Document) AddListFromSlice(items []string, ordered bool) *ListBuilder {
	if d == nil {
		panic("wordingo: AddListFromSlice called on nil Document")
	}
	lb := d.AddList(ordered)
	for _, item := range items {
		lb.AddItem(item, 0)
	}
	return lb
}

// AddNumberingDef creates a custom numbering definition with the given
// numFmt and start value applied to all 9 levels. Returns a ListBuilder
// linked to the new definition. Level text uses "%N." template for each
// level N+1.
//
// Example: doc.AddNumberingDef("upperRoman", 1) creates Roman-numeral
// numbering across all 9 levels.
func (d *Document) AddNumberingDef(numFmt string, start int) *ListBuilder {
	if d == nil {
		panic("wordingo: AddNumberingDef called on nil Document")
	}
	nb := d.readOrCreateNumbering()

	nextAbsID := findMaxAbstractNumID(nb)
	nextNumID := findMaxNumID(nb)

	levels := make([]*wml.CT_Lvl, 9)
	for i := 0; i < 9; i++ {
		ilvl := int64(i)
		startVal := int64(start)
		text := fmt.Sprintf("%%%d.", i+1)
		levels[i] = &wml.CT_Lvl{
			ILvl:    &ilvl,
			NumFmt:  &wml.CT_NumFmt{Val: &numFmt},
			LvlText: &wml.CT_LvlText{Val: &text},
			Start:   &wml.CT_Start{Val: &startVal},
		}
	}

	absNum := &wml.CT_AbstractNum{
		AbstractNumID: &nextAbsID,
		Lvl:           levels,
	}
	nb.AbstractNum = append(nb.AbstractNum, absNum)

	num := &wml.CT_Num{
		NumID:         &nextNumID,
		AbstractNumID: &wml.CT_AbstractNumID{Val: &nextAbsID},
	}
	nb.Num = append(nb.Num, num)

	d.writeNumbering(nb)
	d.dirty = true
	return &ListBuilder{numID: nextNumID, doc: d}
}
