package render

import "github.com/google/go-github/v79/github"

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
		if field == "BODY" {
			return ToString(v.Body)
		}
		if field == "CREATED_AT" {
			return ToString(v.CreatedAt)
		}
	}
	return ""
}

func (r *Renderer) RenderComments(comments []any, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(comments)
		return
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
	table.Render()
}

func (r *Renderer) RenderCommentDefault(comments any) {
	if v, ok := comments.([]*github.IssueComment); ok {
		r.RenderIssueCommentsDefault(v)
	}
	if v, ok := comments.([]*github.PullRequestComment); ok {
		r.RenderPullRequestCommentsDefault(v)
	}
	if v, ok := comments.([]any); ok {
		if _, ok := comments.([]*github.Comment); ok {
			r.RenderComments(v, []string{"BODY"})
		} else {
			r.RenderComments(v, PullRequestCommentDefaultHeaders)
		}
	}
}
