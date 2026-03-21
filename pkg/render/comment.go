package render

import (
	"strings"

	"github.com/google/go-github/v79/github"
)

type commentFieldGetters struct {
	issueCommentGetter       *issueCommentFieldGetters
	pullRequestCommentGetter *prCommentFieldGetters
}

func NewCommentFieldGetters() *commentFieldGetters {
	return &commentFieldGetters{
		issueCommentGetter:       NewIssueCommentFieldGetters(),
		pullRequestCommentGetter: NewPullRequestCommentFieldGetters(),
	}
}

func (g *commentFieldGetters) GetField(comment any, field string) string {
	if v, ok := comment.(*github.IssueComment); ok {
		return g.issueCommentGetter.GetField(v, field)
	}
	if v, ok := comment.(*github.PullRequestComment); ok {
		return g.pullRequestCommentGetter.GetField(v, field)
	}
	if v, ok := comment.(*github.Comment); ok {
		field = strings.ToUpper(field)
		if field == "BODY" {
			return ToString(v.Body)
		}
		if field == "CREATED_AT" {
			return ToString(v.CreatedAt)
		}
	}
	return ""
}

func (r *Renderer) RenderComments(comments []*github.Comment, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(comments)
	}

	if len(comments) == 0 {
		r.WriteLine("no comments found")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"BODY", "CREATED_AT"}
	}

	commentGetter := NewCommentFieldGetters()
	table := r.newTableWriter(headers)
	for _, comment := range comments {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = commentGetter.GetField(comment, header)
		}
		table.Append(row)
	}
	return table.Render()
}

func (r *Renderer) RenderAnyComments(comments []any, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(comments)
	}

	if len(comments) == 0 {
		r.WriteLine("no comments found")
		return nil
	}

	commentGetter := NewCommentFieldGetters()
	table := r.newTableWriter(headers)
	for _, comment := range comments {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = commentGetter.GetField(comment, header)
		}
		table.Append(row)
	}
	return table.Render()
}

func (r *Renderer) RenderCommentDefault(comments any) error {
	if v, ok := comments.([]*github.IssueComment); ok {
		return r.RenderIssueComments(v, nil)
	}
	if v, ok := comments.([]*github.PullRequestComment); ok {
		return r.RenderPullRequestComments(v, nil)
	}
	if v, ok := comments.([]*github.Comment); ok {
		return r.RenderComments(v, []string{"BODY"})
	}
	if v, ok := comments.([]any); ok {
		return r.RenderAnyComments(v, PullRequestCommentDefaultHeaders)
	}
	return nil
}
