package copilotext

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

const (
	UsageRoleMain     = "main"
	UsageRoleSubagent = "subagent"
	UsageRoleUnknown  = "unknown"
)

type rawUsageRecord struct {
	Model            string  `json:"model"`
	Initiator        string  `json:"initiator"`
	InteractionType  string  `json:"interactionType"`
	ParentToolCallID string  `json:"parentToolCallId"`
	APICallID        string  `json:"apiCallId"`
	InputTokens      int64   `json:"inputTokens"`
	CacheReadTokens  int64   `json:"cacheReadTokens"`
	CacheWriteTokens int64   `json:"cacheWriteTokens"`
	OutputTokens     int64   `json:"outputTokens"`
	Cost             float64 `json:"cost"`
	Duration         float64 `json:"duration"`
	Accounting       struct {
		SourceSessionID string `json:"sourceSessionId"`
		UsageID         string `json:"usageId"`
		Sequence        *int64 `json:"sequence"`
	} `json:"accounting"`
	CopilotUsage *struct {
		TotalNanoAIU *int64 `json:"totalNanoAiu"`
		TokenDetails []struct {
			TokenType  string `json:"tokenType"`
			TokenCount int64  `json:"tokenCount"`
		} `json:"tokenDetails"`
	} `json:"copilotUsage"`
}

func (r rawUsageRecord) key(ev rawEvent, sessionID string) string {
	source := r.Accounting.SourceSessionID
	if source == "" {
		source = ev.AgentID
	}
	if source == "" {
		source = sessionID
	}
	if r.Accounting.UsageID != "" {
		return "usage:" + source + ":" + r.Accounting.UsageID
	}
	if r.Accounting.Sequence != nil {
		return "sequence:" + source + ":" + strconv.FormatInt(*r.Accounting.Sequence, 10)
	}
	if r.APICallID != "" {
		return "api:" + source + ":" + r.APICallID
	}
	if ev.ID != "" {
		return "event:" + ev.ID
	}
	return ""
}

func (r rawUsageRecord) role(ev rawEvent) string {
	if ev.AgentID != "" || r.ParentToolCallID != "" || r.Initiator == "sub-agent" ||
		r.InteractionType == "conversation-subagent" {
		return UsageRoleSubagent
	}
	switch r.Initiator {
	case "", "user", "agent":
		return UsageRoleMain
	default:
		return UsageRoleUnknown
	}
}

func (r rawUsageRecord) usage(at time.Time) (SessionUsage, error) {
	if r.Model == "" {
		return SessionUsage{}, fmt.Errorf("missing model")
	}
	u := SessionUsage{At: at, Requests: 1, APIDurationMs: r.Duration}
	// Cost is a multiplier, not a per-call premium request charge.
	if r.Initiator == "" || r.Initiator == "user" {
		u.PremiumRequests = r.Cost
	}
	if r.CopilotUsage != nil && r.CopilotUsage.TotalNanoAIU != nil {
		u.AIU = float64(*r.CopilotUsage.TotalNanoAIU) / 1e9
	} else {
		u.Incomplete = true
	}
	if r.CopilotUsage != nil && len(r.CopilotUsage.TokenDetails) > 0 {
		for _, token := range r.CopilotUsage.TokenDetails {
			if token.TokenCount < 0 {
				return SessionUsage{}, fmt.Errorf("negative token count")
			}
			switch token.TokenType {
			case "input":
				u.InputTokens += token.TokenCount
			case "cache_read":
				u.CacheReadTokens += token.TokenCount
			case "cache_write":
				u.CacheWriteTokens += token.TokenCount
			case "output":
				u.OutputTokens += token.TokenCount
			default:
				u.Incomplete = true
			}
		}
	} else {
		u.InputTokens = r.InputTokens - r.CacheReadTokens - r.CacheWriteTokens
		u.CacheReadTokens = r.CacheReadTokens
		u.CacheWriteTokens = r.CacheWriteTokens
		u.OutputTokens = r.OutputTokens
	}
	if u.InputTokens < 0 || u.CacheReadTokens < 0 || u.CacheWriteTokens < 0 ||
		u.OutputTokens < 0 || u.AIU < 0 || u.PremiumRequests < 0 || u.APIDurationMs < 0 {
		return SessionUsage{}, fmt.Errorf("negative usage value")
	}
	return u, nil
}

func (u *SessionUsage) add(other SessionUsage) {
	u.Requests += other.Requests
	u.PremiumRequests += other.PremiumRequests
	u.AIU += other.AIU
	u.InputTokens += other.InputTokens
	u.CacheReadTokens += other.CacheReadTokens
	u.CacheWriteTokens += other.CacheWriteTokens
	u.OutputTokens += other.OutputTokens
	u.APIDurationMs += other.APIDurationMs
	u.Incomplete = u.Incomplete || other.Incomplete
}

type usageEntry struct {
	at       time.Time
	snapshot *rawSessionShutdown
	record   SessionUsage
	model    string
	role     string
}

// usageCollector folds request deltas and cumulative snapshots in timestamp order.
type usageCollector struct {
	session Session
	entries []usageEntry
	seen    map[string]bool
}

func (uc *usageCollector) handle(ev rawEvent) {
	switch ev.Type {
	case "session.shutdown", "session.usage_checkpoint":
		if ev.AgentID != "" {
			return
		}
		var data rawSessionShutdown
		if err := json.Unmarshal(ev.Data, &data); err != nil {
			logger.Warn("skipping unparsable usage snapshot", "dir", uc.session.Dir, "type", ev.Type, "error", err)
			return
		}
		if data.hasUsage() {
			uc.entries = append(uc.entries, usageEntry{at: ev.Timestamp, snapshot: &data})
		}
	case "session.usage_record":
		var data struct {
			Usage *rawUsageRecord `json:"usage"`
		}
		if err := json.Unmarshal(ev.Data, &data); err != nil {
			logger.Warn("skipping unparsable usage record", "dir", uc.session.Dir, "error", err)
			return
		}
		if data.Usage == nil {
			logger.Warn("skipping usage record without usage", "dir", uc.session.Dir)
			return
		}
		key := data.Usage.key(ev, uc.session.ID)
		if uc.seen[key] && key != "" {
			return
		}
		u, err := data.Usage.usage(ev.Timestamp)
		if err != nil {
			logger.Warn("skipping invalid usage record", "dir", uc.session.Dir, "error", err)
			return
		}
		if uc.seen == nil {
			uc.seen = make(map[string]bool)
		}
		if key != "" {
			uc.seen[key] = true
		}
		uc.entries = append(uc.entries, usageEntry{
			at: ev.Timestamp, record: u, model: data.Usage.Model, role: data.Usage.role(ev),
		})
	}
}

func (uc *usageCollector) result() (SessionUsage, bool) {
	return uc.resultWithin(time.Time{}, time.Time{})
}

// resultWithin folds usage entries recorded at since <= At < until (zero bounds are
// open). Cumulative snapshots cover usage before since, so the state folded before
// since is subtracted from the state folded before until.
func (uc *usageCollector) resultWithin(since, until time.Time) (SessionUsage, bool) {
	sort.SliceStable(uc.entries, func(i, j int) bool {
		return uc.entries[i].at.Before(uc.entries[j].at)
	})
	beforeUntil := len(uc.entries)
	if !until.IsZero() {
		beforeUntil = sort.Search(len(uc.entries), func(i int) bool { return !uc.entries[i].at.Before(until) })
	}
	beforeSince := 0
	if !since.IsZero() {
		beforeSince = sort.Search(beforeUntil, func(i int) bool { return !uc.entries[i].at.Before(since) })
	}
	if beforeSince >= beforeUntil {
		return SessionUsage{}, false
	}
	u, roles := foldUsageEntries(uc.entries[:beforeUntil])
	if beforeSince > 0 {
		base, baseRoles := foldUsageEntries(uc.entries[:beforeSince])
		u.sub(base)
		for model, baseModel := range base.Models {
			modelUsage := u.Models[model]
			modelUsage.sub(baseModel)
			if modelUsage.Requests == 0 && modelUsage.AIU == 0 && modelUsage.PremiumRequests == 0 &&
				modelUsage.InputTokens == 0 && modelUsage.CacheReadTokens == 0 &&
				modelUsage.CacheWriteTokens == 0 && modelUsage.OutputTokens == 0 {
				delete(u.Models, model)
				continue
			}
			u.Models[model] = modelUsage
		}
		for key, baseRole := range baseRoles {
			role := roles[key]
			role.sub(baseRole)
			if role.Requests == 0 && role.AIU == 0 && role.PremiumRequests == 0 &&
				role.InputTokens == 0 && role.CacheReadTokens == 0 &&
				role.CacheWriteTokens == 0 && role.OutputTokens == 0 {
				delete(roles, key)
				continue
			}
			roles[key] = role
		}
	}
	unknown := SessionUsage{
		Requests: u.Requests, PremiumRequests: u.PremiumRequests, AIU: u.AIU,
		InputTokens: u.InputTokens, CacheReadTokens: u.CacheReadTokens,
		CacheWriteTokens: u.CacheWriteTokens, OutputTokens: u.OutputTokens,
	}
	for key, role := range roles {
		role.APIDurationMs = 0
		u.Roles[key] = role
		unknown.Requests -= role.Requests
		unknown.PremiumRequests -= role.PremiumRequests
		unknown.AIU -= role.AIU
		unknown.InputTokens -= role.InputTokens
		unknown.CacheReadTokens -= role.CacheReadTokens
		unknown.CacheWriteTokens -= role.CacheWriteTokens
		unknown.OutputTokens -= role.OutputTokens
	}
	if math.Abs(unknown.AIU) < 1e-9 {
		unknown.AIU = 0
	}
	if math.Abs(unknown.PremiumRequests) < 1e-9 {
		unknown.PremiumRequests = 0
	}
	if unknown.Requests < 0 || unknown.AIU < 0 || unknown.PremiumRequests < 0 ||
		unknown.InputTokens < 0 || unknown.CacheReadTokens < 0 || unknown.CacheWriteTokens < 0 || unknown.OutputTokens < 0 {
		logger.Warn("usage records exceed snapshot totals; attribution is unknown", "dir", uc.session.Dir)
		u.Roles = map[string]SessionUsage{UsageRoleUnknown: {
			Requests: u.Requests, PremiumRequests: u.PremiumRequests, AIU: u.AIU,
			InputTokens: u.InputTokens, CacheReadTokens: u.CacheReadTokens,
			CacheWriteTokens: u.CacheWriteTokens, OutputTokens: u.OutputTokens,
		}}
		u.Incomplete = true
	} else if len(roles) == 0 || unknown.Requests > 0 || unknown.AIU > 0 || unknown.PremiumRequests > 0 ||
		unknown.InputTokens > 0 || unknown.CacheReadTokens > 0 || unknown.CacheWriteTokens > 0 || unknown.OutputTokens > 0 {
		role := u.Roles[UsageRoleUnknown]
		role.add(unknown)
		u.Roles[UsageRoleUnknown] = role
	}
	var modelUsage SessionUsage
	for _, model := range u.Models {
		modelUsage.add(model)
	}
	u.Incomplete = u.Incomplete || len(u.Models) == 0 || math.Abs(u.AIU-modelUsage.AIU) > 1e-9 ||
		math.Abs(u.PremiumRequests-modelUsage.PremiumRequests) > 1e-9 ||
		u.Requests != modelUsage.Requests || u.InputTokens != modelUsage.InputTokens ||
		u.CacheReadTokens != modelUsage.CacheReadTokens || u.CacheWriteTokens != modelUsage.CacheWriteTokens ||
		u.OutputTokens != modelUsage.OutputTokens
	if len(u.Models) == 0 {
		u.Models = nil
	}
	return u, true
}

func (u *SessionUsage) sub(other SessionUsage) {
	u.Requests -= other.Requests
	u.PremiumRequests -= other.PremiumRequests
	u.AIU -= other.AIU
	u.InputTokens -= other.InputTokens
	u.CacheReadTokens -= other.CacheReadTokens
	u.CacheWriteTokens -= other.CacheWriteTokens
	u.OutputTokens -= other.OutputTokens
	u.APIDurationMs -= other.APIDurationMs
}

// foldUsageEntries folds sorted request deltas and cumulative snapshots.
func foldUsageEntries(entries []usageEntry) (SessionUsage, map[string]SessionUsage) {
	u := SessionUsage{Models: make(map[string]SessionUsage), Roles: make(map[string]SessionUsage)}
	roles := make(map[string]SessionUsage)
	for _, entry := range entries {
		u.At = entry.at
		if entry.snapshot == nil {
			u.add(entry.record)
			model := u.Models[entry.model]
			model.add(entry.record)
			u.Models[entry.model] = model
			role := roles[entry.role]
			role.add(entry.record)
			roles[entry.role] = role
			continue
		}
		data := entry.snapshot
		snapshot := data.sessionUsage(entry.at)
		// Snapshots replace covered counters; subsequent records add only new usage.
		if data.TotalPremiumRequests != nil {
			u.PremiumRequests = snapshot.PremiumRequests
		}
		if data.TotalNanoAiu != nil {
			u.AIU = snapshot.AIU
		}
		if data.TotalAPIDurationMs != nil {
			u.APIDurationMs = snapshot.APIDurationMs
		}
		if data.TokenDetails != nil {
			u.InputTokens = snapshot.InputTokens
			u.CacheReadTokens = snapshot.CacheReadTokens
			u.CacheWriteTokens = snapshot.CacheWriteTokens
			u.OutputTokens = snapshot.OutputTokens
		}
		if len(data.ModelMetrics) > 0 {
			u.Models = snapshot.Models
			u.Requests = snapshot.Requests
		}
	}
	return u, roles
}
