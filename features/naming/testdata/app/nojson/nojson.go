package nojson

import "strconv"

// Count implements encoding/json's Marshaler without the package importing encoding/json.
type Count int

// MarshalJSON keeps the name encoding/json calls.
func (c Count) MarshalJSON() ([]byte, error) { return []byte(strconv.Itoa(int(c))), nil }

// UnmarshalJSON keeps its name, though this signature is not json.Unmarshaler's.
func (c *Count) UnmarshalJSON(string) error { return nil } // want `UnmarshalJSON: write each initialism as a word, as in UnmarshalJson`
