package gardrails

import "sync"

type Grardrail struct {
	readonly bool
}

var gardrail *Grardrail
var gardrailOnce sync.Once

type GradrailOption interface {
	Apply(*Grardrail)
}

type readOnlyOption struct {
	readonly bool
}

// ReadOnlyOption creates an option to configure read-only mode.
func ReadOnlyOption(readonly bool) *readOnlyOption {
	return &readOnlyOption{readonly: readonly}
}

func (o *readOnlyOption) Apply(g *Grardrail) {
	g.readonly = o.readonly
}

func NewGardrail(options ...GradrailOption) *Grardrail {
	gardrailOnce.Do(func() {
		gardrail = &Grardrail{readonly: false}
		for _, opt := range options {
			opt.Apply(gardrail)
		}
	})
	return gardrail
}

func GetGardrail() *Grardrail {
	return gardrail
}

// IsReadonly returns whether the guardrail is in read-only mode.
func (g *Grardrail) IsReadonly() bool {
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
