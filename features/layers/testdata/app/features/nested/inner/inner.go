// This package sits in a nested Go module, and that module is no part of the app.
// The linter must not count this import as a consumer of `entities/lone`.
package inner

import _ "example.com/app/entities/lone"
