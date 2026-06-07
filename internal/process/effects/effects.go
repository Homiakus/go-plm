// Package effects provides FSM effect implementations.
// Effects are side-effects executed after a successful transition.
package effects

import (
	"context"
	"fmt"
)

// EffectInput carries the data needed to apply an effect.
type EffectInput struct {
	ObjectID       string
	TransitionName string
	FromState      string
	ToState        string
}

// Effect is a side-effect executed after a lifecycle transition.
type Effect interface {
	ID() string
	Apply(ctx context.Context, input EffectInput) error
}

// CreateGitTag signals that a git tag should be created (actual tagging done by app layer).
type CreateGitTag struct{}

func (e *CreateGitTag) ID() string { return "create_git_tag" }
func (e *CreateGitTag) Apply(ctx context.Context, input EffectInput) error {
	return nil // actual tagging is done by app layer
}

// CreateNewRevision signals that a new revision should be created.
type CreateNewRevision struct{}

func (e *CreateNewRevision) ID() string { return "create_new_revision" }
func (e *CreateNewRevision) Apply(ctx context.Context, input EffectInput) error {
	return nil
}

// CreateChangeEvent signals that a change event should be recorded.
type CreateChangeEvent struct{}

func (e *CreateChangeEvent) ID() string { return "create_change_event" }
func (e *CreateChangeEvent) Apply(ctx context.Context, input EffectInput) error {
	return nil
}

// LogTransition logs the transition for observability.
type LogTransition struct{}

func (e *LogTransition) ID() string { return "log_transition" }
func (e *LogTransition) Apply(ctx context.Context, input EffectInput) error {
	fmt.Printf("[fsm] %s: %s -> %s via %s\n", input.ObjectID, input.FromState, input.ToState, input.TransitionName)
	return nil
}

// DefaultEffects returns a map of effect name -> implementation for MVP.
func DefaultEffects() map[string]Effect {
	return map[string]Effect{
		"create_git_tag":       &CreateGitTag{},
		"create_new_revision":  &CreateNewRevision{},
		"create_change_event":  &CreateChangeEvent{},
		"log_transition":       &LogTransition{},
	}
}
