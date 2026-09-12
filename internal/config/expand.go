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

// EnvLookupFunc resolves an environment variable by name. It mirrors
// os.LookupEnv: ok is false when the variable is unset.
//
// Injecting the lookup keeps expansion a pure function of its inputs, so tests
// can supply a deterministic, in-memory source instead of mutating
// process-global environment state (review finding #1). The production entry
// point ExpandString defaults to os.LookupEnv.
type EnvLookupFunc func(key string) (value string, ok bool)

// ExpandString expands ${VAR} and ${VAR:-default} expressions in the given
// string using the process environment (os.LookupEnv).
func ExpandString(s string) (string, error) {
	return ExpandStringWithLookup(s, os.LookupEnv)
}

// ExpandStringWithLookup expands ${VAR} and ${VAR:-default} expressions in the
// given string using the supplied environment lookup.
//
// Semantics: a variable that is unset or empty falls back to its default when
// one is present (an empty default `:-` yields ""), otherwise expansion fails
// with ErrUnsetVariable. Malformed ${...} syntax (an unclosed `${`) fails with
// ErrMalformedVariable.
func ExpandStringWithLookup(s string, lookup EnvLookupFunc) (string, error) {
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

		val, found := lookup(varName)
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
