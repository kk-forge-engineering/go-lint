package g

// HostUrl writes its initialism as a word, as naming asks, and as staticcheck's ST1003 would refuse.
var HostUrl = "localhost"

// HostURL writes its initialism in capitals.
var HostURL = "localhost"

// Count implements encoding/json's Marshaler, though the package never imports encoding/json.
type Count int

// MarshalJSON keeps the name encoding/json calls.
func (c Count) MarshalJSON() ([]byte, error) { return []byte("1"), nil }
