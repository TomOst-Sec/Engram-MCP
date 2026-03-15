package conventions

import "regexp"

var (
	snakeCaseRe  = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)
	camelCaseRe  = regexp.MustCompile(`^[a-z][a-zA-Z0-9]*$`)
	pascalCaseRe = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)
)

// IsSnakeCase returns true if the name follows snake_case convention.
func IsSnakeCase(name string) bool {
	return snakeCaseRe.MatchString(name)
}

// IsCamelCase returns true if the name follows camelCase convention.
func IsCamelCase(name string) bool {
	return camelCaseRe.MatchString(name) && !snakeCaseRe.MatchString(name)
}

// IsPascalCase returns true if the name follows PascalCase convention.
func IsPascalCase(name string) bool {
	return pascalCaseRe.MatchString(name)
}

// DetectNamingStyle returns the dominant naming style and confidence for a list of names.
// Returns ("", 0) if no clear pattern is detected (<60% consistency).
func DetectNamingStyle(names []string) (style string, confidence float64) {
	if len(names) == 0 {
		return "", 0
	}

	counts := map[string]int{
		"snake_case":  0,
		"camelCase":   0,
		"PascalCase":  0,
	}

	for _, name := range names {
		switch {
		case IsSnakeCase(name):
			counts["snake_case"]++
		case IsCamelCase(name):
			counts["camelCase"]++
		case IsPascalCase(name):
			counts["PascalCase"]++
		}
	}

	total := len(names)
	bestStyle := ""
	bestCount := 0
	for style, count := range counts {
		if count > bestCount {
			bestCount = count
			bestStyle = style
		}
	}

	conf := float64(bestCount) / float64(total)
	if conf < 0.6 {
		return "", 0
	}
	return bestStyle, conf
}
