package wml

import (
	"encoding/xml"
	"strings"

	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

type BodyElemType int

const (
	BodyP BodyElemType = iota
	BodyTbl
)

// CT_Theme is an opaque theme part root with RawXML body
// (full DrawingML is out of scope for Phase 1).
type CT_Theme struct {
	XMLName xml.Name        `xml:"http://schemas.openxmlformats.org/drawingml/2006/main theme"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// ---- Root / Part types ----

// CT_Document is the root element of word/document.xml.
type CT_Document struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main document"`
	Body    *CT_Body `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main body"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_Body is the document body container.
type CT_Body struct {
	XMLName   xml.Name          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main body"`
	P         []*CT_P          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Tbl       []*CT_Tbl        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tbl"`
	SectPr    *CT_SectPr       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sectPr"`
	Raw       []xmlutil.RawXML `xml:",any"`
	ElemOrder []BodyElemType   `xml:"-"` // insertion order: BodyP or BodyTbl
}

func (b *CT_Body) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if len(b.ElemOrder) > 0 {
		pIdx := 0
		tblIdx := 0
		for _, elemType := range b.ElemOrder {
			switch elemType {
			case BodyP:
				if pIdx < len(b.P) {
					if err := e.Encode(b.P[pIdx]); err != nil {
						return err
					}
					pIdx++
				}
			case BodyTbl:
				if tblIdx < len(b.Tbl) {
					if err := e.Encode(b.Tbl[tblIdx]); err != nil {
						return err
					}
					tblIdx++
				}
			}
		}
	} else {
		for _, p := range b.P {
			if err := e.Encode(p); err != nil {
				return err
			}
		}
		for _, tbl := range b.Tbl {
			if err := e.Encode(tbl); err != nil {
				return err
			}
		}
	}
	for _, raw := range b.Raw {
		if err := e.Encode(raw); err != nil {
			return err
		}
	}
	if b.SectPr != nil {
		if err := e.Encode(b.SectPr); err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// AppendP adds a paragraph to the body and records it in ElemOrder.
func (b *CT_Body) AppendP(p *CT_P) {
	b.P = append(b.P, p)
	b.ElemOrder = append(b.ElemOrder, BodyP)
}

// AppendTbl adds a table to the body and records it in ElemOrder.
func (b *CT_Body) AppendTbl(tbl *CT_Tbl) {
	b.Tbl = append(b.Tbl, tbl)
	b.ElemOrder = append(b.ElemOrder, BodyTbl)
}

// CT_Hdr is a header part root.
type CT_Hdr struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hdr"`
	P       []*CT_P `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_Ftr is a footer part root.
type CT_Ftr struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ftr"`
	P       []*CT_P `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// ---- Paragraph / Run core ----

// CT_P is a paragraph.
type CT_P struct {
	XMLName      xml.Name          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	PPr          *CT_PPr           `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
	BookmarkStart []*CT_BookmarkStart `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bookmarkStart"`
	R            []*CT_R           `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
	Hyperlink    []*CT_Hyperlink   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hyperlink"`
	BookmarkEnd  []*CT_BookmarkEnd `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bookmarkEnd"`
	Raw          []xmlutil.RawXML  `xml:",any"`

	// HyperlinkFirst orders hyperlink children before run children when
	// marshaling (xml:"-" so it never serializes). The TOC generator
	// uses it to emit the closing field character of a TOC field after
	// the entry's hyperlink, matching Word's own TOC layout. All other
	// paragraphs keep the default order: runs, then hyperlinks.
	HyperlinkFirst bool `xml:"-"`
}

// MarshalXML emits the paragraph children in schema order. Runs are
// emitted before hyperlink elements by default; when HyperlinkFirst is
// set the order is reversed (TOC entries need their PAGEREF-bearing
// hyperlink before the closing field character run).
//
// Go's xml package ignores the XMLName field for types that implement
// xml.Marshaler when they are passed directly to Encoder.Encode — the
// start element then carries the Go type name ("CT_P") or an empty
// name. Substitute the canonical name in those cases (the XMLName
// field value itself is only populated on unmarshal); struct-field
// encoding already passes the correct start element.
func (p *CT_P) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if start.Name.Local == "" || start.Name.Local == "CT_P" {
		start.Name = xml.Name{Space: NSWMLMain, Local: "p"}
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if p.PPr != nil {
		if err := e.Encode(p.PPr); err != nil {
			return err
		}
	}
	for _, bs := range p.BookmarkStart {
		if err := e.Encode(bs); err != nil {
			return err
		}
	}
	if p.HyperlinkFirst {
		for _, h := range p.Hyperlink {
			if err := e.Encode(h); err != nil {
				return err
			}
		}
		for _, r := range p.R {
			if err := e.Encode(r); err != nil {
				return err
			}
		}
	} else {
		for _, r := range p.R {
			if err := e.Encode(r); err != nil {
				return err
			}
		}
		for _, h := range p.Hyperlink {
			if err := e.Encode(h); err != nil {
				return err
			}
		}
	}
	for _, be := range p.BookmarkEnd {
		if err := e.Encode(be); err != nil {
			return err
		}
	}
	for _, raw := range p.Raw {
		if err := e.Encode(raw); err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// CT_PPr holds paragraph properties.
type CT_PPr struct {
	XMLName        xml.Name        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
	PStyle         *CT_PStyle      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pStyle"`
	KeepNext       *CT_OnOff       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main keepNext"`
	KeepLines      *CT_OnOff       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main keepLines"`
	PageBreakBefore *CT_OnOff      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pageBreakBefore"`
	WidowControl   *CT_OnOff       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main widowControl"`
	NumPr         *CT_NumPr        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numPr"`
	Spacing       *CT_Spacing      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main spacing"`
	Ind           *CT_Ind          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ind"`
	Jc            *CT_Jc           `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main jc"`
	Tabs          *CT_Tabs         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tabs"`
	SectPr        *CT_SectPr       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sectPr"`
	Shd           *CT_Shd          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main shd"`
	OutlineLvl    *CT_OutlineLvl   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main outlineLvl"`
	Raw           []xmlutil.RawXML `xml:",any"`
}

// CT_R is a run (text with formatting scope).
type CT_R struct {
	XMLName  xml.Name        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
	RPr      *CT_RPr         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
	FldChar  *CT_FldChar     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main fldChar"`
	InstrText *CT_InstrText  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main instrText"`
	T        *CT_Text        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main t"`
	Br       *CT_Br          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main br"`
	Tab      *CT_Tab         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tab"`
	Cr       *CT_Cr          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main cr"`
	Drawing  *CT_Drawing     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main drawing"`
	Raw      []xmlutil.RawXML `xml:",any"`
}

// CT_RPr holds run properties.
type CT_RPr struct {
	XMLName  xml.Name        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
	RStyle   *CT_RStyle      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rStyle"`
	RFonts   *CT_RFonts      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rFonts"`
	B        *CT_OnOff       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main b"`
	I        *CT_OnOff       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main i"`
	U        *CT_U           `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main u"`
	Sz       *CT_Sz          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sz"`
	SzCs     *CT_Sz          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main szCs"`
	Color    *CT_Color       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main color"`
	Highlight *CT_Highlight  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main highlight"`
	VertAlign *CT_VertAlign  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main vertAlign"`
	Lang     *CT_Lang        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lang"`
	NoProof  *CT_OnOff       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main noProof"`
	Strike   *CT_OnOff       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main strike"`
	DStrike  *CT_OnOff       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main dstrike"`
	Vanish   *CT_OnOff       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main vanish"`
	Kern     *CT_Kern        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main kern"`
	Shd      *CT_Shd         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main shd"`
	SmallCaps *CT_OnOff      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main smallCaps"`
	Caps     *CT_OnOff       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main caps"`
	Raw      []xmlutil.RawXML `xml:",any"`
}

// ---- Text / Special content ----

// CT_Text is a text element (w:t) with xml:space preservation (WML-03).
type CT_Text struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main t"`
	Space   *string  `xml:"http://www.w3.org/XML/1998/namespace space,attr"`
	Value   string   `xml:",chardata"`
}

// MarshalXML writes the text element, adding xml:space="preserve" when
// the value contains leading/trailing whitespace or double spaces.
func (t *CT_Text) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	needsSpace := (t.Space != nil && *t.Space == "preserve") ||
		(len(t.Value) > 0 && (t.Value[0] == ' ' || t.Value[len(t.Value)-1] == ' ' || strings.Contains(t.Value, "  ")))
	if needsSpace {
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Space: NSXML, Local: "space"},
			Value: "preserve",
		})
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if err := e.EncodeToken(xml.CharData(t.Value)); err != nil {
		return err
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// CT_Br is a line break.
type CT_Br struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main br"`
	Type    *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr,omitempty"`
}

// CT_Tab is a tab character.
type CT_Tab struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tab"`
}

// CT_Cr is a carriage return.
type CT_Cr struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main cr"`
}

// CT_SoftHyphen is an optional hyphen.
type CT_SoftHyphen struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main softHyphen"`
}

// CT_NoBreakHyphen is a non-breaking hyphen.
type CT_NoBreakHyphen struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main noBreakHyphen"`
}

// ---- Section properties ----

// CT_SectPr holds section properties.
type CT_SectPr struct {
	XMLName  xml.Name         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sectPr"`
	PgSz     *CT_PgSz        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pgSz"`
	PgMar    *CT_PgMar        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pgMar"`
	Cols     *CT_Cols         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main cols"`
	DocGrid  *CT_DocGrid      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main docGrid"`
	HdrFtrRef []*CT_HdrFtrRef `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main headerReference"`
	FtrRef   []*CT_HdrFtrRef  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main footerReference"`
	TitlePg  *CT_OnOff        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main titlePg"`
	Raw      []xmlutil.RawXML `xml:",any"`
}

// CT_PgSz is page size (w:h, w:w attributes).
type CT_PgSz struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pgSz"`
	W       *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main w,attr,omitempty"`
	H       *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main h,attr,omitempty"`
	Code    *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main code,attr,omitempty"`
}

// CT_PgMar is page margins.
type CT_PgMar struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pgMar"`
	Top     *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main top,attr,omitempty"`
	Right   *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main right,attr,omitempty"`
	Bottom  *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bottom,attr,omitempty"`
	Left    *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main left,attr,omitempty"`
	Header  *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main header,attr,omitempty"`
	Footer  *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main footer,attr,omitempty"`
}

// CT_Cols is column definitions.
type CT_Cols struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main cols"`
	Space   *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main space,attr,omitempty"`
	Num     *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main num,attr,omitempty"`
}

// CT_DocGrid is document grid settings.
type CT_DocGrid struct {
	XMLName  xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main docGrid"`
	Type     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr,omitempty"`
	LinePitch *int64  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main linePitch,attr,omitempty"`
}

// CT_HdrFtrRef references a header or footer part.
// The element name is determined by the parent struct field tag
// (headerReference for HdrFtrRef, footerReference for FtrRef).
type CT_HdrFtrRef struct {
	ID   string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
	Type string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr"`
}
