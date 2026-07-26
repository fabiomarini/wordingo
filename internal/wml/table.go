package wml

import (
	"encoding/xml"

	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// CT_Tbl is a table.
type CT_Tbl struct {
	XMLName xml.Name    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tbl"`
	TblPr   *CT_TblPr   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblPr"`
	TblGrid *CT_TblGrid `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblGrid"`
	Tr      []*CT_Tr    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tr"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_TblPr holds table-level properties.
type CT_TblPr struct {
	XMLName  xml.Name        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblPr"`
	TblStyle *CT_TblStyle    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblStyle"`
	TblW     *CT_TblW        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblW"`
	Borders  *CT_TblBorders  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblBorders"`
	Shd      *CT_Shd         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main shd"`
	Raw      []xmlutil.RawXML `xml:",any"`
}

// CT_TblStyle is a table style reference.
type CT_TblStyle struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblStyle"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_TblW is table / cell width.
// NOTE: No XMLName — the containing struct's field tag provides the
// element name ("tblW" or "tcW").
type CT_TblW struct {
	W    *int64  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main w,attr,omitempty"`
	Type *string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr,omitempty"`
}

// CT_TblBorders holds table border definitions.
type CT_TblBorders struct {
	XMLName xml.Name       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblBorders"`
	Top     *CT_TblBorder  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main top"`
	Left    *CT_TblBorder  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main left"`
	Bottom  *CT_TblBorder  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bottom"`
	Right   *CT_TblBorder  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main right"`
	InsideH *CT_TblBorder  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main insideH"`
	InsideV *CT_TblBorder  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main insideV"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_TblBorder is a single table border.
// NOTE: No XMLName — the containing CT_TblBorders's field tag provides
// the element name (e.g. "top", "left", "bottom").
type CT_TblBorder struct {
	Val        *string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
	Sz         *int64  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sz,attr,omitempty"`
	Color      *string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main color,attr,omitempty"`
	Space      *int64  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main space,attr,omitempty"`
	ThemeColor *string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main themeColor,attr,omitempty"`
}

// CT_TblGrid holds column definitions.
type CT_TblGrid struct {
	XMLName xml.Name       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblGrid"`
	GridCol []*CT_GridCol  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main gridCol"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_GridCol is a single grid column.
type CT_GridCol struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main gridCol"`
	W       *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main w,attr,omitempty"`
}

// CT_Tr is a table row.
type CT_Tr struct {
	XMLName xml.Name    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tr"`
	TrPr    *CT_TrPr    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main trPr"`
	Tc      []*CT_Tc    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tc"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_TrPr holds table row properties.
type CT_TrPr struct {
	XMLName xml.Name        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main trPr"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_Tc is a table cell.
type CT_Tc struct {
	XMLName xml.Name    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tc"`
	TcPr    *CT_TcPr    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tcPr"`
	P       []*CT_P     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_TcPr holds table cell properties.
type CT_TcPr struct {
	XMLName  xml.Name        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tcPr"`
	GridSpan *CT_GridSpan    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main gridSpan"`
	VMerge   *CT_VMerge      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main vMerge"`
	TcW      *CT_TblW        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tcW"`
	Shd      *CT_Shd         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main shd"`
	Borders  *CT_TcBorders   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tcBorders"`
	Raw      []xmlutil.RawXML `xml:",any"`
}

// CT_TcBorders holds table cell border definitions.
type CT_TcBorders struct {
	XMLName xml.Name       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tcBorders"`
	Top     *CT_TblBorder  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main top"`
	Left    *CT_TblBorder  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main left"`
	Bottom  *CT_TblBorder  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bottom"`
	Right   *CT_TblBorder  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main right"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_GridSpan is horizontal cell merge width.
type CT_GridSpan struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main gridSpan"`
	Val     *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_VMerge is vertical cell merge.
type CT_VMerge struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main vMerge"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}
