package gardrails

import "sync"

type Gardrail struct {
	readonly bool
}

var gardrail *Gardrail
var gardrailOnce sync.Once

type GardrailOption interface {
	Apply(*Gardrail)
}

type readOnlyOption struct {
	readonly bool
}

// ReadOnlyOption creates an option to configure read-only mode.
func ReadOnlyOption(readonly bool) *readOnlyOption {
	return &readOnlyOption{readonly: readonly}
}

func (o *readOnlyOption) Apply(g *Gardrail) {
	g.readonly = o.readonly
}

func NewGardrail(options ...GardrailOption) *Gardrail {
	gardrailOnce.Do(func() {
		gardrail = &Gardrail{readonly: false}
		for _, opt := range options {
			opt.Apply(gardrail)
		}
	})
	return gardrail
}

func GetGardrail() *Gardrail {
	return gardrail
}

// IsReadonly returns whether the guardrail is in read-only mode.
func (g *Gardrail) IsReadonly() bool {
	if g == nil {
		return false
	}
	return g.readonly
}

func IsReadonly() bool {
	if gardrail == nil {
		return false
	}
	return gardrail.IsReadonly()
}
