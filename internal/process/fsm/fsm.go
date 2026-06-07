// Package fsm provides a pure finite state machine engine.
// Loads YAML process definitions, checks transitions, applies guards and effects.
package fsm

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// State is a named lifecycle state.
type State string

// Transition represents a possible state transition with guard and effect references.
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

// Machine executes transitions on a Definition.
type Machine struct {
	Def        Definition
	transCache map[string]*Transition // O(1) lookup cache
}

// New creates a Machine from a Definition.
func New(def Definition) *Machine {
	m := &Machine{Def: def, transCache: make(map[string]*Transition, len(def.Transitions))}
	for i := range def.Transitions {
		m.transCache[def.Transitions[i].Name] = &def.Transitions[i]
	}
	return m
}

// NewFromYAML parses a YAML process definition.
func NewFromYAML(data []byte) (*Machine, error) {
	var def Definition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("fsm: parse definition: %w", err)
	}
	return New(def), nil
}

// Can returns true if the transition is valid from the current state.
func (m *Machine) Can(currentState string, transitionName string) bool {
	tr := m.getTransition(transitionName)
	if tr == nil {
		return false
	}
	for _, from := range tr.From {
		if from == "*" || from == currentState {
			return true
		}
	}
	return false
}

// Next returns the target state for a transition, or an error.
func (m *Machine) Next(currentState string, transitionName string) (string, error) {
	tr := m.getTransition(transitionName)
	if tr == nil {
		return "", fmt.Errorf("fsm: transition %q not found in process %q", transitionName, m.Def.Name)
	}
	for _, from := range tr.From {
		if from == "*" || from == currentState {
			return tr.To, nil
		}
	}
	return "", fmt.Errorf("fsm: transition %q is not available from state %q", transitionName, currentState)
}

// Available returns all transitions available from the current state.
func (m *Machine) Available(currentState string) []Transition {
	var result []Transition
	for _, tr := range m.Def.Transitions {
		for _, from := range tr.From {
			if from == "*" || from == currentState {
				result = append(result, tr)
				break
			}
		}
	}
	return result
}

// GuardsFor returns the guard names for a specific transition.
func (m *Machine) GuardsFor(transitionName string) []string {
	tr := m.getTransition(transitionName)
	if tr == nil {
		return nil
	}
	return tr.Guards
}

// EffectsFor returns the effect names for a specific transition.
func (m *Machine) EffectsFor(transitionName string) []string {
	tr := m.getTransition(transitionName)
	if tr == nil {
		return nil
	}
	return tr.Effects
}

// GetTransition returns the Transition definition for a given name, or nil.
func (m *Machine) GetTransition(name string) *Transition {
	return m.transCache[name]
}

func (m *Machine) getTransition(name string) *Transition {
	return m.transCache[name]
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
