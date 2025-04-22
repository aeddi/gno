package rule

import (
	"std"
	"testing"

	"gno.land/p/demo/urequire"
)

var rules = []Rule{
	NewMockRule("rule1", true),
	NewMockRule("rule2", true),
	NewMockRule("rule3", true),
}

func TestRuleSetBasic(t *testing.T) {
	// Create a new RuleSet with 3 valid rules.
	rs, err := NewRuleSet(rules...)
	urequire.NoError(t, err)

	// Test if the RuleSet contains the correct number of rules.
	urequire.Equal(t, rs.Size(), len(rules))

	// Test if the RuleSet contains the correct rules IDs.
	ids := rs.IDs()
	for i := range rules {
		urequire.Equal(t, ids[i], rules[i].ID())
		urequire.True(t, rs.HasRule(rules[i].ID()))
	}

	// Test if the validation is successful for all rules.
	urequire.NoError(t, rs.Validate(nil))
}

func TestRuleSetAddRule(t *testing.T) {
	// Create a new RuleSet with 3 valid rules.
	rs, err := NewRuleSet(rules...)
	urequire.NoError(t, err)

	// Test if the RuleSet contains the correct number of rules.
	urequire.Equal(t, rs.Size(), len(rules))

	// Add a new rule to the RuleSet and test if it is added correctly.
	newRule := NewMockRule("rule4", true)
	urequire.NoError(t, rs.AddRule(newRule))

	// Test if the RuleSet contains the correct number of rules.
	urequire.Equal(t, rs.Size(), len(rules)+1)

	// Test if the RuleSet contains the new rule.
	urequire.True(t, rs.HasRule(newRule.ID()))
	rule, err := rs.GetRule(newRule.ID())
	urequire.NoError(t, err)
	urequire.Equal(t, rule.String(), newRule.String())

	// Test if the validation is successful for all rules.
	urequire.NoError(t, rs.Validate(nil))

	// Test to add a rule that already exists.
	urequire.ErrorIs(t, rs.AddRule(newRule), ErrRuleAlreadyExists)
}

func TestRuleSetAddRuleSet(t *testing.T) {
	// Create a new RuleSet with 3 valid rules.
	rs, err := NewRuleSet(rules...)
	urequire.NoError(t, err)

	// Test if the RuleSet contains the correct number of rules.
	urequire.Equal(t, rs.Size(), len(rules))

	// Add a new rule to the RuleSet and test if it is added correctly.
	newRules := []Rule{
		NewMockRule("rule4", true),
		NewMockRule("rule5", true),
		NewMockRule("rule6", true),
	}
	newRS, err := NewRuleSet(newRules...)
	urequire.NoError(t, err)
	urequire.NoError(t, rs.AddRuleSet(newRS))

	// Test if the RuleSet contains the correct number of rules.
	urequire.Equal(t, rs.Size(), len(rules)+len(newRules))

	// Test if the RuleSet contains the new rules.
	for _, rule := range newRules {
		urequire.True(t, rs.HasRule(rule.ID()))
		rule, err := rs.GetRule(rule.ID())
		urequire.NoError(t, err)
		urequire.Equal(t, rule.String(), rule.String())
	}

	// Test if the validation is successful for all rules.
	urequire.NoError(t, rs.Validate(nil))

	// Test to add a rule that already exists.
	urequire.ErrorIs(t, rs.AddRule(newRules[0]), ErrRuleAlreadyExists)
}

func TestRuleSetUpdateRule(t *testing.T) {
	// Create a new RuleSet with 3 valid rules.
	rs, err := NewRuleSet(rules...)
	urequire.NoError(t, err)

	// Test to update a rule that does not exist.
	urequire.ErrorIs(t, rs.UpdateRule(NewMockRule("non_existent_rule", true)), ErrUnknownRule)

	// Update a rule in the RuleSet and test if it is updated correctly.
	updatedRule := NewMockRule("rule1", false)
	urequire.NoError(t, rs.UpdateRule(updatedRule))

	// Now validation should fail for the updated rule.
	urequire.ErrorIs(t, rs.Validate(nil), ErrMockRuleInvalid)
}

func TestRuleSetRemoveRule(t *testing.T) {
	// Create a new RuleSet with 3 valid rules.
	rs, err := NewRuleSet(rules...)
	urequire.NoError(t, err)

	// Test to remove a rule that does not exist.
	urequire.ErrorIs(t, rs.RemoveRule("non_existent_rule"), ErrUnknownRule)

	// Test if the RuleSet contains the correct number of rules.
	urequire.Equal(t, rs.Size(), len(rules))

	// Add an invalid rule to the RuleSet, the validation should then fail.
	invalidRule := NewMockRule("invalid_rule", false)
	urequire.NoError(t, rs.AddRule(invalidRule))
	urequire.ErrorIs(t, rs.Validate(nil), ErrMockRuleInvalid)

	// Test if the RuleSet contains the correct number of rules.
	urequire.Equal(t, rs.Size(), len(rules)+1)

	// Remove the invalid rule and test if the validation is successful.
	urequire.NoError(t, rs.RemoveRule(invalidRule.ID()))
	urequire.NoError(t, rs.Validate(nil))

	// Test if the RuleSet contains the correct number of rules.
	urequire.Equal(t, rs.Size(), len(rules))
}

func TestRuleValidateEach(t *testing.T) {
	// Create a new RuleSet with 3 valid rules.
	rs, err := NewRuleSet(rules...)
	urequire.NoError(t, err)

	// The validation should fail if no input is provided.
	urequire.ErrorContains(t, rs.ValidateEach(nil), ErrNoInputProvided.Error())

	// Create a map of inputs for each rule.
	inputs := map[string]any{
		"rule1": nil,
		"rule2": nil,
		"rule3": nil,
	}

	// The validation should now succeed.
	urequire.NoError(t, rs.ValidateEach(inputs))

	// Add a whitelist rule to the RuleSet.
	var (
		whitelisted    = std.Address("whitelisted_address")
		notWhitelisted = std.Address("not_whitelisted_address")
	)
	whitelistRule := NewWhitelistRule("rule4", []std.Address{whitelisted})
	urequire.NoError(t, rs.AddRule(whitelistRule))

	// Add a non-whitelisted address to the inputs.
	inputs[whitelistRule.ID()] = notWhitelisted

	// The validation should now fail.
	err = rs.ValidateEach(inputs)
	urequire.ErrorContains(t, err, ErrAddressNotWhitelisted.Error())

	// Check if the error is identified correctly.
	idErr := err.(*IdentifiedError)
	urequire.ErrorIs(t, idErr.Get(whitelistRule.ID()), ErrAddressNotWhitelisted)

	// Set the input to a whitelisted address.
	inputs[whitelistRule.ID()] = whitelisted

	// The validation should now succeed.
	urequire.NoError(t, rs.ValidateEach(inputs))
}
