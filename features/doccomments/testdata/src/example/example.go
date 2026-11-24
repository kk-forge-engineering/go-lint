/* Package example is the fixture of the doccomments analyzer. */ // want `a doc comment is // lines, not a /\* \*/ block`
package example

/*
#include <stdlib.h>
*/
import "C"

/* Limit is the most a caller may ask for. */ // want `a doc comment is // lines, not a /\* \*/ block`
const Limit = 10

/** Account holds a balance. */ // want `a doc comment is // lines, not a /\* \*/ block`
type Account struct {
	/* Balance is in cents. */ // want `a doc comment is // lines, not a /\* \*/ block`
	Balance int
	// Owner names the holder.
	Owner string
	Note  string /* a trailing comment is no doc comment */
}

// Store keeps the accounts.
type Store interface {
	/* Find returns the account, or nil. */ // want `a doc comment is // lines, not a /\* \*/ block`
	Find(id string) *Account
}

var (
	/* Default is the account a test starts from. */ // want `a doc comment is // lines, not a /\* \*/ block`
	Default = Account{}
	// Empty has no owner.
	Empty = Account{}
)

// Deposit adds amount to the balance.
// A line doc comment passes.
func (account *Account) Deposit(amount int) {
	/* A block inside a body is no doc comment. */
	account.Balance += amount
}

// Mixed starts with lines,
/* and ends in a block. */ // want `a doc comment is // lines, not a /\* \*/ block`
func Mixed() {}

