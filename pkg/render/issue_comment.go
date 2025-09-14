package render

import (
	"github.com/google/go-github/v73/github"
)

type issueCommentFieldGetter func(comment *github.IssueComment) string
type issueCommentFieldGetters struct {
	Func map[string]issueCommentFieldGetter
}

func NewIssueCommentFieldGetters() *issueCommentFieldGetters {
	return &issueCommentFieldGetters{
		Func: map[string]issueCommentFieldGetter{
			"ID": func(comment *github.IssueComment) string {
				return ToString(comment.ID)
			},
			"BODY": func(comment *github.IssueComment) string {
				return ToString(comment.Body)
			},
			"USER": func(comment *github.IssueComment) string {
				return ToString(comment.User.Login)
			},
			"CREATED_AT": func(comment *github.IssueComment) string {
				return ToString(comment.CreatedAt)
			},
			"UPDATED_AT": func(comment *github.IssueComment) string {
				return ToString(comment.UpdatedAt)
			},
		},
	}
}

func (g *issueCommentFieldGetters) GetField(comment *github.IssueComment, field string) string {
	if getter, ok := g.Func[field]; ok {
		return getter(comment)
	}
	return ""
}

func (r *Renderer) RenderIssueComments(comments []*github.IssueComment, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(comments)
		return
	}
	getter := NewIssueCommentFieldGetters()
	table := r.newTableWriter(headers)

	for _, comment := range comments {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(comment, header)
		}
		table.Append(row)
	}
	table.Render()
}

func (r *Renderer) RenderIssueCommentsDefault(comments []*github.IssueComment) {
	headers := []string{"ID", "BODY", "USER", "CREATED_AT", "UPDATED_AT"}
	r.RenderIssueComments(comments, headers)
}
