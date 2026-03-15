package cihook

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var (
	ghaErrorRe    = regexp.MustCompile(`##\[error\](.*)`)
	ghaWarningRe  = regexp.MustCompile(`##\[warning\](.*)`)
	ghaTestFailRe = regexp.MustCompile(`^---\s*FAIL:\s*(.+)`)
	goTestFailRe  = regexp.MustCompile(`^FAIL\s+(\S+)`)
)

// ParseGitHubActions parses GitHub Actions log output.
func ParseGitHubActions(input io.Reader) ([]CIEvent, error) {
	scanner := bufio.NewScanner(input)
	var events []CIEvent

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if m := ghaErrorRe.FindStringSubmatch(trimmed); m != nil {
			event := CIEvent{
				Type:    "failure",
				Summary: truncate(m[1], 200),
				Details: trimmed,
				Tags:    []string{"ci", "github-actions", "error"},
			}
			if matches := filePathRe.FindAllStringSubmatch(m[1], -1); matches != nil {
				for _, fm := range matches {
					event.Files = append(event.Files, fm[1])
				}
			}
			events = append(events, event)
			continue
		}

		if m := ghaWarningRe.FindStringSubmatch(trimmed); m != nil {
			events = append(events, CIEvent{
				Type:    "warning",
				Summary: truncate(m[1], 200),
				Details: trimmed,
				Tags:    []string{"ci", "github-actions", "warning"},
			})
			continue
		}

		if m := ghaTestFailRe.FindStringSubmatch(trimmed); m != nil {
			events = append(events, CIEvent{
				Type:    "failure",
				Summary: "Test failure: " + m[1],
				Details: trimmed,
				Tags:    []string{"ci", "github-actions", "test-failure"},
			})
			continue
		}

		if m := goTestFailRe.FindStringSubmatch(trimmed); m != nil {
			events = append(events, CIEvent{
				Type:    "failure",
				Summary: "Package test failure: " + m[1],
				Details: trimmed,
				Tags:    []string{"ci", "github-actions", "test-failure", m[1]},
			})
			continue
		}
	}

	return events, scanner.Err()
}
