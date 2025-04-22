package rule

import (
	"errors"

	"gno.land/p/demo/avl"
)

var (
	// ErrUnknownRule is returned when a rule is not found in the RuleSet.
	ErrUnknownRule = errors.New("unknown rule")

	// ErrRuleAlreadyExists is returned when a rule already exists in the RuleSet.
	ErrRuleAlreadyExists = errors.New("rule already exists")
)

// RuleSet defines a set of rules.
type RuleSet struct {
	// Rules contains the rules of the set.
	Rules *avl.Tree
}

// AddRule adds a rule to the RuleSet if it does not already exist.
func (r *RuleSet) AddRule(rule Rule) error {
	if r.Rules.Has(rule.ID()) {
		return ErrRuleAlreadyExists
	}

	r.Rules.Set(rule.ID(), rule)
	return nil
}

// AddRuleSet adds another RuleSet to the current RuleSet.
func (r *RuleSet) AddRuleSet(rs *RuleSet) error {
	if rs == nil {
		return nil
	}

	for _, id := range rs.IDs() {
		rule, _ := rs.GetRule(id)

		if err := r.AddRule(rule); err != nil {
			return err
		}
	}

	return nil
}

// UpdateRule updates a rule in the RuleSet.
func (r *RuleSet) UpdateRule(rule Rule) error {
	if !r.Rules.Has(rule.ID()) {
		return ErrUnknownRule
	}

	r.Rules.Set(rule.ID(), rule)
	return nil
}

// RemoveRule removes a rule from the RuleSet.
func (r *RuleSet) RemoveRule(id string) error {
	if !r.Rules.Has(id) {
		return ErrUnknownRule
	}

	r.Rules.Remove(id)
	return nil
}

// HasRule checks if a rule exists in the RuleSet.
func (r *RuleSet) HasRule(id string) bool {
	return r.Rules.Has(id)
}

// GetRule retrieves a rule from the RuleSet by its ID.
func (r *RuleSet) GetRule(id string) (Rule, error) {
	value, exists := r.Rules.Get(id)
	if !exists {
		return nil, ErrUnknownRule
	}

	return value.(Rule), nil
}

// Size returns the number of rules in the RuleSet.
func (r *RuleSet) Size() int {
	return r.Rules.Size()
}

// IDs returns the IDs of all rules in the RuleSet.
func (r *RuleSet) IDs() []string {
	ids := make([]string, 0, r.Rules.Size())

	r.Rules.IterateByOffset(0, r.Rules.Size(), func(id string, _ any) bool {
		ids = append(ids, id)
		return false
	})

	return ids
}

// Validate checks if all rules in the RuleSet are valid according to the same input.
// It returns the first error encountered.
func (r *RuleSet) Validate(input any) error {
	var err error

	r.Rules.IterateByOffset(0, r.Rules.Size(), func(id string, rule any) bool {
		if err = rule.(Rule).Validate(input); err != nil {
			return true
		}
		return false
	})

	return err
}

// ValidateEach checks if each rule in the RuleSet is valid according to a different input.
// It returns an IdentifiedError containing all errors encountered associated with their rule IDs.
func (r *RuleSet) ValidateEach(inputs map[string]any) error {
	var (
		ids    = r.IDs()
		errors = new(IdentifiedError)
	)

	for _, id := range ids {
		// Check if an input is provided for this rule ID.
		input, exists := inputs[id]
		if !exists {
			errors.Add(id, ErrNoInputProvided)
			continue // Skip to the next rule if no input is provided.
		}

		// Validate the rule with the provided input.
		rule, _ := r.GetRule(id)
		if err := rule.Validate(input); err != nil {
			errors.Add(id, err)
		}
	}

	if errors.Size() == 0 {
		return nil
	}

	return errors
}

// NewRuleSet creates a new RuleSet with the provided rules.
// It returns an error if any of the rules already exist in the set.
func NewRuleSet(rules ...Rule) (*RuleSet, error) {
	rs := &RuleSet{
		Rules: avl.NewTree(),
	}

	for _, rule := range rules {
		if err := rs.AddRule(rule); err != nil {
			return nil, err
		}
	}

	return rs, nil
}
