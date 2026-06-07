// Package engine provides the validation rule registry and execution engine.
package engine

import (
	"context"

	"github.com/Homiakus/go-plm/internal/core/diagnostic"
	"github.com/Homiakus/go-plm/internal/core/object"
)

// Target represents what is being validated.
type Target struct {
	Object   object.Object
	AllObjects []object.Object
}

// Rule is a single validation check.
type Rule interface {
	ID() string
	Check(ctx context.Context, target Target) ([]diagnostic.Diagnostic, error)
}

// Registry stores all registered validation rules in registration order.
type Registry struct {
	rules []Rule
	index map[string]int
}

// NewRegistry creates an empty rule registry.
func NewRegistry() *Registry {
	return &Registry{
		rules: make([]Rule, 0),
		index: make(map[string]int),
	}
}

// Register adds a rule to the registry in order.
func (r *Registry) Register(rule Rule) {
	r.index[rule.ID()] = len(r.rules)
	r.rules = append(r.rules, rule)
}

// Get retrieves a rule by ID.
func (r *Registry) Get(id string) (Rule, bool) {
	idx, ok := r.index[id]
	if !ok {
		return nil, false
	}
	return r.rules[idx], true
}

// Engine executes validation rules.
type Engine struct {
	Registry *Registry
}

// NewEngine creates a new validation engine.
func NewEngine(reg *Registry) *Engine {
	return &Engine{Registry: reg}
}

// ValidateObject runs all registered rules in registration order.
func (e *Engine) ValidateObject(ctx context.Context, obj object.Object, all []object.Object) ([]diagnostic.Diagnostic, error) {
	target := Target{Object: obj, AllObjects: all}
	var allDiags []diagnostic.Diagnostic

	for _, rule := range e.Registry.rules {
		diags, err := rule.Check(ctx, target)
		if err != nil {
			return nil, err
		}
		allDiags = append(allDiags, diags...)
	}

	return allDiags, nil
}

// ValidateProject runs all registered rules against every object.
func (e *Engine) ValidateProject(ctx context.Context, objects []object.Object) ([]diagnostic.Diagnostic, error) {
	var allDiags []diagnostic.Diagnostic
	for _, obj := range objects {
		diags, err := e.ValidateObject(ctx, obj, objects)
		if err != nil {
			return nil, err
		}
		allDiags = append(allDiags, diags...)
	}
	return allDiags, nil
}
