package start

import "time"

// An entrypoint reads the clock through infra too.
func now() time.Time { return time.Now() } // want `time\.Now reads the clock`
