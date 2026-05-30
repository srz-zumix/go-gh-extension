package gh

import (
	"regexp"
)

// BuildAssetURLPatterns returns compiled regular expressions matching GitHub-hosted
// asset URLs for the given host.
//
// For github.com the patterns also cover the separate CDN hostnames
// (user-images.githubusercontent.com and private-user-images.githubusercontent.com).
// For GitHub Enterprise Server instances only the GHES host itself is used, because
// assets are served from the same hostname rather than a separate CDN.
func BuildAssetURLPatterns(host string) []*regexp.Regexp {
	escapedHost := regexp.QuoteMeta(host)
	patterns := []*regexp.Regexp{
		// https://<host>/user-attachments/assets/...
		regexp.MustCompile(`https://` + escapedHost + `/user-attachments/assets/[^\s)<>"]+`),
		// https://<host>/<owner>/<repo>/assets/...
		regexp.MustCompile(`https://` + escapedHost + `/[^/\s]+/[^/\s]+/assets/[^\s)<>"]+`),
	}
	// GitHub.com additionally serves images from a separate CDN hostname.
	if host == "github.com" {
		patterns = append(patterns,
			regexp.MustCompile(`https://user-images\.githubusercontent\.com/[^\s)<>"]+`),
			regexp.MustCompile(`https://private-user-images\.githubusercontent\.com/[^\s)<>"]+`),
		)
	}
	return patterns
}
