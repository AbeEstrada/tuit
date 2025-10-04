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
	var lines []string
	if maxWidth <= 0 {
		return []string{text}
	}
	words := strings.Fields(text)
	var currentLine strings.Builder
	for _, word := range words {
		if len(word) > maxWidth {
			if currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}
			lines = append(lines, word)
			continue
		}
		testLine := currentLine.String()
		if testLine != "" {
			testLine += " "
		}
		testLine += word
		if len(testLine) > maxWidth {
			if currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}
			currentLine.WriteString(word)
		} else {
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		}
	}
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}
	return lines
}
