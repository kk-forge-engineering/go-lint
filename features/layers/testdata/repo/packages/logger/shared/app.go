package shared

import "example.com/repo/web/shared" // want `a package imports no app \(example\.com/repo/web/shared\)`

// App is what a package may not take from an app.
const App = shared.Name
