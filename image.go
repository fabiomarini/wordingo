package wordingo

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
	"math"
	"os"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
)

// EMU conversion constants.
const (
	emusPerInch = 914400
	emusPerCm   = 360000
)

// inchesToEMU converts inches to EMU (English Metric Units).
func inchesToEMU(inches float64) int64 {
	return int64(math.Round(inches * emusPerInch))
}

// defaultImageWidth is the default auto-size width for images when DPI
// detection fails or image dimensions cannot be determined.
const defaultImageWidthInches = 3.0

// ---- DPI detection ----

// jpegDPI parses the JFIF APP0 or EXIF segment for DPI info from a
// JPEG byte stream. Returns the horizontal DPI, or 72 if unknown.
func jpegDPI(data []byte) int {
	// Look for JFIF APP0 marker: FF E0 <size> <4A 46 49 46 00>
	if len(data) < 14 {
		return 72
	}
	if data[0] != 0xFF || data[1] != 0xD8 {
		return 72
	}
	// Scan APP markers for JFIF (FF E0) or EXIF (FF E1)
	for i := 2; i < len(data)-5; {
		if data[i] != 0xFF {
			break
		}
		marker := data[i+1]
		if marker == 0 { // padding
			i++
			continue
		}
		if i+3 >= len(data) {
			break
		}
		segLen := int(data[i+2])<<8 | int(data[i+3])
		if marker == 0xE0 && segLen >= 7 && i+9 < len(data) {
			// JFIF: FF E0 <len> "JFIF\x00" <ver-major> <ver-minor> <density-units> <Xdpi> <Ydpi>
			if string(data[i+4:i+8]) == "JFIF" {
				units := data[i+9]
				xDpi := int(data[i+10])<<8 | int(data[i+11])
				if units == 0 { // aspect ratio only
					return 72
				}
				if xDpi == 0 {
					return 72
				}
				return xDpi
			}
		}
		i += segLen + 2
	}
	return 72
}

// pngDPI parses the pHYs chunk from PNG data for DPI info.
// Returns the horizontal DPI, or 72 if unknown.
func pngDPI(data []byte) int {
	// pHYs chunk: "pHYs" <4B pixels-per-unit-x> <4B pixels-per-unit-y> <1B unit>
	if len(data) < 33 {
		return 72
	}
	// Find the IHDR first (must appear first in PNG)
	// Then scan for pHYs
	for i := 8; i < len(data)-12; {
		chunkLen := int(data[i])<<24 | int(data[i+1])<<16 | int(data[i+2])<<8 | int(data[i+3])
		chunkType := string(data[i+4 : i+8])
		if chunkType == "pHYs" && chunkLen >= 9 && i+17 < len(data) {
			ppuX := int(data[i+8])<<24 | int(data[i+9])<<16 | int(data[i+10])<<8 | int(data[i+11])
			unit := data[i+16]
			if unit == 1 { // meter
				// Convert pixels-per-meter to DPI
				return int(math.Round(float64(ppuX) * 0.0254))
			}
			// unit == 0 means unknown, default
			return 72
		}
		i += 4 + chunkLen + 4
		if chunkType == "IEND" {
			break
		}
	}
	return 72
}

// detectImageDPI reads image magic bytes and returns the detected
// DPI. Falls back to 72 for unknown formats.
func detectImageDPI(data []byte) int {
	if len(data) < 8 {
		return 72
	}
	// PNG magic: 89 50 4E 47 0D 0A 1A 0A
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return pngDPI(data)
	}
	// JPEG magic: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 {
		return jpegDPI(data)
	}
	return 72
}

// detectContentType returns the MIME type based on image magic bytes,
// or "" if unknown.
func detectContentType(data []byte) string {
	if len(data) < 8 {
		return ""
	}
	// PNG
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return ctPng
	}
	// JPEG
	if data[0] == 0xFF && data[1] == 0xD8 {
		return ctJpeg
	}
	return ""
}

// detectImageExt returns the file extension for an image based on magic
// bytes.
func detectImageExt(data []byte) string {
	if len(data) < 8 {
		return "bin"
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png"
	}
	if data[0] == 0xFF && data[1] == 0xD8 {
		return "jpg"
	}
	return "bin"
}

// ---- Image dimension helpers ----

// imageDimensions returns the pixel dimensions of a PNG or JPEG image.
func imageDimensions(data []byte) (width, height int, err error) {
	// Try PNG first
	if len(data) > 8 && data[0] == 0x89 && data[1] == 0x50 {
		img, err := png.DecodeConfig(bytes.NewReader(data))
		if err == nil {
			return img.Width, img.Height, nil
		}
	}
	// Try JPEG
	if len(data) > 2 && data[0] == 0xFF && data[1] == 0xD8 {
		img, err := jpeg.DecodeConfig(bytes.NewReader(data))
		if err == nil {
			return img.Width, img.Height, nil
		}
	}
	return 0, 0, fmt.Errorf("wordingo: unknown image format or corrupt data")
}

// ---- Document image methods ----

// AddImage reads an image file from path, embeds it as a DrawingML
// inline image, and returns the run containing the drawing. The image
// is added to the last paragraph in the document.
func (d *Document) AddImage(path string) (*Run, error) {
	if d == nil {
		panic("wordingo: AddImage called on nil Document")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("wordingo: read image %s: %w", path, err)
	}
	ct := detectContentType(data)
	if ct == "" {
		return nil, fmt.Errorf("wordingo: unsupported image format in %s", path)
	}
	return d.AddImageBytes(path, data, ct)
}

// AddImageBytes embeds image data as a DrawingML inline image and
// returns the run containing the drawing. The image is added to the
// last paragraph in the document.
//
// name is used as a display name (not a file path). data is the raw
// image bytes. ct is the MIME content type (e.g. "image/png").
//
// The image is stored as a media part in word/media/imageN.ext with
// the appropriate relationship and content type override.
func (d *Document) AddImageBytes(name string, data []byte, ct string) (*Run, error) {
	if d == nil {
		panic("wordingo: AddImageBytes called on nil Document")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("wordingo: AddImageBytes: empty image data")
	}

	// Determine extension
	ext := detectImageExt(data)

	// Build media path
	imageID := d.nextImageID
	d.nextImageID++
	mediaPath := fmt.Sprintf("word/media/image%d.%s", imageID, ext)

	// Add media part
	d.pkg.MarkModified(mediaPath, data)

	// Add content type override
	d.pkg.ContentTypes.Overrides["/"+mediaPath] = ct

	// Add relationship from document.xml to media
	rels := d.pkg.Rels["word/document.xml"]
	if rels == nil {
		rels = &opc.Relationships{}
		d.pkg.Rels["word/document.xml"] = rels
	}
	relID := rels.NextRID()
	rels.Rels = append(rels.Rels, opc.Relationship{
		ID:     relID,
		Type:   relImage,
		Target: "media/" + imageFileName(imageID, ext),
	})

	// Compute default size
	widthEMU, heightEMU := d.computeImageSize(data)
	dpi := detectImageDPI(data)
	if dpi > 0 {
		// Recompute size based on DPI if pixel dimensions available
		if pw, ph, err := imageDimensions(data); err == nil && pw > 0 && ph > 0 {
			widthEMU = inchesToEMU(float64(pw) / float64(dpi))
			heightEMU = inchesToEMU(float64(ph) / float64(dpi))
		}
	}

	// Cap at defaultImageWidth and scale height proportionally
	if widthEMU > inchesToEMU(defaultImageWidthInches) {
		ratio := float64(inchesToEMU(defaultImageWidthInches)) / float64(widthEMU)
		widthEMU = inchesToEMU(defaultImageWidthInches)
		heightEMU = int64(math.Round(float64(heightEMU) * ratio))
	}
	if widthEMU <= 0 {
		widthEMU = inchesToEMU(defaultImageWidthInches)
	}
	if heightEMU <= 0 {
		heightEMU = inchesToEMU(defaultImageWidthInches)
	}

	// Build DrawingML inline structure
	drawing := buildImageDrawing(relID, imageID, name, widthEMU, heightEMU)

	// Create run with drawing
	r := &wml.CT_R{Drawing: drawing}

	// Append to last paragraph
	if d.doc.Body == nil {
		d.doc.Body = &wml.CT_Body{SectPr: defaultSectPr()}
	}
	if len(d.doc.Body.P) == 0 {
		d.doc.Body.AppendP(&wml.CT_P{})
	}
	lastP := d.doc.Body.P[len(d.doc.Body.P)-1]
	lastP.R = append(lastP.R, r)

	// Ensure at least a Tbl slice for body
	if d.doc.Body.Tbl == nil {
		d.doc.Body.Tbl = []*wml.CT_Tbl{}
	}

	d.dirty = true
	return &Run{ct: r, para: &Paragraph{ct: lastP, doc: d}}, nil
}

// computeImageSize returns default EMU dimensions for an image.
// Attempts pixel-based sizing at 72 DPI fallback.
func (d *Document) computeImageSize(data []byte) (widthEMU, heightEMU int64) {
	if pw, ph, err := imageDimensions(data); err == nil && pw > 0 && ph > 0 {
		// Assume 72 DPI as fallback
		widthEMU = inchesToEMU(float64(pw) / 72.0)
		heightEMU = inchesToEMU(float64(ph) / 72.0)
		return
	}
	return inchesToEMU(defaultImageWidthInches), inchesToEMU(defaultImageWidthInches)
}

// buildImageDrawing constructs the full DrawingML inline element for an image.
func buildImageDrawing(relID string, imageID int64, name string, cx, cy int64) *wml.CT_Drawing {
	prst := "rect"

	return &wml.CT_Drawing{
		Inline: &wml.CT_Inline{
			DistT: ptrInt64(0),
			DistB: ptrInt64(0),
			DistL: ptrInt64(0),
			DistR: ptrInt64(0),
			Extent: &wml.CT_Extent{
				Cx: cx,
				Cy: cy,
			},
			EffectExtent: &wml.CT_EffectExtent{
				L: 0,
				T: 0,
				R: 0,
				B: 0,
			},
			DocPr: &wml.CT_DocPr{
				ID:   imageID,
				Name: name,
			},
			CNvGraphicFramePr: &wml.CT_CNvGraphicFramePr{
				GraphicFrameLocks: &wml.CT_GraphicFrameLocks{
					NoChangeAspect: ptrBool(true),
				},
			},
			Graphic: &wml.CT_Graphic{
				GraphicData: &wml.CT_GraphicData{
					Uri: "http://schemas.openxmlformats.org/drawingml/2006/picture",
					Pic: &wml.CT_Pic{
						NvPicPr: &wml.CT_NonVisualPicProps{
							CNvPr: &wml.CT_CNvPr{
								ID:   imageID,
								Name: name,
							},
							CNvPicPr: &wml.CT_CNvPicPr{
								PicLocks: &wml.CT_PicLocks{
									NoChangeAspect: ptrBool(true),
								},
								PreferRel: ptrBool(true),
							},
						},
						BlipFill: &wml.CT_BlipFill{
							Blip: &wml.CT_Blip{
								Embed: relID,
							},
							Stretch: &wml.CT_Stretch{
								FillRect: &wml.CT_FillRect{},
							},
						},
						SpPr: &wml.CT_SpPr{
							Xfrm: &wml.CT_Xfrm{
								Off: &wml.CT_Point2D{
									X: 0,
									Y: 0,
								},
								Ext: &wml.CT_PositiveSize2D{
									Cx: cx,
									Cy: cy,
								},
							},
							PrstGeom: &wml.CT_PresetGeometry{
								Prst:  prst,
								AvLst: &wml.CT_AvLst{},
							},
						},
					},
				},
			},
		},
	}
}

// imageFileName returns the media file name for an image with the given
// ID and extension (e.g. image1.png).
func imageFileName(id int64, ext string) string {
	return fmt.Sprintf("image%d.%s", id, ext)
}

// ---- Run image sizing ----

// SetImageWidth sets the image display width in inches. Mutates the
// DrawingML inline extent on the run.
func (r *Run) SetImageWidth(inches float64) *Run {
	if r == nil {
		panic("wordingo: SetImageWidth called on nil Run")
	}
	if r.ct.Drawing != nil && r.ct.Drawing.Inline != nil {
		cx := inchesToEMU(inches)
		r.ct.Drawing.Inline.Extent.Cx = cx
		if r.ct.Drawing.Inline.Graphic != nil &&
			r.ct.Drawing.Inline.Graphic.GraphicData != nil &&
			r.ct.Drawing.Inline.Graphic.GraphicData.Pic != nil &&
			r.ct.Drawing.Inline.Graphic.GraphicData.Pic.SpPr != nil &&
			r.ct.Drawing.Inline.Graphic.GraphicData.Pic.SpPr.Xfrm != nil {
			r.ct.Drawing.Inline.Graphic.GraphicData.Pic.SpPr.Xfrm.Ext.Cx = cx
		}
		r.para.doc.dirty = true
	}
	return r
}

// SetImageHeight sets the image display height in inches. Mutates the
// DrawingML inline extent on the run.
func (r *Run) SetImageHeight(inches float64) *Run {
	if r == nil {
		panic("wordingo: SetImageHeight called on nil Run")
	}
	if r.ct.Drawing != nil && r.ct.Drawing.Inline != nil {
		cy := inchesToEMU(inches)
		r.ct.Drawing.Inline.Extent.Cy = cy
		if r.ct.Drawing.Inline.Graphic != nil &&
			r.ct.Drawing.Inline.Graphic.GraphicData != nil &&
			r.ct.Drawing.Inline.Graphic.GraphicData.Pic != nil &&
			r.ct.Drawing.Inline.Graphic.GraphicData.Pic.SpPr != nil &&
			r.ct.Drawing.Inline.Graphic.GraphicData.Pic.SpPr.Xfrm != nil {
			r.ct.Drawing.Inline.Graphic.GraphicData.Pic.SpPr.Xfrm.Ext.Cy = cy
		}
		r.para.doc.dirty = true
	}
	return r
}

// ptrBool is a helper to create *bool.
func ptrBool(v bool) *bool {
	return &v
}
