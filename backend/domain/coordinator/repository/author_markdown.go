package repository

import (
	"fmt"
	"strings"

	"coordinator/model"
)

func rewriteAuthorMarkdown(content string, members []model.TeamPerson) string {
	if strings.TrimSpace(content) == "" {
		content = defaultAuthorMarkdown()
	}

	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines)+len(members)+2)
	tableWritten := false

	for i := 0; i < len(lines); {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "| alias |") && !tableWritten {
			out = append(out, "| alias | name |")
			out = append(out, "|-------|------|")
			for _, member := range members {
				out = append(out, fmt.Sprintf("| %s | %s |", member.Alias, member.Name))
			}
			i++
			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "|") {
				i++
			}
			tableWritten = true
			continue
		}
		out = append(out, lines[i])
		i++
	}

	if !tableWritten {
		if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		out = append(out, "| alias | name |")
		out = append(out, "|-------|------|")
		for _, member := range members {
			out = append(out, fmt.Sprintf("| %s | %s |", member.Alias, member.Name))
		}
		out = append(out, "")
	}

	text := strings.Join(out, "\n")
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return text
}

func defaultAuthorMarkdown() string {
	return `# Authors

Shared map of people who write docs, branches, and other artifacts that carry an **alias**. No emails here — the current person on a machine is ` + "`Common/data/.current_author`" + ` (local, not in git).

Structured roster (role, focus): **` + "`team.json`" + `**. This table is generated from the roster — edit people in Coordinator.

| alias | name |
|-------|------|
`
}

func parseCurrentAuthor(data []byte) string {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
}
