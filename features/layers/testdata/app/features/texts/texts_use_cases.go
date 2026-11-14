// Package texts is a feature whose surface gives out its use cases and its types.
package texts

// Options holds the settings and dependencies of the use cases.
type Options struct{ Retries int }

// UseCases is the feature's behaviour, one method per scenario.
type UseCases struct{ retries int }

// New builds the use cases from their options.
func New(o Options) *UseCases { return &UseCases{retries: o.Retries} }

// Save is one scenario.
func (u *UseCases) Save() int { return u.retries }
