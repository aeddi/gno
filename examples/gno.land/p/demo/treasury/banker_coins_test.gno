package treasury

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestNewCoinsBanker(t *testing.T) {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test/test"))
	current := std.CurrentRealm().Address()

	b := NewCoinsBanker()
	uassert.NotNil(t, b)
	uassert.Equal(t, b.Address(), current)

	b = NewCoinsBankerWithOwner(current)
	uassert.NotNil(t, b)
	uassert.Equal(t, b.Address(), current)
}
