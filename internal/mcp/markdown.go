package mcp

import (
	"strings"
)

// markdownToHTML converts simple markdown to HTML for TipTap editor rendering.
// Handles headings, paragraphs, and list items.
func markdownToHTML(md string) string {
	lines := strings.Split(md, "\n")
	var html strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		switch {
		case trimmed == "":
			// empty line
		case strings.HasPrefix(trimmed, "### "):
			html.WriteString("<h3>" + trimmed[4:] + "</h3>")
		case strings.HasPrefix(trimmed, "## "):
			html.WriteString("<h2>" + trimmed[3:] + "</h2>")
		case strings.HasPrefix(trimmed, "# "):
			html.WriteString("<h1>" + trimmed[2:] + "</h1>")
		case strings.HasPrefix(trimmed, "- "):
			html.WriteString("<ul><li>" + trimmed[2:] + "</li></ul>")
		case strings.HasPrefix(trimmed, "* "):
			html.WriteString("<ul><li>" + trimmed[2:] + "</li></ul>")
		default:
			html.WriteString("<p>" + trimmed + "</p>")
		}
	}

	return html.String()
}
