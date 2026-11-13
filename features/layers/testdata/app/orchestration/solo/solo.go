// This fixture builds on one feature only. It reaches that feature through an inner package.
// The check counts that import the same as a direct import of the feature.
package solo // want `orchestration/solo imports only features/b: move it into features/b, and let the entrypoint start that feature directly, since an orchestration module builds one action from two or more features`

import _ "example.com/app/entities/solo"
