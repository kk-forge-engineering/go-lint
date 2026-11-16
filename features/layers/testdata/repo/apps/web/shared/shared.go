package shared

import "example.com/repo/api/shared" // want `an import into another app's source is forbidden \(example\.com/repo/api/shared\)`

// Name is what a package may not import from the web app.
const Name = shared.Name
