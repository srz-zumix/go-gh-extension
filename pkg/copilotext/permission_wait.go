package copilotext

import (
	"time"

	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// PermissionDecisionWait measures recorded request-to-decision intervals for one source.
type PermissionDecisionWait struct {
	Completed   int
	Timed       int
	Untimed     int
	TotalWaitMs float64
	AvgWaitMs   *float64
	MaxWaitMs   *float64
}

// PermissionWaitStats keeps confirmed human responses separate from other decisions.
type PermissionWaitStats struct {
	HumanResponses         PermissionDecisionWait
	NonHumanDecisions      PermissionDecisionWait
	UnknownSourceDecisions PermissionDecisionWait
	Unresolved             int
}

func (stats *PermissionWaitStats) add(rec PermissionRecord) {
	if !rec.Completed {
		stats.Unresolved++
		return
	}
	var timing *PermissionDecisionWait
	switch rec.DecisionSource {
	case DecisionSourceHumanResponse:
		timing = &stats.HumanResponses
	case "", DecisionSourceUnknown:
		timing = &stats.UnknownSourceDecisions
	default:
		timing = &stats.NonHumanDecisions
	}
	timing.Completed++
	if rec.Request.RequestedAt.IsZero() || rec.DecidedAt.IsZero() || rec.DecidedAt.Before(rec.Request.RequestedAt) {
		timing.Untimed++
		logger.Warn("permission decision has missing or invalid timestamps",
			"session", rec.Session.ID, "requestId", rec.Request.RequestID)
		return
	}
	wait := float64(rec.DecidedAt.Sub(rec.Request.RequestedAt)) / float64(time.Millisecond)
	timing.Timed++
	timing.TotalWaitMs += wait
	if timing.MaxWaitMs == nil || wait > *timing.MaxWaitMs {
		timing.MaxWaitMs = &wait
	}
	avg := timing.TotalWaitMs / float64(timing.Timed)
	timing.AvgWaitMs = &avg
}
