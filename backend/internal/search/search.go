// Package search implements typo- and partial-name-tolerant matching in memory.
// Firestore has no full-text search, so product names are indexed here.
// Scoring: exact/prefix/substring boosts + trigram (Jaccard) similarity.
package search

import (
	"sort"
	"strings"
	"unicode"
)

type Item struct {
	ID   string
	Name string
}

type Result struct {
	ID    string
	Score float64
}

type Index struct {
	items []entry
}

type entry struct {
	id    string
	norm  string
	grams map[string]struct{}
}

func Build(items []Item) *Index {
	idx := &Index{items: make([]entry, 0, len(items))}
	for _, it := range items {
		n := normalize(it.Name)
		idx.items = append(idx.items, entry{id: it.ID, norm: n, grams: trigrams(n)})
	}
	return idx
}

// Query returns matches best-first. Results below minScore are dropped.
func (i *Index) Query(q string, limit int) []Result {
	const minScore = 0.25
	nq := normalize(q)
	if nq == "" {
		return nil
	}
	qg := trigrams(nq)
	var out []Result
	for _, e := range i.items {
		s := score(nq, qg, e)
		if s >= minScore {
			out = append(out, Result{ID: e.id, Score: s})
		}
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].Score > out[b].Score })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func score(nq string, qg map[string]struct{}, e entry) float64 {
	s := jaccard(qg, e.grams)
	switch {
	case e.norm == nq:
		s += 2
	case strings.HasPrefix(e.norm, nq):
		s += 1.2
	case strings.Contains(e.norm, nq):
		s += 0.8
	}
	// every query word is a prefix of some name word ("witch 3" -> "the witcher 3")
	if allWordsPrefix(nq, e.norm) {
		s += 0.6
	}
	return s
}

func allWordsPrefix(q, name string) bool {
	nw := strings.Fields(name)
	for _, qw := range strings.Fields(q) {
		ok := false
		for _, w := range nw {
			if strings.HasPrefix(w, qw) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

func normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func trigrams(s string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, w := range strings.Fields(s) {
		p := []rune("  " + w + " ")
		for i := 0; i+3 <= len(p); i++ {
			out[string(p[i:i+3])] = struct{}{}
		}
	}
	return out
}

func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter := 0
	for k := range a {
		if _, ok := b[k]; ok {
			inter++
		}
	}
	return float64(inter) / float64(len(a)+len(b)-inter)
}
