package render

import (
	"fmt"

	"github.com/google/go-github/v79/github"
)

// RenderBranchProtections renders a list of protected branches in a table format.
func (r *Renderer) RenderBranchProtections(branches []*github.Branch) error {
	if r.exporter != nil {
		return r.RenderExportedData(branches)
	}

	if len(branches) == 0 {
		r.writeLine("No protected branches.")
		return nil
	}

	table := newStickyTable(r.newTableWriter([]string{"BRANCH"}))
	for _, b := range branches {
		table.Append([]string{ToString(b.Name)})
	}
	return table.Render()
}

// RenderBranchProtection renders detailed branch protection settings as a key-value table.
func (r *Renderer) RenderBranchProtection(branch string, protection *github.Protection) error {
	if r.exporter != nil {
		return r.RenderExportedData(protection)
	}

	r.writeLine(fmt.Sprintf("Branch: %s", branch))
	r.writeLine("")

	table := newStickyTable(r.newTableWriter([]string{"SETTING", "VALUE"}))

	// EnforceAdmins
	if protection.EnforceAdmins != nil {
		table.Append([]string{"Enforce Admins", ToString(protection.EnforceAdmins.Enabled)})
	}

	// RequireLinearHistory
	if protection.RequireLinearHistory != nil {
		table.Append([]string{"Require Linear History", ToString(protection.RequireLinearHistory.Enabled)})
	}

	// AllowForcePushes
	if protection.AllowForcePushes != nil {
		table.Append([]string{"Allow Force Pushes", ToString(protection.AllowForcePushes.Enabled)})
	}

	// AllowDeletions
	if protection.AllowDeletions != nil {
		table.Append([]string{"Allow Deletions", ToString(protection.AllowDeletions.Enabled)})
	}

	// BlockCreations
	if protection.BlockCreations != nil {
		table.Append([]string{"Block Creations", ToString(protection.BlockCreations.Enabled)})
	}

	// LockBranch
	if protection.LockBranch != nil {
		table.Append([]string{"Lock Branch", ToString(protection.LockBranch.Enabled)})
	}

	// AllowForkSyncing
	if protection.AllowForkSyncing != nil {
		table.Append([]string{"Allow Fork Syncing", ToString(protection.AllowForkSyncing.Enabled)})
	}

	// RequiredConversationResolution
	if protection.RequiredConversationResolution != nil {
		table.Append([]string{"Required Conversation Resolution", ToString(protection.RequiredConversationResolution.Enabled)})
	}

	// RequiredSignatures
	if protection.RequiredSignatures != nil {
		table.Append([]string{"Required Signatures", ToString(protection.RequiredSignatures.Enabled)})
	}

	if err := table.Render(); err != nil {
		return err
	}

	// RequiredStatusChecks
	if protection.RequiredStatusChecks != nil {
		r.writeLine("")
		r.writeLine("Required Status Checks:")
		scTable := newStickyTable(r.newTableWriter([]string{"CONTEXT", "APP_ID", "STRICT"}))
		if protection.RequiredStatusChecks.Checks != nil {
			for _, check := range *protection.RequiredStatusChecks.Checks {
				scTable.Append([]string{
					check.Context,
					ToString(check.AppID),
					ToString(protection.RequiredStatusChecks.Strict),
				})
			}
		} else if protection.RequiredStatusChecks.Contexts != nil {
			for _, ctx := range *protection.RequiredStatusChecks.Contexts {
				scTable.Append([]string{ctx, "", ToString(protection.RequiredStatusChecks.Strict)})
			}
		}
		if err := scTable.Render(); err != nil {
			return err
		}
	}

	// RequiredPullRequestReviews
	if protection.RequiredPullRequestReviews != nil {
		pr := protection.RequiredPullRequestReviews
		r.writeLine("")
		r.writeLine("Required Pull Request Reviews:")
		prTable := newStickyTable(r.newTableWriter([]string{"SETTING", "VALUE"}))
		prTable.Append([]string{"Required Approving Review Count", fmt.Sprintf("%d", pr.RequiredApprovingReviewCount)})
		prTable.Append([]string{"Dismiss Stale Reviews", ToString(pr.DismissStaleReviews)})
		prTable.Append([]string{"Require Code Owner Reviews", ToString(pr.RequireCodeOwnerReviews)})
		prTable.Append([]string{"Require Last Push Approval", ToString(pr.RequireLastPushApproval)})
		if err := prTable.Render(); err != nil {
			return err
		}
	}

	// Restrictions
	if protection.Restrictions != nil {
		r.writeLine("")
		r.writeLine("Push Restrictions:")
		restrTable := newStickyTable(r.newTableWriter([]string{"TYPE", "NAME"}))
		for _, u := range protection.Restrictions.Users {
			restrTable.Append([]string{"User", ToString(u.Login)})
		}
		for _, t := range protection.Restrictions.Teams {
			restrTable.Append([]string{"Team", ToString(t.Slug)})
		}
		for _, a := range protection.Restrictions.Apps {
			restrTable.Append([]string{"App", ToString(a.Slug)})
		}
		if len(protection.Restrictions.Users) == 0 && len(protection.Restrictions.Teams) == 0 && len(protection.Restrictions.Apps) == 0 {
			restrTable.Append([]string{"(none)", ""})
		}
		if err := restrTable.Render(); err != nil {
			return err
		}
	}
	return nil
}
