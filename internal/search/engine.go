// Package search ranks entries from any number of providers against
// the text typed in the bar.
package search

import (
	"sort"
	"strings"

	"github.com/lucas77x/laucha/internal/launcher"
)

// Provider supplies entries from one source: applications, files, etc.
type Provider interface {
	Entries() []launcher.Entry
}

type Engine struct {
	providers []Provider
}

func NewEngine(providers ...Provider) *Engine {
	return &Engine{providers: providers}
}

// Query returns up to limit entries ranked by match quality.
func (e *Engine) Query(query string, limit int) []launcher.Entry {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}

	type scored struct {
		entry launcher.Entry
		score int
	}
	var matches []scored
	for _, p := range e.providers {
		for _, entry := range p.Entries() {
			if s := score(strings.ToLower(entry.Name), q); s > 0 {
				matches = append(matches, scored{entry, s})
			}
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return matches[i].entry.Name < matches[j].entry.Name
	})
	if len(matches) > limit {
		matches = matches[:limit]
	}

	results := make([]launcher.Entry, len(matches))
	for i, m := range matches {
		results[i] = m.entry
	}
	return results
}

// score ranks name against q: exact beats prefix beats word prefix
// beats substring beats subsequence. Zero means no match.
func score(name, q string) int {
	switch {
	case name == q:
		return 100
	case strings.HasPrefix(name, q):
		return 90
	case hasWordPrefix(name, q):
		return 80
	case strings.Contains(name, q):
		return 60
	case isSubsequence(name, q):
		return 30
	default:
		return 0
	}
}

func hasWordPrefix(name, q string) bool {
	words := strings.FieldsFunc(name, func(r rune) bool {
		return r == ' ' || r == '-' || r == '_' || r == '.'
	})
	for _, w := range words {
		if strings.HasPrefix(w, q) {
			return true
		}
	}
	return false
}

func isSubsequence(name, q string) bool {
	runes := []rune(q)
	i := 0
	for _, r := range name {
		if i < len(runes) && r == runes[i] {
			i++
		}
	}
	return i == len(runes)
}
