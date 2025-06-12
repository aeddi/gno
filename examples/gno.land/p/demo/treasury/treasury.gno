package treasury

import (
	"gno.land/p/demo/avl"
)

// List of all available bankers.
var AllBankers = []Banker{
	CoinsBanker{},
	GRC20Banker{},
}

// New creates a new Treasury instance with the given bankers.
func New(bankers []Banker) *Treasury {
	if len(bankers) == 0 {
		panic("at least one banker must be registered")
	}

	// Create a new Treasury instance.
	treasury := &Treasury{
		bankers: avl.NewTree(),
	}

	// Register the bankers.
	for _, banker := range bankers {
		treasury.bankers.Set(banker.ID(), &bankerRecord{
			banker:  banker,
			history: []Payment{},
		})
	}

	// Register the Render routes.
	treasury.initRenderRouter()

	return treasury
}

// Send sends a payment using the corresponding banker.
func (t *Treasury) Send(p Payment) {
	// Get the banker record corresponding to this Payment.
	br := t.mustGetBankerRecord(p.BankerID())

	// Send the payment using the corresponding banker.
	br.banker.Send(p)

	// Add the payment to the history of the banker.
	br.history = append(br.history, p)
}

// History returns the n last payments sent by the banker with the given ID.
// Returns the full history if n is 0 or greater than the number of payments.
func (t *Treasury) History(bankerID string, n uint) []Payment {
	// Get the history of payments.
	history := t.mustGetBankerRecord(bankerID).history

	// If n is 0, return the full history.
	if n == 0 || uint(len(history)) <= n {
		return history
	}

	// Return the last n payments.
	return history[len(history)-int(n):]
}

// HasBanker checks if a banker with the given ID is registered.
func (t *Treasury) HasBanker(bankerID string) bool {
	return t.bankers.Has(bankerID)
}

// ListBankerIDs returns a list of all registered banker IDs.
func (t *Treasury) ListBankerIDs() []string {
	var bankerIDs []string

	t.bankers.Iterate("", "", func(bankerID string, _ any) bool {
		bankerIDs = append(bankerIDs, bankerID)
		return false
	})

	return bankerIDs
}

func (t *Treasury) mustGetBankerRecord(bankerID string) *bankerRecord {
	// Get the banker record associated to this ID.
	br, ok := bankers.Get(bankerID)
	if !ok {
		panic("banker not found with ID: " + bankerID)
	}

	return br.(*bankerRecord)
}
