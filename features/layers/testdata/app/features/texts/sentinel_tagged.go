//go:build integration

package texts

import "errors"

// The check skips `ErrTagged`. The `integration` build tag keeps this file
// out of the host build, so the loader gives its variables no type information.
var ErrTagged = errors.New("tagged")
