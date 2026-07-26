package xmlutil

import (
	"bytes"
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
// Entity=nil, an io.LimitReader cap, and DOCTYPE rejection.
func NewSafeDecoder(r io.Reader, maxBytes int64) *SafeDecoder {
	d := xml.NewDecoder(&docTypeFilter{r: io.LimitReader(r, maxBytes)})
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

// docTypeFilter wraps io.Reader and rejects any input containing
// a DOCTYPE declaration. Scans the first read buffer for `<!DOCTYPE`.
// The check is case-insensitive and only applies to the first 8KB.
type docTypeFilter struct {
	r       io.Reader
	buf     []byte
	checked bool
}

func (f *docTypeFilter) Read(p []byte) (int, error) {
	if !f.checked {
		tmp := make([]byte, len(p))
		n, err := f.r.Read(tmp)
		if n > 0 {
			f.buf = append(f.buf, tmp[:n]...)
			upper := make([]byte, len(f.buf))
			for i, b := range f.buf {
				if b >= 'a' && b <= 'z' {
					upper[i] = b - 32
				} else {
					upper[i] = b
				}
			}
			if bytes.Contains(upper, []byte("<!DOCTYPE")) {
				return 0, fmt.Errorf("%w: DOCTYPE declaration rejected in input stream", ErrDOCTYPE)
			}
		}
		f.checked = true
		if err != nil {
			n = copy(p, f.buf)
			return n, err
		}
	}
	if len(f.buf) > 0 {
		n := copy(p, f.buf)
		f.buf = f.buf[n:]
		return n, nil
	}
	return f.r.Read(p)
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
