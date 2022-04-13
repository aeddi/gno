package tests

import (
	"strconv"
)

type Stringer interface {
	String() string
}

var stringers []Stringer

func AddStringer(str Stringer) {
	// NOTE: this is ridiculous, a slice that will become too long
	// eventually.  Don't do this in production programs; use
	// gno.land/p/avl or similar structures.
	stringers = append(stringers, str)
}

func Render(path string) string {
	res := ""
	// NOTE: like the function above, this function too will eventually
	// become too expensive to call.
	for i, stringer := range stringers {
		res += strconv.Itoa(i) + ": " + stringer.String() + "\n"
	}
	return res
}
