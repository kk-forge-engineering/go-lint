package lang

import "net/url"

// ParseURL parses a URL.
func ParseURL(raw string) (*url.URL, error) { return url.Parse(raw) }
