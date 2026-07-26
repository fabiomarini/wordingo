package wml

import (
	"encoding/xml"

	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// ---- DrawingML types for image embedding (Phase 5) ----

// CT_Drawing is a drawing element wrapper (w:drawing).
type CT_Drawing struct {
	XMLName xml.Name    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main drawing"`
	Inline  *CT_Inline  `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing inline"`
	Anchor  *CT_Anchor  `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing anchor"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_Inline is an inline DrawingML object (wp:inline).
type CT_Inline struct {
	XMLName          xml.Name             `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing inline"`
	DistT            *int64               `xml:"distT,attr,omitempty"`
	DistB            *int64               `xml:"distB,attr,omitempty"`
	DistL            *int64               `xml:"distL,attr,omitempty"`
	DistR            *int64               `xml:"distR,attr,omitempty"`
	Extent           *CT_Extent           `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing extent"`
	EffectExtent     *CT_EffectExtent     `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing effectExtent"`
	DocPr            *CT_DocPr            `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing docPr"`
	CNvGraphicFramePr *CT_CNvGraphicFramePr `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing cNvGraphicFramePr"`
	Graphic          *CT_Graphic          `xml:"http://schemas.openxmlformats.org/drawingml/2006/main graphic"`
	Raw              []xmlutil.RawXML     `xml:",any"`
}

// CT_Anchor is an anchored DrawingML object (wp:anchor) — placeholder for future use.
type CT_Anchor struct {
	XMLName xml.Name        `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing anchor"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_Extent is the drawing object extent in EMU (wp:extent).
type CT_Extent struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing extent"`
	Cx      int64    `xml:"cx,attr"`
	Cy      int64    `xml:"cy,attr"`
}

// CT_EffectExtent is the effect extent (wp:effectExtent).
type CT_EffectExtent struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing effectExtent"`
	L       int64    `xml:"l,attr"`
	T       int64    `xml:"t,attr"`
	R       int64    `xml:"r,attr"`
	B       int64    `xml:"b,attr"`
}

// CT_DocPr is the document property (wp:docPr).
type CT_DocPr struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing docPr"`
	ID      int64    `xml:"id,attr"`
	Name    string   `xml:"name,attr"`
}

// CT_CNvGraphicFramePr is non-visual graphic frame properties (wp:cNvGraphicFramePr).
type CT_CNvGraphicFramePr struct {
	XMLName            xml.Name             `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing cNvGraphicFramePr"`
	GraphicFrameLocks *CT_GraphicFrameLocks `xml:"http://schemas.openxmlformats.org/drawingml/2006/main graphicFrameLocks"`
}

// CT_GraphicFrameLocks is graphic frame locks (a:graphicFrameLocks).
type CT_GraphicFrameLocks struct {
	XMLName         xml.Name `xml:"http://schemas.openxmlformats.org/drawingml/2006/main graphicFrameLocks"`
	NoChangeAspect *bool     `xml:"noChangeAspect,attr,omitempty"`
}

// ---- Graphic container ----

// CT_Graphic is the graphic object wrapper (a:graphic).
type CT_Graphic struct {
	XMLName     xml.Name        `xml:"http://schemas.openxmlformats.org/drawingml/2006/main graphic"`
	GraphicData *CT_GraphicData `xml:"http://schemas.openxmlformats.org/drawingml/2006/main graphicData"`
	Raw         []xmlutil.RawXML `xml:",any"`
}

// CT_GraphicData is the graphic data payload (a:graphicData).
type CT_GraphicData struct {
	XMLName xml.Name        `xml:"http://schemas.openxmlformats.org/drawingml/2006/main graphicData"`
	Uri     string          `xml:"uri,attr"`
	Pic     *CT_Pic         `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture pic"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// ---- Picture (pic:) types ----

// CT_Pic is a picture element (pic:pic).
type CT_Pic struct {
	XMLName        xml.Name            `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture pic"`
	NvPicPr        *CT_NonVisualPicProps `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture nvPicPr"`
	BlipFill       *CT_BlipFill        `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture blipFill"`
	SpPr           *CT_SpPr            `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture spPr"`
	Raw            []xmlutil.RawXML    `xml:",any"`
}

// CT_NonVisualPicProps is non-visual picture properties (pic:nvPicPr).
type CT_NonVisualPicProps struct {
	XMLName  xml.Name      `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture nvPicPr"`
	CNvPr    *CT_CNvPr     `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture cNvPr"`
	CNvPicPr *CT_CNvPicPr  `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture cNvPicPr"`
	Raw      []xmlutil.RawXML `xml:",any"`
}

// CT_CNvPr is non-visual drawing properties (pic:cNvPr).
type CT_CNvPr struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture cNvPr"`
	ID      int64    `xml:"id,attr"`
	Name    string   `xml:"name,attr"`
}

// CT_CNvPicPr is non-visual picture drawing properties (pic:cNvPicPr).
type CT_CNvPicPr struct {
	XMLName     xml.Name        `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture cNvPicPr"`
	PicLocks    *CT_PicLocks    `xml:"http://schemas.openxmlformats.org/drawingml/2006/main picLocks"`
	PreferRel   *bool           `xml:"preferRelativeResize,attr,omitempty"`
	Raw         []xmlutil.RawXML `xml:",any"`
}

// CT_PicLocks is picture locks (a:picLocks).
type CT_PicLocks struct {
	XMLName         xml.Name `xml:"http://schemas.openxmlformats.org/drawingml/2006/main picLocks"`
	NoChangeAspect *bool     `xml:"noChangeAspect,attr,omitempty"`
}

// ---- Blip fill ----

// CT_BlipFill is a blip fill (pic:blipFill).
type CT_BlipFill struct {
	XMLName xml.Name        `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture blipFill"`
	Blip    *CT_Blip        `xml:"http://schemas.openxmlformats.org/drawingml/2006/main blip"`
	Stretch *CT_Stretch     `xml:"http://schemas.openxmlformats.org/drawingml/2006/main stretch"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_Blip is a picture reference (a:blip) with r:embed relationship.
type CT_Blip struct {
	XMLName xml.Name        `xml:"http://schemas.openxmlformats.org/drawingml/2006/main blip"`
	Embed   string          `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships embed,attr"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_Stretch is a stretch fill (a:stretch).
type CT_Stretch struct {
	XMLName  xml.Name        `xml:"http://schemas.openxmlformats.org/drawingml/2006/main stretch"`
	FillRect *CT_FillRect    `xml:"http://schemas.openxmlformats.org/drawingml/2006/main fillRect"`
	Raw      []xmlutil.RawXML `xml:",any"`
}

// CT_FillRect is a fill rectangle (a:fillRect).
type CT_FillRect struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/drawingml/2006/main fillRect"`
}

// ---- Shape properties ----

// CT_SpPr is shape properties (pic:spPr).
type CT_SpPr struct {
	XMLName     xml.Name          `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture spPr"`
	Xfrm        *CT_Xfrm          `xml:"http://schemas.openxmlformats.org/drawingml/2006/main xfrm"`
	PrstGeom    *CT_PresetGeometry `xml:"http://schemas.openxmlformats.org/drawingml/2006/main prstGeom"`
	Raw         []xmlutil.RawXML  `xml:",any"`
}

// CT_Xfrm is a 2D transform (a:xfrm).
type CT_Xfrm struct {
	XMLName xml.Name            `xml:"http://schemas.openxmlformats.org/drawingml/2006/main xfrm"`
	Off     *CT_Point2D         `xml:"http://schemas.openxmlformats.org/drawingml/2006/main off"`
	Ext     *CT_PositiveSize2D  `xml:"http://schemas.openxmlformats.org/drawingml/2006/main ext"`
	Raw     []xmlutil.RawXML    `xml:",any"`
}

// CT_Point2D is a 2D point (a:off).
type CT_Point2D struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/drawingml/2006/main off"`
	X       int64    `xml:"x,attr"`
	Y       int64    `xml:"y,attr"`
}

// CT_PositiveSize2D is a positive size 2D (a:ext).
type CT_PositiveSize2D struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/drawingml/2006/main ext"`
	Cx      int64    `xml:"cx,attr"`
	Cy      int64    `xml:"cy,attr"`
}

// CT_PresetGeometry is a preset geometry (a:prstGeom).
type CT_PresetGeometry struct {
	XMLName xml.Name        `xml:"http://schemas.openxmlformats.org/drawingml/2006/main prstGeom"`
	Prst    string          `xml:"prst,attr"`
	AvLst   *CT_AvLst       `xml:"http://schemas.openxmlformats.org/drawingml/2006/main avLst"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

// CT_AvLst is a list of shape adjust values (a:avLst).
type CT_AvLst struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/drawingml/2006/main avLst"`
}
