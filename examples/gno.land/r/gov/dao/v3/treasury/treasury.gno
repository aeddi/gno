package treasury

import (
	t "gno.land/p/demo/treasury"

	"gno.land/r/gov/dao"
)

// Init a new Treasury instance with all available bankers.
var treasury = t.New(t.AllBankers)

// Send sends a payment using the treasury instance.
func Send(cur realm, payment t.Payment) {
	// Check if the current realm is allowed to send payments.
	if !dao.InAllowedDAOs(cur.PkgPath()) {
		panic("this Realm is not allowed to send payment: " + cur.PkgPath())
	}

	// Send the payment using the treasury instance.
	treasury.Send(payment)
}

// History returns the n last payments sent by the banker with the given ID.
// Returns the full history if n is 0 or greater than the number of payments.
func History(bankerID string, n uint) []t.Payment {
	return treasury.History(bankerID, n)
}

// HasBanker checks if a banker with the given ID is registered.
func HasBanker(bankerID string) bool {
	return treasury.HasBanker(bankerID)
}

// ListBankerIDs returns a list of all registered banker IDs.
func ListBankerIDs() []string {
	return treasury.ListBankerIDs()
}

func Render(path string) string {
	return treasury.Render(path)
}
