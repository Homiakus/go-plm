// Package lifecycle defines pure lifecycle/FSM types for process definitions.
// No filesystem, history writing, or Git dependencies — just the state machine model.
package lifecycle

import "fmt"

// State is a named lifecycle state.
type State string

// Transition represents a possible state transition with guards and effects.
type Transition struct {
	Name    string   `yaml:"name" json:"name"`
	From    []string `yaml:"from" json:"from"`
	To      string   `yaml:"to" json:"to"`
	Guards  []string `yaml:"guards,omitempty" json:"guards,omitempty"`
	Effects []string `yaml:"effects,omitempty" json:"effects,omitempty"`
}

// Definition describes a complete lifecycle process.
type Definition struct {
	Name        string       `yaml:"name" json:"name"`
	Initial     string       `yaml:"initial" json:"initial"`
	States      []string     `yaml:"states" json:"states"`
	Transitions []Transition `yaml:"transitions" json:"transitions"`
}

// Machine is a state machine instance operating on a Definition.
type Machine struct {
	Definition Definition
}

// NewMachine creates a new FSM from a Definition.
func NewMachine(def Definition) *Machine {
	return &Machine{Definition: def}
}

// Can returns true if the given transition is valid from the current state.
func (m *Machine) Can(currentState string, transitionName string) bool {
	for _, t := range m.Definition.Transitions {
		if t.Name != transitionName {
			continue
		}
		for _, from := range t.From {
			if from == "*" || from == currentState {
				return true
			}
		}
	}
	return false
}

// Next returns the target state for a transition, or an error if invalid.
func (m *Machine) Next(currentState string, transitionName string) (string, error) {
	for _, t := range m.Definition.Transitions {
		if t.Name != transitionName {
			continue
		}
		for _, from := range t.From {
			if from == "*" || from == currentState {
				return t.To, nil
			}
		}
	}
	return "", fmt.Errorf("lifecycle: transition %q is not available from state %q in process %q",
		transitionName, currentState, m.Definition.Name)
}

// Available returns all transitions available from the current state.
func (m *Machine) Available(currentState string) []Transition {
	var result []Transition
	for _, t := range m.Definition.Transitions {
		for _, from := range t.From {
			if from == "*" || from == currentState {
				result = append(result, t)
				break
			}
		}
	}
	return result
}

// GetTransition returns the Transition definition for a given name, or nil.
func (m *Machine) GetTransition(name string) *Transition {
	for i := range m.Definition.Transitions {
		if m.Definition.Transitions[i].Name == name {
			return &m.Definition.Transitions[i]
		}
	}
	return nil
}

// StandardObjectLifecycle returns the default object lifecycle definition.
func StandardObjectLifecycle() Definition {
	return Definition{
		Name:    "object_lifecycle",
		Initial: "draft",
		States:  []string{"draft", "in_review", "approved", "released", "blocked", "obsolete", "archived"},
		Transitions: []Transition{
			{Name: "submit_review", From: []string{"draft"}, To: "in_review", Guards: []string{"valid_name", "required_metadata", "no_broken_relations"}},
			{Name: "reject", From: []string{"in_review"}, To: "draft"},
			{Name: "approve", From: []string{"in_review"}, To: "approved", Guards: []string{"no_blocking_issues", "checksums_actual"}},
			{Name: "release", From: []string{"approved"}, To: "released", Guards: []string{"no_release_blockers", "children_released", "bom_valid"}, Effects: []string{"create_git_tag"}},
			{Name: "revise", From: []string{"released"}, To: "draft", Effects: []string{"create_new_revision", "create_change_event"}},
			{Name: "block", From: []string{"*"}, To: "blocked"},
			{Name: "unblock", From: []string{"blocked"}, To: "draft"},
			{Name: "obsolete", From: []string{"released"}, To: "obsolete"},
			{Name: "archive", From: []string{"obsolete"}, To: "archived"},
		},
	}
}
