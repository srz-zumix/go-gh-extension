package gardrails

import "sync"

type Grardrail struct {
	readonly bool
}

var gardrail *Grardrail
var gardrailOnce sync.Once

func NewGardrail(readonly bool) *Grardrail {
	gardrailOnce.Do(func() {
		gardrail = &Grardrail{readonly: readonly}
	})
	return gardrail
}

func GetGardrail() *Grardrail {
	return gardrail
}

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
