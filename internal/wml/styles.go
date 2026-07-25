package wml

import (
	"encoding/xml"

	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// CT_Settings is the root element of word/settings.xml.
type CT_Settings struct {
	XMLName        xml.Name         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main settings"`
	Zoom           *CT_Zoom         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main zoom"`
	DefaultTabStop *CT_DefaultTabStop `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main defaultTabStop"`
	Compat         *CT_Compat       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main compat"`
	Raw            []xmlutil.RawXML `xml:",any"`
}

// CT_Zoom is the zoom setting.
type CT_Zoom struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main zoom"`
	Percent *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main percent,attr,omitempty"`
}

// CT_DefaultTabStop is the default tab stop interval.
type CT_DefaultTabStop struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main defaultTabStop"`
	Val     *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_Compat holds compatibility settings.
type CT_Compat struct {
	XMLName xml.Name        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main compat"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_Styles is the root element of word/styles.xml.
type CT_Styles struct {
	XMLName    xml.Name         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main styles"`
	DocDefaults *CT_DocDefaults `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main docDefaults"`
	LatentStyles *CT_LatentStyles `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main latentStyles"`
	Style      []*CT_Style      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main style"`
	Raw        []xmlutil.RawXML `xml:",any"`
}

// CT_Style is a single style definition.
type CT_Style struct {
	XMLName  xml.Name        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main style"`
	Type     *string         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr,omitempty"`
	StyleID  *string         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main styleId,attr,omitempty"`
	Default  *string         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main default,attr,omitempty"`
	CustomStyle *string      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main customStyle,attr,omitempty"`
	Name     *CT_StyleName   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main name"`
	BasedOn  *CT_BasedOn     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main basedOn"`
	Next     *CT_Next        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main next"`
	Link     *CT_Link        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main link"`
	QFormat  *CT_OnOff       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main qFormat"`
	PPr      *CT_PPr         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
	RPr      *CT_RPr         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
	Raw      []xmlutil.RawXML `xml:",any"`
}

// CT_StyleName is the friendly name of a style (w:name).
type CT_StyleName struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main name"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_BasedOn references a style this one inherits from.
type CT_BasedOn struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main basedOn"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_Next references the next paragraph style.
type CT_Next struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main next"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_Link references a linked character style.
type CT_Link struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main link"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}

// CT_DocDefaults holds document-level default run/paragraph properties.
type CT_DocDefaults struct {
	XMLName  xml.Name      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main docDefaults"`
	RPrDefault *CT_RPrDefault `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPrDefault"`
	PPrDefault *CT_PPrDefault `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPrDefault"`
	Raw      []xmlutil.RawXML `xml:",any"`
}

// CT_RPrDefault wraps default run properties.
type CT_RPrDefault struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPrDefault"`
	RPr     *CT_RPr  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_PPrDefault wraps default paragraph properties.
type CT_PPrDefault struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPrDefault"`
	PPr     *CT_PPr  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_LatentStyles holds latent (default) style definitions.
type CT_LatentStyles struct {
	XMLName xml.Name           `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main latentStyles"`
	LsdException []*CT_LsdException `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lsdException"`
	Raw      []xmlutil.RawXML    `xml:",any"`
	DefLockedState  *string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main defLockedState,attr,omitempty"`
	DefUIPriority   *int64  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main defUIPriority,attr,omitempty"`
	DefSemiHidden   *string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main defSemiHidden,attr,omitempty"`
	DefUnhideWhenUsed *string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main defUnhideWhenUsed,attr,omitempty"`
	DefQFormat      *string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main defQFormat,attr,omitempty"`
}

// CT_LsdException is a latent style exception entry.
type CT_LsdException struct {
	XMLName       xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lsdException"`
	Name          *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main name,attr,omitempty"`
	Locked        *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main locked,attr,omitempty"`
	UIPriority    *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main uiPriority,attr,omitempty"`
	SemiHidden    *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main semiHidden,attr,omitempty"`
	UnhideWhenUsed *string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main unhideWhenUsed,attr,omitempty"`
	QFormat       *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main qFormat,attr,omitempty"`
}


