package utils

import (
	"html"
	"net/url"
	"path"
	"regexp"
	"strings"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/mattn/go-mastodon"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TitleCase(text string) string {
	caser := cases.Title(language.English)
	return caser.String(text)
}

func StripTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}

// Parses an HTML string and converts it into a slice of styled segments
func ParseStatus(content string, tags []mastodon.Tag) []vaxis.Segment {
	// Create a set of known tag names for efficient lookup
	knownTagNames := make(map[string]struct{})
	for _, tag := range tags {
		knownTagNames[tag.Name] = struct{}{}
	}

	var segments []vaxis.Segment
	cursor := 0

	for cursor < len(content) {
		// Find the start of the next HTML tag '<'
		tagStart := strings.Index(content[cursor:], "<")

		// 1 Handle text before the tag
		var preText string
		if tagStart == -1 {
			// No more tags found, the rest of the string is plain text
			preText = content[cursor:]
			cursor = len(content)
		} else {
			tagStart += cursor // Adjust index to be absolute
			preText = content[cursor:tagStart]
		}

		if len(preText) > 0 {
			// Add the collected plain text as a segment
			// html.UnescapeString handles entities like &amp;
			segments = append(segments, vaxis.Segment{
				Text: html.UnescapeString(preText),
			})
		}

		if tagStart == -1 {
			break // Exit loop if no more tags
		}

		// 2 Process the tag itself
		tagEnd := strings.Index(content[tagStart:], ">")
		if tagEnd == -1 {
			// Malformed HTML, treat the rest as text and exit
			segments = append(segments, vaxis.Segment{
				Text: content[tagStart:],
			})
			break
		}
		tagEnd += tagStart // Adjust index to be absolute
		fullTag := content[tagStart : tagEnd+1]

		// Check if the tag is an anchor `<a>`
		if strings.HasPrefix(fullTag, "<a ") {
			closeTag := "</a>"
			closeTagStart := strings.Index(content[tagEnd:], closeTag)
			if closeTagStart == -1 {
				// Malformed link, skip the opening tag and continue parsing
				cursor = tagEnd + 1
				continue
			}
			closeTagStart += tagEnd // Adjust index to be absolute

			// Extract URL from the href attribute
			hrefRegex := regexp.MustCompile(`href="([^"]*)"`)
			matches := hrefRegex.FindStringSubmatch(fullTag)
			var linkURL string
			if len(matches) > 1 {
				linkURL = matches[1]
			}

			isKnownTag := false
			if len(matches) > 1 {
				parsedURL, err := url.Parse(linkURL)
				// Check if the link corresponds to a known hashtag
				if err == nil {
					tagName := path.Base(parsedURL.Path)
					if _, ok := knownTagNames[tagName]; ok {
						segments = append(segments, vaxis.Segment{
							Text: "#" + tagName,
							Style: vaxis.Style{
								Hyperlink:      linkURL,
								UnderlineStyle: vaxis.UnderlineSingle,
							},
						})
						isKnownTag = true
					}
				}
			}

			// If it's not a known hashtag, treat it as a regular link
			if !isKnownTag {
				linkContent := content[tagEnd+1 : closeTagStart]
				// Strip any inner tags (like <span>) to get the clean text
				linkText := StripTags(linkContent)
				segments = append(segments, vaxis.Segment{
					Text: html.UnescapeString(linkText),
					Style: vaxis.Style{
						Hyperlink:      linkURL,
						UnderlineStyle: vaxis.UnderlineSingle,
					},
				})
			}

			// Move the cursor past the entire processed <a>...</a> block
			cursor = closeTagStart + len(closeTag)

		} else if strings.EqualFold(fullTag, "<br>") {
			// Convert <br> and </p> tags into newlines
			segments = append(segments, vaxis.Segment{Text: "\n"})
			cursor = tagEnd + 1
		} else if strings.EqualFold(fullTag, "</p>") {
			segments = append(segments, vaxis.Segment{Text: "\n\n"})
			cursor = tagEnd + 1
		} else {
			// For any other tag, just skip over it
			cursor = tagEnd + 1
		}
	}

	return segments
}
