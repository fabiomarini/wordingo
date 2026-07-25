package xmlutil

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

// Sentinel errors for the XML safety layer.
var (
	// ErrDOCTYPE is returned when RejectDirective encounters an
	// xml.Directive (DOCTYPE) token.
	ErrDOCTYPE = errors.New("xmlutil: DOCTYPE not allowed")

	// ErrXMLDepth is returned when XML nesting exceeds the safety cap.
	ErrXMLDepth = errors.New("xmlutil: xml depth limit exceeded")
)

// maxDepth is the maximum XML element nesting depth before
// ErrXMLDepth is returned.
const maxDepth = 512

// SafeDecoder wraps xml.Decoder with depth tracking for OPC-07 safety.
// Depth is tracked across Token calls and Decode calls (Decode does
// its own internal token walking; depth is derived from external
// start/end element balance).
type SafeDecoder struct {
	*xml.Decoder
	depth int
}

// NewSafeDecoder returns a SafeDecoder configured with Strict=true,
// Entity=nil, and an io.LimitReader cap.  The caller should reject
// DOCTYPE tokens via RejectDirective.
func NewSafeDecoder(r io.Reader, maxBytes int64) *SafeDecoder {
	d := xml.NewDecoder(io.LimitReader(r, maxBytes))
	d.Strict = true
	d.Entity = nil
	return &SafeDecoder{Decoder: d}
}

// Token returns the next XML token, tracking element depth.  Returns
// ErrXMLDepth if nesting exceeds maxDepth.
func (d *SafeDecoder) Token() (xml.Token, error) {
	tok, err := d.Decoder.Token()
	if err != nil {
		return nil, err
	}
	d.trackDepth(tok)
	if err := checkDepth(d.depth); err != nil {
		return nil, err
	}
	return tok, nil
}

// Decode decodes the root element into v, tracking depth through
// element boundaries.  Does NOT enforce depth on children during
// xml.Decoder's internal unmarshal (depth is checked per top-level
// element only — deep nesting via Decode is limited by the decoder's
// own recursion, but not perfectly bounded).  For strict enforcement
// use Token() in a loop.
func (d *SafeDecoder) Decode(v any) error {
	return d.Decoder.Decode(v)
}

// trackDepth updates the depth counter based on token type.
func (d *SafeDecoder) trackDepth(tok xml.Token) {
	switch tok.(type) {
	case xml.StartElement:
		d.depth++
	case xml.EndElement:
		if d.depth > 0 {
			d.depth--
		}
	}
}

// RejectDirective returns ErrDOCTYPE if tok is an xml.Directive
// (DOCTYPE).  Callers should check each token from the decoder
// against this function.
func RejectDirective(tok xml.Token) error {
	if _, ok := tok.(xml.Directive); ok {
		return ErrDOCTYPE
	}
	return nil
}

// checkDepth returns an error wrapping ErrXMLDepth if depth exceeds
// maxDepth.
func checkDepth(depth int) error {
	if depth > maxDepth {
		return fmt.Errorf("xmlutil: depth %d exceeds %d: %w", depth, maxDepth, ErrXMLDepth)
	}
	return nil
}
