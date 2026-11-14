package use

import "example.com/app/names"

// Proxy forwards to a names.Doer.
type Proxy struct{ next names.Doer }

// DoHTTP implements names.Doer, an interface of another package of the module, so that package chose the name.
func (p Proxy) DoHTTP() { p.next.DoHTTP() }
