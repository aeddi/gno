package rule

import (
	"std"
	"strconv"
	"strings"
	"testing"

	"gno.land/p/demo/urequire"
)

func TestBlacklistRule(t *testing.T) {
	var (
		id        = "testBR"
		blacklist = []std.Address{"blacklisted1", "blacklisted2", "blacklisted3"}
		rule      = NewBlacklistRule(id, blacklist)
	)

	// Test ID.
	urequire.Equal(t, rule.ID(), id)

	for _, blacklisted := range blacklist {
		// Test String.
		urequire.True(t, strings.Contains(rule.String(), string(blacklisted)))

		// Test Validate blacklisted address.
		urequire.ErrorIs(t, rule.Validate(blacklisted), ErrAddressBlacklisted)
	}

	// Test Validate not blacklisted address.
	urequire.NoError(t, rule.Validate(std.Address("not_blacklisted")))

	// Test wrong input type.
	urequire.ErrorContains(t, rule.Validate(0), ErrInvalidInput.Error())
}

func TestWhitelistRule(t *testing.T) {
	var (
		id        = "testWR"
		whitelist = []std.Address{"whitelist1", "whitelist2", "whitelist3"}
		rule      = NewWhitelistRule(id, whitelist)
	)

	// Test ID.
	urequire.Equal(t, rule.ID(), id)

	for _, whitelisted := range whitelist {
		// Test String.
		urequire.True(t, strings.Contains(rule.String(), string(whitelisted)))

		// Test Validate whitelisted address.
		urequire.NoError(t, rule.Validate(whitelisted))
	}

	// Test Validate not whitelisted address.
	urequire.ErrorIs(t, rule.Validate(std.Address("not_whitelisted")), ErrAddressNotWhitelisted)

	// Test wrong input type.
	urequire.ErrorContains(t, rule.Validate(0), ErrInvalidInput.Error())
}

func TestCountdownRule(t *testing.T) {
	var (
		id        = "testCR"
		countdown = uint(5)
		rule      = NewCountdownRule(id, countdown)
	)

	// Test ID.
	urequire.Equal(t, rule.ID(), id)

	for ; countdown > 0; countdown-- {
		// Test String.
		urequire.True(t, strings.Contains(rule.String(), strconv.Itoa(int(countdown))))

		// Test Validate countdown address.
		urequire.NoError(t, rule.Validate(nil))
	}

	// Test Validate with countdown exceeded.
	urequire.ErrorIs(t, rule.Validate(nil), ErrCountdownExceeded)
}
