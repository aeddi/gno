package treasury

import (
	"errors"
	"std"
	"strconv"

	"gno.land/p/demo/grc/grc20"
)

// GRC20Banker is a Banker that sends GRC20 tokens listed using a getter
// set during initialization.
type GRC20Banker struct {
	owner  std.Address     // The address of this GRC20 banker owner.
	lister TokenListerFunc // Allows to list tokens from methods that require it.
}

// TokenListerFunc is a function type that returns a list of GRC20 tokens.
type TokenListerFunc func() []*grc20.Token

var _ Banker = (*GRC20Banker)(nil)

// ID implements Banker.
func (GRC20Banker) ID() string {
	return "GRC20"
}

// Send implements Banker.
func (gb *GRC20Banker) Send(p Payment) error {
	if std.CurrentRealm().Address() != gb.owner {
		return ErrCurrentRealmIsNotOwner
	}

	payment, ok := p.(grc20Payment)
	if !ok {
		return errors.New("Payment must be of type grc20Payment")
	}

	// Get the GRC20 tokens using the lister.
	tokens := gb.lister()

	// Find the token corresponding to the payment symbol.
	var token *grc20.Token
	for i := range tokens {
		if tokens[i].GetSymbol() == payment.symbol {
			token = tokens[i]
			break
		}
	}
	if token == nil {
		return errors.New("GRC20 token not found with symbol: " + payment.symbol)
	}

	// Send the token.
	return token.RealmTeller().Transfer(payment.toAddress, payment.amount)
}

// Balances implements Banker.
func (gb *GRC20Banker) Balances() []Balance {
	// Get the GRC20 tokens from the lister.
	tokens := gb.lister()

	// Convert GRC20 tokens to []Balance.
	var balances []Balance
	for _, token := range tokens {
		// Skip this balance if our address is a not known account for this token.
		if !token.IsKnownAccount(gb.owner) {
			continue
		}

		balances = append(balances, Balance{
			Denom:  token.GetSymbol(),
			Amount: token.BalanceOf(gb.owner),
		})
	}
	return balances
}

// Address implements Banker.
func (gb *GRC20Banker) Address() string {
	return gb.owner.String()
}

// NewGRC20Banker creates a new GRC20Banker with the current Realm's address
// as the owner.
func NewGRC20Banker(lister TokenListerFunc) *GRC20Banker {
	owner := std.CurrentRealm().Address()

	return NewGRC20BankerWithOwner(owner, lister)
}

// NewGRC20BankerWithOwner creates a new GRC20Banker with the given address.
func NewGRC20BankerWithOwner(owner std.Address, lister TokenListerFunc) *GRC20Banker {
	if owner == "" {
		panic("owner address cannot be empty")
	}

	if lister == nil {
		panic("lister function cannot be nil")
	}

	return &GRC20Banker{
		owner:  owner,
		lister: lister,
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
