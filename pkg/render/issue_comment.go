package render

import (
	"strings"

	"github.com/google/go-github/v84/github"
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
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(comment)
	}
	return ""
}

func (r *Renderer) RenderIssueComments(comments []*github.IssueComment, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(comments)
	}

	if len(comments) == 0 {
		r.WriteLine("no issue comments found")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"ID", "BODY", "USER", "CREATED_AT", "UPDATED_AT"}
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
	return table.Render()
}
