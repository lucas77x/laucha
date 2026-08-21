// Package search ranks entries from any number of providers against
// the text typed in the bar.
package search

import (
	"sort"
	"strings"
	"time"

	"github.com/lucas77x/laucha/internal/launcher"
)

// Provider supplies entries from one source: applications, files, etc.
type Provider interface {
	Entries() []launcher.Entry
}

// Usage supplies open statistics for the frecency boost: better
// matches always win, equally good matches are ordered by use.
type Usage interface {
	Stats(path string) (count int, lastOpened time.Time)
}

type Engine struct {
	providers []Provider
	usage     Usage
}

func NewEngine(providers ...Provider) *Engine {
	return &Engine{providers: providers}
}

// SetUsage enables the frecency boost; a nil Engine usage means
// ranking by match quality and name only.
func (e *Engine) SetUsage(u Usage) { e.usage = u }

// Query returns up to limit entries ranked by match quality. The
// query is split into terms; every term must match the name or the
// path, in any order, so "not nextcl" and "nextcl not" both narrow
// notas.txt down to the copy under ~/Nextcloud.
func (e *Engine) Query(query string, limit int) []launcher.Entry {
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return nil
	}

	type scored struct {
		entry launcher.Entry
		score int
	}
	var matches []scored
	for _, p := range e.providers {
		for _, entry := range p.Entries() {
			if s := scoreEntry(entry, terms); s > 0 {
				matches = append(matches, scored{entry, s})
			}
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		if e.usage != nil {
			countI, lastI := e.usage.Stats(matches[i].entry.Path)
			countJ, lastJ := e.usage.Stats(matches[j].entry.Path)
			if countI != countJ {
				return countI > countJ
			}
			if !lastI.Equal(lastJ) {
				return lastI.After(lastJ)
			}
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

// scoreEntry sums the per-term scores; a term that matches neither
// the name nor the path disqualifies the entry.
func scoreEntry(entry launcher.Entry, terms []string) int {
	name := strings.ToLower(entry.Name)
	path := strings.ToLower(entry.Path)

	total := 0
	for _, term := range terms {
		s := score(name, term)
		if ps := pathScore(path, term); ps > s {
			s = ps
		}
		if s == 0 {
			return 0
		}
		total += s
	}
	return total
}

// pathScore rewards terms that match the location rather than the
// name; it ranks between a name substring and a name subsequence.
func pathScore(path, term string) int {
	if path != "" && strings.Contains(path, term) {
		return 40
	}
	return 0
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
