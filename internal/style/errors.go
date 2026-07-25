package style

import "errors"

// ErrStylesParseFailed is returned when styles.xml cannot be parsed.
// Callers match with errors.Is; all parse errors wrap this via
// fmt.Errorf("style: decode styles.xml: %v: %w", err, ErrStylesParseFailed).
var ErrStylesParseFailed = errors.New("style: styles.xml parse failed")

// ErrCloneTargetNotEmpty is returned when CloneStyles is called on a
// target package that already has any of the 5 style parts (D-08).
// Clone is a fresh-empty-target operation only; merge-by-styleId is
// deferred to a future phase.  Callers match with errors.Is; all
// CloneStyles errors wrap this via
// fmt.Errorf("style: clone %s: %w", partName, ErrCloneTargetNotEmpty).
var ErrCloneTargetNotEmpty = errors.New("style: clone target not empty")
