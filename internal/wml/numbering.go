package wml

import (
	"encoding/xml"

	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// CT_Numbering is the root element of word/numbering.xml.
type CT_Numbering struct {
	XMLName     xml.Name          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numbering"`
	Num         []*CT_Num         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main num"`
	AbstractNum []*CT_AbstractNum `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main abstractNum"`
	Raw         []xmlutil.RawXML  `xml:",any"`
}

// CT_AbstractNum is an abstract numbering definition.
type CT_AbstractNum struct {
	XMLName       xml.Name        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main abstractNum"`
	AbstractNumID *int64          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main abstractNumId,attr,omitempty"`
	Lvl           []*CT_Lvl       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lvl"`
	Raw           []xmlutil.RawXML `xml:",any"`
}

// CT_Num is a numbering instance (w:num).
//
// LvlOverride captures w:num/w:lvlOverride (ISO §17.9.18) — overrides
// abstract numbering level values for this instance.  Added per RESEARCH
// Pitfall 4; backward compatible (previously hoarded into Raw).
type CT_Num struct {
	XMLName       xml.Name          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main num"`
	NumID         *int64            `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numId,attr,omitempty"`
	AbstractNumID *CT_AbstractNumID `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main abstractNumId"`
	LvlOverride   []*CT_LvlOverride `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lvlOverride"`
	Raw           []xmlutil.RawXML  `xml:",any"`
}

// CT_LvlOverride overrides abstract numbering level values for a
// specific ilvl in a numbering instance (w:lvlOverride, ISO §17.9.18).
type CT_LvlOverride struct {
	XMLName       xml.Name          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lvlOverride"`
	ILvl          *int64            `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ilvl,attr,omitempty"`
	StartOverride *CT_StartOverride `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main startOverride"`
	Lvl           *CT_Lvl           `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lvl,omitempty"`
}

// CT_StartOverride overrides the starting number for a level
// (w:startOverride, ISO §17.9.24).
type CT_StartOverride struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main startOverride"`
	Val     *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_AbstractNumID references an abstract numbering definition.
type CT_AbstractNumID struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main abstractNumId"`
	Val     *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_Lvl is a numbering level definition.
type CT_Lvl struct {
	XMLName xml.Name    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lvl"`
	ILvl    *int64      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ilvl,attr,omitempty"`
	NumFmt  *CT_NumFmt  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numFmt"`
	LvlText *CT_LvlText `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lvlText"`
	Start   *CT_Start   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main start"`
	PPr     *CT_PPr     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
	RPr     *CT_RPr     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_NumFmt is a numbering format (decimal, upperRoman, bullet, etc.).
type CT_NumFmt struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numFmt"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_LvlText is the text for a numbering level ("%1.", "%2)", etc.).
type CT_LvlText struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lvlText"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_Start is the starting number for a level.
type CT_Start struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main start"`
	Val     *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}


