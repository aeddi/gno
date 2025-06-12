package treasury

import (
	"std"
)

// CoinsBanker is a Banker that sends std.Coins.
type CoinsBanker struct {
	ownAddress std.Address // The address of this coins banker owner.
}

var _ Banker = (*CoinsBanker)(nil)

// ID implements Banker.
func (CoinsBanker) ID() string {
	return "Coins"
}

// Send implements Banker.
func (cb *CoinsBanker) Send(p Payment) {
	// Check if payment is of type coinsPayment.
	payment, ok := p.(coinsPayment)
	if !ok {
		panic("Payment must be of type coinsPayment")
	}

	// Send the coins.
	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(cb.ownAddress, payment.toAddress, payment.coins)
}

// Balances implements Banker.
func (cb *CoinsBanker) Balances() []Balance {
	// Get the coins from the banker.
	banker := std.NewBanker(std.BankerTypeReadOnly)
	coins := banker.GetCoins(cb.ownAddress)

	// Convert std.Coins to []Balance.
	balances := make([]Balance, len(coins))
	for i := range coins {
		balances[i] = Balance{
			Denom:  coins[i].Denom,
			Amount: coins[i].Amount,
		}
	}

	return balances
}

// Address implements Banker.
func (cb *CoinsBanker) Address() string {
	return cb.ownAddress.String()
}

// NewCoinsBanker creates a new CoinsBanker with the current Realm's address
// as the owner.
func NewCoinsBanker() *CoinsBanker {
	owner := std.CurrentRealm().Address()

	return NewCoinsBankerWithOwner(std.Address{})
}

// NewCoinsBankerWithOwner creates a new CoinsBanker with the given address.
func NewCoinsBankerWithOwner(owner std.Address) *CoinsBanker {
	if owner == "" {
		panic("owner address cannot be empty")
	}

	return &CoinsBanker{
		ownAddress: owner,
	}
}

// coinsPayment represents a payment that is issued by a CoinsBanker.
type coinsPayment struct {
	coins     std.Coins   // The coins being sent.
	toAddress std.Address // The recipient of the payment.
}

var _ Payment = (*coinsPayment)(nil)

// BankerID implements Payment.
func (coinsPayment) BankerID() string {
	return CoinsBanker{}.ID()
}

// String implements Payment.
func (cp coinsPayment) String() string {
	return cp.coins.String() + " to " + cp.to.String()
}

// NewCoinsPayment creates a new coinsPayment.
func NewCoinsPayment(coins std.Coins, toAddress std.Address) Payment {
	return coinsPayment{
		coins:     coins,
		toAddress: toAddress,
	}
}
