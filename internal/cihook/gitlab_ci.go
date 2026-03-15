package cihook

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var (
	gitlabErrorRe   = regexp.MustCompile(`(?i)^ERROR:\s*(.*)`)
	gitlabJobFailRe = regexp.MustCompile(`(?i)^ERROR: Job failed`)
	gitlabWarnRe    = regexp.MustCompile(`(?i)^WARNING:\s*(.*)`)
)

// ParseGitLabCI parses GitLab CI log output.
func ParseGitLabCI(input io.Reader) ([]CIEvent, error) {
	scanner := bufio.NewScanner(input)
	var events []CIEvent

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if gitlabJobFailRe.MatchString(trimmed) {
			events = append(events, CIEvent{
				Type:    "failure",
				Summary: "GitLab CI job failed",
				Details: trimmed,
				Tags:    []string{"ci", "gitlab-ci", "job-failure"},
			})
			continue
		}

		if m := gitlabErrorRe.FindStringSubmatch(trimmed); m != nil {
			event := CIEvent{
				Type:    "failure",
				Summary: truncate(m[1], 200),
				Details: trimmed,
				Tags:    []string{"ci", "gitlab-ci", "error"},
			}
			if matches := filePathRe.FindAllStringSubmatch(m[1], -1); matches != nil {
				for _, fm := range matches {
					event.Files = append(event.Files, fm[1])
				}
			}
			events = append(events, event)
			continue
		}

		if m := gitlabWarnRe.FindStringSubmatch(trimmed); m != nil {
			events = append(events, CIEvent{
				Type:    "warning",
				Summary: truncate(m[1], 200),
				Details: trimmed,
				Tags:    []string{"ci", "gitlab-ci", "warning"},
			})
			continue
		}
	}

	return events, scanner.Err()
}
