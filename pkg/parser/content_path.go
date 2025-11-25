package parser

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
)

type ContentPath struct {
	Repo *repository.Repository
	Path *string
	Ref  *string
}

func ParseContentPathFromUses(uses string) (*ContentPath, error) {
	if len(uses) == 0 {
		return nil, fmt.Errorf("uses string is empty")
	}
	if len(uses) >= 2 && uses[0] == '.' && uses[1] == '/' {
		path := uses[2:]
		return &ContentPath{
			Repo: nil,
			Path: &path,
			Ref:  nil,
		}, nil
	}
	parts := strings.SplitN(uses, "/", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid uses format: %s", uses)
	}
	last := parts[len(parts)-1]
	parts2 := strings.SplitN(last, "@", 2)
	if len(parts2) != 2 {
		return nil, fmt.Errorf("invalid uses format: %s", uses)
	}
	if len(parts) == 2 {
		return &ContentPath{
			Repo: &repository.Repository{
				Owner: parts[0],
				Name:  parts2[0],
			},
			Path: nil,
			Ref:  &parts2[1],
		}, nil
	}
	return &ContentPath{
		Repo: &repository.Repository{
			Owner: parts[0],
			Name:  parts[1],
		},
		Path: &parts2[0],
		Ref:  &parts2[1],
	}, nil
}

func ParseContentPathFromURL(htmlUrl string) (*ContentPath, error) {
	u, err := url.Parse(htmlUrl)
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(u.Path, "/", 4)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid URL format: %s", htmlUrl)
	}
	// parts[0] is empty string, parts[1] is owner, parts[2] is repo
	if len(parts) == 3 {
		return &ContentPath{
			Repo: &repository.Repository{
				Owner: parts[1],
				Name:  parts[2],
			},
			Path: nil,
			Ref:  nil,
		}, nil
	}
	paths := strings.SplitN(parts[3], "/", 2)
	if len(paths) < 2 {
		return &ContentPath{
			Repo: &repository.Repository{
				Owner: parts[1],
				Name:  parts[2],
			},
			Path: nil,
			Ref:  nil,
		}, nil
	}

	if paths[0] != "blob" && paths[0] != "tree" {
		return &ContentPath{
			Repo: &repository.Repository{
				Owner: parts[1],
				Name:  parts[2],
			},
			Path: nil,
			Ref:  nil,
		}, nil
	}

	if paths[0] == "tree" {
		return &ContentPath{
			Repo: &repository.Repository{
				Owner: parts[1],
				Name:  parts[2],
			},
			Path: nil,
			Ref:  &paths[1],
		}, nil
	}

	blob := strings.SplitN(paths[1], "/", 2)
	if len(blob) < 2 {
		return nil, fmt.Errorf("invalid URL format: %s", htmlUrl)
	}

	return &ContentPath{
		Repo: &repository.Repository{
			Owner: parts[1],
			Name:  parts[2],
		},
		Path: &blob[1],
		Ref:  &blob[0],
	}, nil
}

func ParseContentPath(s string) (*ContentPath, error) {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return ParseContentPathFromURL(s)
	}
	contentPath, err := ParseContentPathFromUses(s)
	if err == nil {
		return contentPath, nil
	}
	return &ContentPath{
		Repo: nil,
		Path: &s,
		Ref:  nil,
	}, nil
}
