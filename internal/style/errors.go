package style

import "errors"

// ErrStylesParseFailed is returned when styles.xml cannot be parsed.
// Callers match with errors.Is; all parse errors wrap this via
// fmt.Errorf("style: decode styles.xml: %v: %w", err, ErrStylesParseFailed).
var ErrStylesParseFailed = errors.New("style: styles.xml parse failed")
