// This fixture checks that the analyzer ignores imports in `_test.go` files.
// A test file in this package imports two features, and the package still imports none.
package empty // want `orchestration/empty imports no feature: move it into the feature or entrypoint that uses it, since an orchestration module builds one action from two or more features`

import _ "example.com/app/shared"
