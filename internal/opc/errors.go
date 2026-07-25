package opc

import "errors"

// Sentinel errors for the OPC package layer (D-11). Callers match with
// errors.Is; all returned errors wrap one of these via fmt.Errorf("...: %w").
var (
	// ErrInvalidPackage marks a structurally invalid OPC package
	// (missing content types, dangling relationship targets, etc.).
	ErrInvalidPackage = errors.New("opc: invalid package")

	// ErrUnsafePath marks a ZIP entry name or relationship target that
	// escapes the package (absolute paths, "..", backslashes, drive
	// letters, empty segments, non-UTF-8).
	ErrUnsafePath = errors.New("opc: unsafe path")

	// ErrDecompressionLimit marks a part or package exceeding the
	// decompressed-size or compression-ratio safety caps (zip bombs).
	ErrDecompressionLimit = errors.New("opc: decompression limit exceeded")

	// ErrTooManyParts marks a package with more than MaxParts entries.
	ErrTooManyParts = errors.New("opc: too many parts")

	// ErrXMLDepth marks XML nesting beyond the safety cap. Enforced by
	// internal/xmlutil; declared here so the taxonomy lives in one place.
	ErrXMLDepth = errors.New("opc: xml depth limit exceeded")
)

// Safety limits (OPC-07). Values per 01-RESEARCH.md §Safety Limits.
const (
	// MaxParts caps the number of ZIP entries parsed.
	MaxParts = 4096

	// MaxPartBytes caps a single part's decompressed size (128 MiB).
	MaxPartBytes = 128 << 20

	// MaxTotalBytes caps total decompressed size across all parts (512 MiB).
	MaxTotalBytes = 512 << 20

	// MaxCompressionRatio flags suspicious entries (zip-bomb heuristic).
	MaxCompressionRatio = 100
)
