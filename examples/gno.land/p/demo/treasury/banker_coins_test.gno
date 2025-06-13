package treasury

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestNewCoinsBanker(t *testing.T) {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test/test"))
	current := std.CurrentRealm().Address()

	// Test working cases.
	b := NewCoinsBanker()
	uassert.NotNil(t, b)
	uassert.Equal(t, b.Address(), string(current))

	b = NewCoinsBankerWithOwner(current)
	uassert.NotNil(t, b)
	uassert.Equal(t, b.Address(), string(current))

	// Test failing case.
	uassert.PanicsWithMessage(t, ErrNoOwnerProvided.Error(), func() {
		NewCoinsBankerWithOwner(std.Address(""))
	})
}
