package treasury

import (
	"std"
	"strconv"

	grc20factory "gno.land/r/demo/grc20factory"
)

// GRC20Banker is a Banker that sends GRC20 tokens, which were previously
// registered in the GRC20 factory.
type GRC20Banker struct {
	ownAddress std.Address // The address of this GRC20 banker owner.
}

var _ Banker = (*GRC20Banker)(nil)

// ID implements Banker.
func (GRC20Banker) ID() string {
	return "GRC20"
}

// Send implements Banker.
func (gb *GRC20Banker) Send(p Payment) {
	// Check if payment is of type grc20Payment.
	payment, ok := p.(grc20Payment)
	if !ok {
		panic("Payment must be of type grc20Payment")
	}

	// Send the GRC20 token.
	grc20factory.TransferFrom(
		cross,
		payment.symbol,
		gb.ownAddress,
		payment.toAddress,
		payment.amount,
	)
}

// Balances implements Banker.
func (gb *GRC20Banker) Balances() []Balance {
	// Get the GRC20 tokens from the factory.
	tokens := grc20factory.ListTokens()

	// Convert GRC20 tokens to []Balance.
	var balances []Balance
	for _, token := range tokens {
		// Skip this balance if our address is a not known account for this token.
		if !token.IsKnownAccount(gb.ownAddress) {
			continue
		}
		balances = append(balances, Balance{
			Denom:  token.GetSymbol(),
			Amount: grc20factory.BalanceOf(token.GetSymbol(), gb.ownAddress),
		})
	}
	return balances
}

// Address implements Banker.
func (gb *GRC20Banker) Address() string {
	return gb.ownAddress.String()
}

// NewGRC20Banker creates a new GRC20Banker with the current Realm's address
// as the owner.
func NewGRC20Banker() *GRC20Banker {
	owner := std.CurrentRealm().Address()

	return NewGRC20BankerWithOwner(std.Address{})
}

// NewGRC20BankerWithOwner creates a new GRC20Banker with the given address.
func NewGRC20BankerWithOwner(owner std.Address) *GRC20Banker {
	if owner == "" {
		panic("owner address cannot be empty")
	}

	return &GRC20Banker{
		ownAddress: owner,
	}
}

// grc20Payment represents a payment that is issued by a GRC20Banker.
type grc20Payment struct {
	symbol    string      // The symbol of the GRC20 token.
	amount    int64       // The amount of token to send.
	toAddress std.Address // The recipient of the payment.
}

var _ Payment = (*grc20Payment)(nil)

// BankerID implements Payment.
func (grc20Payment) BankerID() string {
	return GRC20Banker{}.ID()
}

// String implements Payment.
func (gp grc20Payment) String() string {
	amount := strconv.Itoa(int(gp.amount))
	return amount + gp.symbol + " to " + gp.toAddress.String()
}

// NewGRC20Payment creates a new grc20Payment.
func NewGRC20Payment(symbol string, amount int64, toAddress std.Address) Payment {
	return grc20Payment{
		symbol:    symbol,
		amount:    amount,
		toAddress: toAddress,
	}
}
