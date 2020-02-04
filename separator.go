package qry

import "strings"

// SeparatorConfig TODO
type SeparatorConfig struct {
	Fields, Values func(string) []string
	KeyVals        func(string) (string, string)
}

type separatorSet map[rune]struct{}

func newSeparatorSet(runes ...rune) separatorSet {
	res := make(separatorSet, len(runes))
	for _, r := range runes {
		res[r] = struct{}{}
	}
	return res
}

func (ss separatorSet) check(r rune) bool {
	_, res := ss[r]
	return res
}

func (ss separatorSet) Split(s string) []string {
	return strings.FieldsFunc(s, ss.check)
}

func (ss separatorSet) Pair(s string) (string, string) {
	if idx := strings.IndexFunc(s, ss.check); idx >= 0 {
		return s[:idx], s[idx+1:]
	}
	return s, ""
}
