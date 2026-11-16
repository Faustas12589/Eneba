// Package search implements typo- and partial-name-tolerant matching in memory.
// Firestore has no full-text search, so product names are indexed here.
// Scoring: exact/prefix/substring boosts + trigram (Jaccard) similarity,
// plus acronyms gamers type ("gta 5", "rdr2", "bg3", "cod").
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
	acr   []string // acronyms of the name, see acronyms()
}

func Build(items []Item) *Index {
	idx := &Index{items: make([]entry, 0, len(items))}
	for _, it := range items {
		n := normalize(it.Name)
		idx.items = append(idx.items, entry{id: it.ID, norm: n, grams: trigrams(n), acr: acronyms(it.Name)})
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
	cq := compact(q) // "GTA 5" -> "gta5", compared with the acronyms
	var out []Result
	for _, e := range i.items {
		s := score(nq, qg, cq, e)
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

func score(nq string, qg map[string]struct{}, cq string, e entry) float64 {
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
	return s + acronymBoost(cq, e.acr)
}

// acronymBoost: +1.5 if the query (spaces removed) is one of the name's acronyms,
// +1.0 if it starts one ("gta" -> "gta5"). Below an exact name match (+2), so a
// real name always wins. Two characters minimum, so single letters don't match everything.
func acronymBoost(cq string, acr []string) float64 {
	if len([]rune(cq)) < 2 {
		return 0
	}
	best := 0.0
	for _, a := range acr {
		if a == cq {
			return 1.5
		}
		if strings.HasPrefix(a, cq) {
			best = 1.0
		}
	}
	return best
}

// Roman numerals that appear in game titles. "i" and "x" are left out: they are
// usually a word or a console name ("Xbox Series X").
var romans = map[string]string{"ii": "2", "iii": "3", "iv": "4", "v": "5", "vi": "6", "vii": "7", "viii": "8", "ix": "9"}

// acronyms of a name, e.g. "Grand Theft Auto V" -> ["gtav", "gta5"]:
// the first letter of each word, numbers kept whole ("Red Dead Redemption 2" -> "rdr2"),
// Roman numerals once as letters and once as digits, and again without a leading
// "the" ("The Witcher 3" -> "tw3", "w3"), and for the subtitle after a colon too
// ("Call of Duty: Modern Warfare III" -> also "mw3"). Apostrophes are dropped first,
// so "Baldur's Gate 3" gives "bg3", not "bsg3". Built once per product in Build.
func acronyms(name string) []string {
	parts := []string{name}
	if i := strings.Index(name, ":"); i > 0 {
		parts = append(parts, name[i+1:])
	}
	var variants [][]string
	for _, part := range parts {
		words := strings.Fields(normalize(dropApostrophes(part)))
		variants = append(variants, words)
		if len(words) > 1 && words[0] == "the" {
			variants = append(variants, words[1:])
		}
	}
	var out []string
	for _, ws := range variants {
		for _, digits := range []bool{false, true} {
			a := initials(ws, digits)
			if len([]rune(a)) >= 2 && !contains(out, a) {
				out = append(out, a)
			}
		}
	}
	return out
}

func initials(words []string, romanAsDigits bool) string {
	var b strings.Builder
	for _, w := range words {
		if isNumber(w) {
			b.WriteString(w)
		} else if d, ok := romans[w]; ok {
			if romanAsDigits {
				b.WriteString(d)
			} else {
				b.WriteString(w)
			}
		} else {
			b.WriteRune([]rune(w)[0])
		}
	}
	return b.String()
}

// compact normalizes a query and removes the spaces: "GTA 5" -> "gta5".
func compact(q string) string {
	return strings.ReplaceAll(normalize(dropApostrophes(q)), " ", "")
}

func dropApostrophes(s string) string {
	return strings.NewReplacer("'", "", "’", "").Replace(s)
}

func isNumber(w string) bool {
	for _, r := range w {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return w != ""
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
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
