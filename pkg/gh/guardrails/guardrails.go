package guardrails

import "sync"

// Guardrail represents a guardrail configuration.
type Guardrail struct {
	readonly bool
}

var guardrail *Guardrail
var guardrailOnce sync.Once

// GuardrailOption defines an option for configuring the Guardrail.
type GuardrailOption interface {
	Apply(*Guardrail)
}

type readOnlyOption struct {
	readonly bool
}

// ReadOnlyOption creates an option to configure read-only mode.
func ReadOnlyOption(readonly bool) *readOnlyOption {
	return &readOnlyOption{readonly: readonly}
}

func (o *readOnlyOption) Apply(g *Guardrail) {
	g.readonly = o.readonly
}

// NewGuardrail creates a new Guardrail instance with the provided options.
func NewGuardrail(options ...GuardrailOption) *Guardrail {
	guardrailOnce.Do(func() {
		guardrail = &Guardrail{readonly: false}
		for _, opt := range options {
			opt.Apply(guardrail)
		}
	})
	return guardrail
}

// GetGuardrail returns the singleton Guardrail instance.
func GetGuardrail() *Guardrail {
	return guardrail
}

// IsReadonly returns whether the guardrail is in read-only mode.
func (g *Guardrail) IsReadonly() bool {
	if g == nil {
		return false
	}
	return g.readonly
}

// IsReadonly returns whether the guardrail is in read-only mode.
func IsReadonly() bool {
	if guardrail == nil {
		return false
	}
	return guardrail.IsReadonly()
}
