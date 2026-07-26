package wml

import (
	"encoding/xml"

	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// ---- Paragraph properties ----

// CT_PStyle is a paragraph style reference (w:pStyle).
type CT_PStyle struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pStyle"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_Jc is paragraph alignment (w:jc).
type CT_Jc struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main jc"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_Spacing is paragraph spacing (w:spacing).
type CT_Spacing struct {
	XMLName  xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main spacing"`
	Before   *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main before,attr,omitempty"`
	After    *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main after,attr,omitempty"`
	Line     *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main line,attr,omitempty"`
	LineRule *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lineRule,attr,omitempty"`
	BeforeLines *int64 `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main beforeLines,attr,omitempty"`
	AfterLines  *int64 `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main afterLines,attr,omitempty"`
}

// CT_Ind is paragraph indentation (w:ind).
type CT_Ind struct {
	XMLName  xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ind"`
	Left     *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main left,attr,omitempty"`
	Right    *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main right,attr,omitempty"`
	FirstLine *int64  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main firstLine,attr,omitempty"`
	Hanging  *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hanging,attr,omitempty"`
}

// CT_NumPr is a numbering properties group (w:numPr).
type CT_NumPr struct {
	XMLName xml.Name        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numPr"`
	ILvl    *CT_ILvl        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ilvl"`
	NumId   *CT_NumId       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numId"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_NumId references a numbering definition instance.
type CT_NumId struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numId"`
	Val     *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_ILvl references a numbering level.
type CT_ILvl struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ilvl"`
	Val     *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_Tabs holds tab stop definitions.
type CT_Tabs struct {
	XMLName xml.Name       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tabs"`
	Tab     []*CT_TabStop  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tab"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_TabStop is a single tab stop.
type CT_TabStop struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tab"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
	Pos     *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pos,attr,omitempty"`
	Leader  *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main leader,attr,omitempty"`
}

// ---- On/Off types ----

// CT_OnOff is a generic on/off toggle (w:b, w:i, w:keepNext, etc.).
// When val is nil the element is equivalent to "true" per OOXML.
// NOTE: No XMLName — the containing struct's field tag provides the
// element name (e.g. "keepNext", "qFormat", "pageBreakBefore").
type CT_OnOff struct {
	Val *bool `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// ---- Shading ----

// CT_Shd is shading (w:shd) for paragraphs, runs, table cells.
type CT_Shd struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main shd"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
	Color   *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main color,attr,omitempty"`
	Fill    *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main fill,attr,omitempty"`
	ThemeFill *string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main themeFill,attr,omitempty"`
	ThemeColor *string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main themeColor,attr,omitempty"`
}

// ---- Run properties ----

// CT_RStyle is a run style reference.
type CT_RStyle struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rStyle"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_RFonts is a run font declaration.
type CT_RFonts struct {
	XMLName    xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rFonts"`
	Ascii      *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ascii,attr,omitempty"`
	HAnsi      *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hAnsi,attr,omitempty"`
	EastAsia   *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main eastAsia,attr,omitempty"`
	CS         *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main cs,attr,omitempty"`
	AsciiTheme *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main asciiTheme,attr,omitempty"`
	HAnsiTheme *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hAnsiTheme,attr,omitempty"`
}

// CT_U is underline.
type CT_U struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main u"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
	Color   *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main color,attr,omitempty"`
}

// CT_Sz is font size in half-points (w:sz, w:szCs).
// NOTE: No XMLName — the containing CT_RPr's field tag provides the
// element name ("sz" or "szCs").
type CT_Sz struct {
	Val *int64 `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_Color is a color value.
type CT_Color struct {
	XMLName    xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main color"`
	Val        *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
	ThemeColor *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main themeColor,attr,omitempty"`
	ThemeShade *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main themeShade,attr,omitempty"`
	ThemeTint  *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main themeTint,attr,omitempty"`
}

// CT_Highlight is highlight (text background).
type CT_Highlight struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main highlight"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_VertAlign is vertical alignment (superscript/subscript).
type CT_VertAlign struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main vertAlign"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_Lang is language declaration.
type CT_Lang struct {
	XMLName  xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lang"`
	Val      *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
	EastAsia *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main eastAsia,attr,omitempty"`
}

// CT_Kern is font kerning.
type CT_Kern struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main kern"`
	Val     *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_OutlineLvl is heading outline level (w:outlineLvl).
// No XMLName — field tag on CT_PPr provides element name.
type CT_OutlineLvl struct {
	Val *int64 `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}


