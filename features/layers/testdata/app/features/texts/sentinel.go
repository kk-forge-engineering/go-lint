package texts

import (
	"errors"
	"fmt"
)

// The check accepts this variable: errors.New makes a sentinel, which a caller matches with
// errors.Is.
var ErrMissing = errors.New("missing")

// The check accepts fmt.Errorf as a sentinel when its format is a constant with no %w verb.
var ErrLimit = fmt.Errorf("over %d", 3)

// The check accepts this variable without the Err prefix. The check reads the type and the call.
var Missing = errors.New("missing")

// The check accepts this format: %% prints a percent sign and wraps no error.
var ErrPercent = fmt.Errorf("100%% done")

var ErrIndexed = fmt.Errorf("wrap: %[1]w", ErrMissing) // want `ErrIndexed is a variable on the surface of features/texts`

var ErrWrapped = fmt.Errorf("wrap: %w", ErrMissing) // want `ErrWrapped is a variable on the surface of features/texts`

var ErrNone error // want `ErrNone is a variable on the surface of features/texts`

var ErrBuilt = build() // want `ErrBuilt is a variable on the surface of features/texts`

var format = "dynamic"

var ErrDynamic = fmt.Errorf(format) // want `ErrDynamic is a variable on the surface of features/texts`

func build() error { return errors.New("built") }
