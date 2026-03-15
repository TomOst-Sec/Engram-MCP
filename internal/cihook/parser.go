package cihook

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// CIEvent represents a parsed event from CI/CD output.
type CIEvent struct {
	Type    string   // "failure", "warning", "success", "deployment"
	Summary string   // one-line description
	Details string   // full context
	Tags    []string // auto-generated tags
	Files   []string // related files (if detectable)
}

var (
	genericErrorRe  = regexp.MustCompile(`(?i)\b(error|fail|fatal)\b`)
	genericWarnRe   = regexp.MustCompile(`(?i)\b(warn|warning)\b`)
	genericDeployRe = regexp.MustCompile(`(?i)\b(deploy|deployed|released)\b`)
	filePathRe      = regexp.MustCompile(`([a-zA-Z0-9_/.-]+\.[a-zA-Z]{1,6}):(\d+)`)
)

// ParseGeneric parses generic build output looking for ERROR/FAIL/WARNING patterns.
func ParseGeneric(input io.Reader) ([]CIEvent, error) {
	scanner := bufio.NewScanner(input)
	var events []CIEvent

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		var event *CIEvent

		if genericErrorRe.MatchString(trimmed) {
			event = &CIEvent{
				Type:    "failure",
				Summary: truncate(trimmed, 200),
				Details: trimmed,
				Tags:    []string{"ci", "generic"},
			}
		} else if genericDeployRe.MatchString(trimmed) {
			event = &CIEvent{
				Type:    "deployment",
				Summary: truncate(trimmed, 200),
				Details: trimmed,
				Tags:    []string{"ci", "generic", "deployment"},
			}
		} else if genericWarnRe.MatchString(trimmed) {
			event = &CIEvent{
				Type:    "warning",
				Summary: truncate(trimmed, 200),
				Details: trimmed,
				Tags:    []string{"ci", "generic"},
			}
		}

		if event != nil {
			// Extract file paths from the line
			if matches := filePathRe.FindAllStringSubmatch(trimmed, -1); matches != nil {
				for _, m := range matches {
					event.Files = append(event.Files, m[1])
				}
			}
			events = append(events, *event)
		}
	}

	return events, scanner.Err()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
