package use

import "example.com/app/names"

type Proxy struct{ next names.Doer }

// DoHTTP implements `names.Doer`, so the package of that interface chose the name.
// The check reports the name there, not on this method.
func (p Proxy) DoHTTP() { p.next.DoHTTP() }
