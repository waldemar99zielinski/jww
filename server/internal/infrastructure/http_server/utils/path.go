package utils

import "strings"

type PathMatch struct {
	Matched bool
	Params  map[string]string
}

// MatchPath matches path against a route-like pattern. A segment enclosed in
// braces captures the corresponding path segment as a named parameter.
func MatchPath(pattern, path string) PathMatch {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")
	if len(patternParts) != len(pathParts) {
		return PathMatch{}
	}

	params := make(map[string]string)
	for i, patternPart := range patternParts {
		name, isParam := strings.CutPrefix(patternPart, "{")
		if isParam {
			name, isParam = strings.CutSuffix(name, "}")
		}
		if isParam && name != "" {
			if pathParts[i] == "" {
				return PathMatch{}
			}
			params[name] = pathParts[i]
			continue
		}
		if patternPart != pathParts[i] {
			return PathMatch{}
		}
	}

	return PathMatch{Matched: true, Params: params}
}
