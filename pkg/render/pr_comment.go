package render

import (
	"github.com/google/go-github/v73/github"
)

type prCommentFieldGetter func(comment *github.PullRequestComment) string
type prCommentFieldGetters struct {
	Func map[string]prCommentFieldGetter
}

func NewPullRequestCommentFieldGetters() *prCommentFieldGetters {
	return &prCommentFieldGetters{
		Func: map[string]prCommentFieldGetter{
			"ID": func(comment *github.PullRequestComment) string {
				return ToString(comment.ID)
			},
			"BODY": func(comment *github.PullRequestComment) string {
				return ToString(comment.Body)
			},
			"USER": func(comment *github.PullRequestComment) string {
				return ToString(comment.User.Login)
			},
			"IN_REPLY_TO": func(comment *github.PullRequestComment) string {
				return ToString(comment.InReplyTo)
			},
			"PATH": func(comment *github.PullRequestComment) string {
				return ToString(comment.Path)
			},
			"DIFF_HUNK": func(comment *github.PullRequestComment) string {
				return ToString(comment.DiffHunk)
			},
			"PULL_REQUEST_REVIEW_ID": func(comment *github.PullRequestComment) string {
				return ToString(comment.PullRequestReviewID)
			},
			"COMMIT_ID": func(comment *github.PullRequestComment) string {
				return ToString(comment.CommitID)
			},
			"ORIGINAL_COMMIT_ID": func(comment *github.PullRequestComment) string {
				return ToString(comment.OriginalCommitID)
			},
			"SUBJECT_TYPE": func(comment *github.PullRequestComment) string {
				return ToString(comment.SubjectType)
			},
			"POSITION": func(comment *github.PullRequestComment) string {
				return ToString(comment.Position)
			},
			"ORIGINAL_POSITION": func(comment *github.PullRequestComment) string {
				return ToString(comment.OriginalPosition)
			},
			"LINE": func(comment *github.PullRequestComment) string {
				return ToString(comment.Line)
			},
			"ORIGINAL_LINE": func(comment *github.PullRequestComment) string {
				return ToString(comment.OriginalLine)
			},
			"START_LINE": func(comment *github.PullRequestComment) string {
				return ToString(comment.StartLine)
			},
			"ORIGINAL_START_LINE": func(comment *github.PullRequestComment) string {
				return ToString(comment.OriginalStartLine)
			},
			"CREATED_AT": func(comment *github.PullRequestComment) string {
				return ToString(comment.CreatedAt)
			},
			"UPDATED_AT": func(comment *github.PullRequestComment) string {
				return ToString(comment.UpdatedAt)
			},
			"URL": func(comment *github.PullRequestComment) string {
				return ToString(comment.URL)
			},
			"HTML_URL": func(comment *github.PullRequestComment) string {
				return ToString(comment.HTMLURL)
			},
			"PULL_REQUEST_ID": func(comment *github.PullRequestComment) string {
				return ToString(comment.PullRequestURL)
			},
		},
	}
}

func (g *prCommentFieldGetters) GetField(comment *github.PullRequestComment, field string) string {
	if getter, ok := g.Func[field]; ok {
		return getter(comment)
	}
	return ""
}

func (r *Renderer) RenderPullRequestComments(comments []*github.PullRequestComment, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(comments)
		return
	}
	getter := NewPullRequestCommentFieldGetters()
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

var PullRequestCommentDefaultHeaders = []string{"ID", "BODY", "USER", "PATH", "HTML_URL"}

func (r *Renderer) RenderPullRequestCommentsDefault(comments []*github.PullRequestComment) {
	r.RenderPullRequestComments(comments, PullRequestCommentDefaultHeaders)
}
