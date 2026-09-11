package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	// ErrUnsetVariable indicates a referenced environment variable without a default is unset or empty.
	ErrUnsetVariable = errors.New("referenced environment variable is unset or empty")

	// ErrMalformedVariable indicates an unclosed or malformed ${...} expression.
	ErrMalformedVariable = errors.New("malformed variable expansion expression")

	varRegex = regexp.MustCompile(`\$\{([a-zA-Z_][a-zA-Z0-9_]*)(?::-([^}]*))?\}`)
)

// ExpandString expands ${VAR} and ${VAR:-default} expressions in the given string using environment variables.
func ExpandString(s string) (string, error) {
	if !strings.Contains(s, "${") {
		return s, nil
	}

	matches := varRegex.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 && strings.Contains(s, "${") {
		return "", ErrMalformedVariable
	}

	var result strings.Builder
	lastIndex := 0

	for _, m := range matches {
		matchStart, matchEnd := m[0], m[1]
		between := s[lastIndex:matchStart]
		if strings.Contains(between, "${") {
			return "", ErrMalformedVariable
		}

		result.WriteString(between)

		varName := s[m[2]:m[3]]
		hasDefault := m[4] >= 0
		var defaultVal string
		if hasDefault {
			defaultVal = s[m[4]:m[5]]
		}

		val, found := os.LookupEnv(varName)
		if !found || val == "" {
			if hasDefault {
				result.WriteString(defaultVal)
			} else {
				return "", fmt.Errorf("%w: %s", ErrUnsetVariable, varName)
			}
		} else {
			result.WriteString(val)
		}

		lastIndex = matchEnd
	}

	trailing := s[lastIndex:]
	if strings.Contains(trailing, "${") {
		return "", ErrMalformedVariable
	}
	result.WriteString(trailing)

	return result.String(), nil
}
