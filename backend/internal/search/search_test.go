package search

import "testing"

func TestQuery(t *testing.T) {
	idx := Build([]Item{
		{"1", "Elden Ring"},
		{"2", "Cyberpunk 2077"},
		{"3", "The Witcher 3: Wild Hunt GOTY"},
		{"4", "Red Dead Redemption 2"},
	})
	cases := []struct {
		q, want string
	}{
		{"elden", "1"},      // partial
		{"eldne rign", "1"}, // typos
		{"cyberpnk", "2"},   // typo
		{"witcher 3", "3"},  // partial words
		{"red dead", "4"},   // partial
		{"ELDEN RING", "1"}, // case
	}
	for _, c := range cases {
		got := idx.Query(c.q, 3)
		if len(got) == 0 || got[0].ID != c.want {
			t.Errorf("Query(%q) = %+v, want first %s", c.q, got, c.want)
		}
	}
	if got := idx.Query("zzzzqqq", 3); len(got) != 0 {
		t.Errorf("expected no results, got %+v", got)
	}
}
