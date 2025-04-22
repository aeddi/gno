// Package rule provides a simple framework for defining and validating rules.
//
// The Rule interface allows for the creation of custom rules that can be validated against
// a given input. The RuleSet struct manages a collection of rules and provides methods
// to add, update, remove, and validate rules.
package rule

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
)

var (
	// ErrNoInputProvided is returned when no input is provided for validation.
	ErrNoInputProvided = errors.New("no input provided")

	// ErrInvalidInput is returned when the input is not valid.
	ErrInvalidInput = errors.New("invalid input")

	// ErrAddressBlacklisted is returned by the BlacklistRule when the address is blacklisted.
	ErrAddressBlacklisted = errors.New("address is blacklisted")

	// ErrAddressNotWhitelisted is returned by the WhitelistRule when the address is not whitelisted.
	ErrAddressNotWhitelisted = errors.New("address is not whitelisted")

	// ErrCountdownExceeded is returned by the CountdownRule when the countdown is exceeded.
	ErrCountdownExceeded = errors.New("countdown exceeded")

	// ErrMockRuleInvalid is returned by the MockRule when the rule is invalid.
	ErrMockRuleInvalid = errors.New("mock rule is invalid")
)

// Rule defines the interface for a rule.
type Rule interface {
	// ID returns the rule's unique identifier.
	ID() string

	// String returns the string representation of the rule.
	String() string

	// Validate checks if the Rule is valid according to the input.
	Validate(input any) error
}

// BlacklistRule defines a rule that checks if an address is blacklisted.
type BlacklistRule struct {
	id        string
	blacklist []std.Address
}

// BlacklistRule implements the Rule interface.
var _ Rule = (*BlacklistRule)(nil)

// ID implements the Rule interface.
func (br *BlacklistRule) ID() string {
	return br.id
}

// addressesToString converts a slice of addresses to a string.
func addressesToString(addresses []std.Address) string {
	var sb strings.Builder

	for i, address := range addresses {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(address.String())
	}

	return sb.String()
}

// String implements the Rule interface.
func (br *BlacklistRule) String() string {
	return ufmt.Sprintf("Blacklist Rule: [%s]", addressesToString(br.blacklist))
}

// Validate implements the Rule interface. It takes an input of type std.Address.
func (br *BlacklistRule) Validate(input any) error {
	if input == nil {
		return ErrNoInputProvided
	}

	// Check if the input is of type std.Address.
	address, ok := input.(std.Address)
	if !ok {
		return ufmt.Errorf("%v: not an std.Address", ErrInvalidInput)
	}

	// Check if the address is in the blacklist.
	for _, blacklisted := range br.blacklist {
		if address == blacklisted {
			return ErrAddressBlacklisted
		}
	}

	return nil
}

// NewBlacklistRule creates a new BlacklistRule with the given ID and blacklist.
func NewBlacklistRule(id string, blacklist []std.Address) *BlacklistRule {
	return &BlacklistRule{
		id:        id,
		blacklist: blacklist,
	}
}

// WhitelistRule defines a rule that checks if an address is whitelisted.
type WhitelistRule struct {
	id        string
	whitelist []std.Address
}

// WhitelistRule implements the Rule interface.
var _ Rule = (*WhitelistRule)(nil)

// ID implements the Rule interface.
func (wr *WhitelistRule) ID() string {
	return wr.id
}

// String implements the Rule interface.
func (wr *WhitelistRule) String() string {
	return ufmt.Sprintf("Whitelist Rule: [%s]", addressesToString(wr.whitelist))
}

// Validate implements the Rule interface. It takes an input of type std.Address.
func (wr *WhitelistRule) Validate(input any) error {
	if input == nil {
		return ErrNoInputProvided
	}

	// Check if the input is of type std.Address.
	address, ok := input.(std.Address)
	if !ok {
		return ufmt.Errorf("%v: not an std.Address", ErrInvalidInput)
	}

	// Check if the address is in the whitelist.
	for _, whitelisted := range wr.whitelist {
		if address == whitelisted {
			return nil
		}
	}

	return ErrAddressNotWhitelisted
}

// NewWhitelistRule creates a new WhitelistRule with the given ID and whitelist.
func NewWhitelistRule(id string, whitelist []std.Address) *WhitelistRule {
	return &WhitelistRule{
		id:        id,
		whitelist: whitelist,
	}
}

// CountdownRule defines a rule that will be valid only for a certain number of times.
type CountdownRule struct {
	id        string
	countdown uint
}

// CountdownRule implements the Rule interface.
var _ Rule = (*CountdownRule)(nil)

// ID implements the Rule interface.
func (cr *CountdownRule) ID() string {
	return cr.id
}

// String implements the Rule interface.
func (cr *CountdownRule) String() string {
	return ufmt.Sprintf("CountdownRule Rule: %d validation remaining", cr.countdown)
}

// Validate implements the Rule interface. It takes no input.
func (cr *CountdownRule) Validate(_ any) error {
	if cr.countdown == 0 {
		return ErrCountdownExceeded
	}

	// Decrement the countdown.
	cr.countdown--

	return nil
}

// NewCountdownRule creates a new CountdownRule with the given ID and whitelist.
func NewCountdownRule(id string, countdown uint) *CountdownRule {
	return &CountdownRule{
		id:        id,
		countdown: countdown,
	}
}

// MockRule is a mock implementation of the Rule interface for testing purposes.
type MockRule struct {
	id    string
	valid bool
}

// MockRule implements the Rule interface.
var _ Rule = (*MockRule)(nil)

// ID implements the Rule interface.
func (mr *MockRule) ID() string {
	return mr.id
}

// String implements the Rule interface.
func (mr *MockRule) String() string {
	return ufmt.Sprintf("Mock Rule with valid state: %v", mr.valid)
}

// Validate implements the Rule interface.
func (mr *MockRule) Validate(_ any) error {
	if mr.valid {
		return nil
	}
	return ErrMockRuleInvalid
}

// NewMockRule creates a new MockRule with the given ID and valid state.
func NewMockRule(id string, valid bool) *MockRule {
	return &MockRule{
		id:    id,
		valid: valid,
	}
}
