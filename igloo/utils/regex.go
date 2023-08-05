package utils

import "regexp"

type (
	Regex struct {
		pattern *regexp.Regexp
	}
	RegexSubmatches struct {
		pattern    *regexp.Regexp
		submatches []string
	}
)

func (match *RegexSubmatches) Find(key string) string {
	return match.submatches[match.pattern.SubexpIndex(key)]
}

func NewRegex(pattern string) *Regex {
	return &Regex{
		pattern: regexp.MustCompile(pattern),
	}
}

func (regex *Regex) Submatch(s string) *RegexSubmatches {
	return &RegexSubmatches{submatches: regex.pattern.FindStringSubmatch(s), pattern: regex.pattern}
}
