// Package markdown provides utilities for parsing Markdown and HTML markup.
package markdown

import (
	"regexp"
	"strings"
)

var (
	// imageRe matches ![alt](url) — GitHub inserts this when a file is attached.
	imageRe = regexp.MustCompile(`!\[([^\]]*)\]\((https?://[^\s)]+)\)`)
	// linkRe matches [text.ext](url) where the link text looks like a filename.
	linkRe = regexp.MustCompile(`\[([^\]\n]+\.[A-Za-z0-9]{2,5})\]\((https?://[^\s)]+)\)`)
	// imgSrcRe matches <img ... src="url" .../> in any attribute order.
	imgSrcRe = regexp.MustCompile(`(?i)<img\s[^>]*\bsrc=["'](https?://[^"']+)["'][^>]*>`)
	imgAltRe = regexp.MustCompile(`(?i)\balt=["']([^"']+)["']`)
)

// ExtractFilenameHints scans Markdown/HTML text for image and link syntax and
// returns a map from URL to the suggested filename.
//
// GitHub embeds the original upload filename in the alt text or link text, e.g.:
//
//	![screenshot.png](https://github.com/user-attachments/assets/<UUID>)
//	[demo.mp4](https://github.com/user-attachments/assets/<UUID>)
//
// The returned map keys are the raw URLs found in the text; values are the
// suggested filenames derived from alt text, link text, or HTML alt attributes.
// When multiple patterns match the same URL, the first match (highest priority)
// wins:
//  1. Alt text from ![alt](url) — highest fidelity for images.
//  2. Link text from [filename.ext](url) — for files without a Markdown preview.
//  3. alt= attribute from <img src="url" alt="name"> HTML tags.
func ExtractFilenameHints(text string) map[string]string {
	hints := make(map[string]string)

	// 1. ![alt](url)
	for _, m := range imageRe.FindAllStringSubmatch(text, -1) {
		alt, u := strings.TrimSpace(m[1]), m[2]
		if _, exists := hints[u]; !exists && alt != "" {
			hints[u] = alt
		}
	}

	// 2. [filename.ext](url)
	for _, m := range linkRe.FindAllStringSubmatch(text, -1) {
		linkText, u := strings.TrimSpace(m[1]), m[2]
		if _, exists := hints[u]; !exists && linkText != "" {
			hints[u] = linkText
		}
	}

	// 3. <img src="url" alt="name">
	for _, m := range imgSrcRe.FindAllStringSubmatch(text, -1) {
		imgTag, u := m[0], m[1]
		if _, exists := hints[u]; exists {
			continue
		}
		if altM := imgAltRe.FindStringSubmatch(imgTag); altM != nil {
			if alt := strings.TrimSpace(altM[1]); alt != "" {
				hints[u] = alt
			}
		}
	}

	return hints
}
