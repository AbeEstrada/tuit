package utils

import (
	"strings"

	"github.com/k3a/html2text"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TitleCase(text string) string {
	caser := cases.Title(language.English)
	return caser.String(text)
}

func HTMLToPlainText(html string) string {
	return html2text.HTML2Text(html)
}

func WrapText(text string, maxWidth int) []string {
	if maxWidth <= 0 || strings.TrimSpace(text) == "" {
		return []string{text}
	}

	words := strings.Fields(text)

	var lines []string
	var currentLine strings.Builder
	currentLine.WriteString(words[0])

	for _, word := range words[1:] {
		if currentLine.Len()+1+len(word) > maxWidth {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		} else {
			currentLine.WriteString(" ")
			currentLine.WriteString(word)
		}
	}

	lines = append(lines, currentLine.String())

	return lines
}
