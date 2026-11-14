package naming

// standardMethods lists, by name, the signatures of the standard library's interface methods that write an initialism
// in capitals. encoding/json calls MarshalJSON on any value that has it, so a type may implement json.Marshaler in a
// package that never imports encoding/json, and renaming the method would quietly stop json from calling it.
// TestStandardMethodsMatchTheStandardLibrary keeps the list equal to the installed Go's.
var standardMethods = map[string][]string{
	"ECDH":              {"func(*crypto/ecdh.PublicKey) ([]byte, error)"},
	"ID":                {"func() uint16"},
	"JSONOptions":       {"func(encoding/json/internal.NotForPublicUse)"},
	"KEM":               {"func() crypto/hpke.KEM"},
	"MarshalJSON":       {"func() ([]byte, error)"},
	"MarshalJSONTo":     {"func(*encoding/json/jsontext.Encoder) error"},
	"MarshalXML":        {"func(*encoding/xml.Encoder, encoding/xml.StartElement) error"},
	"MarshalXMLAttr":    {"func(encoding/xml.Name) (encoding/xml.Attr, error)"},
	"RGBA":              {"func() (uint32, uint32, uint32, uint32)"},
	"RGBA64At":          {"func(int, int) image/color.RGBA64"},
	"ServeHTTP":         {"func(net/http.ResponseWriter, *net/http.Request)"},
	"SetRGBA64":         {"func(int, int, image/color.RGBA64)"},
	"UnmarshalJSON":     {"func([]byte) error"},
	"UnmarshalJSONFrom": {"func(*encoding/json/jsontext.Decoder) error"},
	"UnmarshalXML":      {"func(*encoding/xml.Decoder, encoding/xml.StartElement) error"},
	"UnmarshalXMLAttr":  {"func(encoding/xml.Attr) error"},
	"XORKeyStream":      {"func([]byte, []byte)"},
}
