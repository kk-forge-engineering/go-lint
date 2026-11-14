package g

// HostUrl writes its initialism as a word, as the go-lint naming check asks.
// Staticcheck's ST1003 would refuse this name.
var HostUrl = "localhost"

// HostURL writes its initialism in capitals, as staticcheck's ST1003 asks.
// The go-lint naming check would refuse this name.
var HostURL = "localhost"

// Count implements encoding/json's Marshaler, though the package never imports encoding/json.
type Count int

// MarshalJSON keeps the name encoding/json calls.
func (c Count) MarshalJSON() ([]byte, error) { return []byte("1"), nil }
