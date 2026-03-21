package render

import (
	"strings"

	"github.com/google/go-github/v84/github"
)

type githubCommentFieldGetter func(comment *github.Comment) string
type githubCommentFieldGetters struct {
	Func map[string]githubCommentFieldGetter
}

func NewGitHubCommentFieldGetters() *githubCommentFieldGetters {
	return &githubCommentFieldGetters{
		Func: map[string]githubCommentFieldGetter{
			"BODY": func(comment *github.Comment) string {
				return comment.Body
			},
			"CREATED_AT": func(comment *github.Comment) string {
				return ToString(comment.CreatedAt)
			},
		},
	}
}

func (g *githubCommentFieldGetters) GetField(comment *github.Comment, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(comment)
	}
	return ""
}

type commentFieldGetters struct {
	issueCommentGetter       *issueCommentFieldGetters
	pullRequestCommentGetter *prCommentFieldGetters
	githubCommentGetter      *githubCommentFieldGetters
}

func NewCommentFieldGetters() *commentFieldGetters {
	return &commentFieldGetters{
		issueCommentGetter:       NewIssueCommentFieldGetters(),
		pullRequestCommentGetter: NewPullRequestCommentFieldGetters(),
		githubCommentGetter:      NewGitHubCommentFieldGetters(),
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
		return g.githubCommentGetter.GetField(v, field)
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

	commentGetter := NewGitHubCommentFieldGetters()
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
