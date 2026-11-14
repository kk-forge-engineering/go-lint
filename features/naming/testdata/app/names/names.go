package names

import (
	"encoding/json"
	serverHTTP "net/http" // want `serverHTTP: write each initialism as a word, as in serverHttp`
	server_url "net/url" // want `server_url: write an import alias in camelCase, as in serverUrl`
	. "unicode/utf8"
	Strconv "strconv" // want `Strconv: write an import alias in camelCase, as in strconv`
	"strings"
)

type HostURL struct { // want `HostURL: write each initialism as a word, as in HostUrl`
	userID int // want `userID: write each initialism as a word, as in userId`
	Name   string `json:"userID"`
	*server_url.URL
}

const MaxID = 1 // want `MaxID: write each initialism as a word, as in MaxId`

var defaultTTL = 1 // want `defaultTTL: write each initialism as a word, as in defaultTtl`

func ParseURL(rawURL string) (*server_url.URL, error) { // want `ParseURL: write each initialism as a word, as in ParseUrl` `rawURL: write each initialism as a word, as in rawUrl`
	return server_url.Parse(strings.TrimSpace(rawURL))
}

func HTTPServerName() string { return "" } // want `HTTPServerName: write each initialism as a word, as in HttpServerName`

// ServeHTTP implements net/http's Handler, whose method name go-lint cannot change.
func (h HostURL) ServeHTTP(serverHTTP.ResponseWriter, *serverHTTP.Request) {}

// MarshalJSON implements encoding/json's Marshaler, so the check keeps the name encoding/json
// calls.
func (h HostURL) MarshalJSON() ([]byte, error) { return json.Marshal(h.Name) }

// FetchHTTP implements no interface's method, so the check reports its name.
func (h HostURL) FetchHTTP() {} // want `FetchHTTP: write each initialism as a word, as in FetchHttp`

type other struct{}

// ServeHTTP has the name of net/http's Handler method but not its signature, so the check reports
// it.
func (other) ServeHTTP() {} // want `ServeHTTP: write each initialism as a word, as in ServeHttp`

// Doer is this package's own interface, so its method names are this package's to choose.
type Doer interface {
	DoHTTP() // want `DoHTTP: write each initialism as a word, as in DoHttp`
}

func (other) DoHTTP() {} // want `DoHTTP: write each initialism as a word, as in DoHttp`

func Map[KV any](values []KV) []KV { return values } // want `KV: write each initialism as a word, as in Kv`

func fine() (hostUrl string, userId int, t int) {
	_ = HostURL{}
	return "", 0, 0
}

// ReadsAValue holds the one-letter word A before a word, not an initialism.
func ReadsAValue() {}

func retry(attempts []int) int {
retryHTTP: // want `retryHTTP: write each initialism as a word, as in retryHttp`
	for range attempts {
		continue retryHTTP
	}
	return Strconv.IntSize + RuneLen(0)
}

var userIDs []int // want `userIDs: write each initialism as a word, as in userIds`

var APIs []string // want `APIs: write each initialism as a word, as in Apis`

func parseURLs(rawURLsToRead []string) {} // want `parseURLs: write each initialism as a word, as in parseUrls` `rawURLsToRead: write each initialism as a word, as in rawUrlsToRead`
