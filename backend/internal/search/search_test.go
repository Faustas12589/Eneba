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

// Acronyms gamers type: first letters of the words, numbers kept whole,
// Roman numerals as letters or digits, "the" optional, apostrophes ignored.
func TestQueryAcronyms(t *testing.T) {
	idx := Build([]Item{
		{"gta5", "Grand Theft Auto V"},
		{"gta6", "Grand Theft Auto VI (PS5)"},
		{"rdr2", "Red Dead Redemption 2"},
		{"bg3", "Baldur's Gate 3"},
		{"cod", "Call of Duty: Modern Warfare III"},
		{"tw3", "The Witcher 3: Wild Hunt GOTY"},
		{"fh5", "Forza Horizon 5"},
		{"elden", "Elden Ring"},
	})
	cases := []struct {
		q, want string
	}{
		{"gtav", "gta5"},
		{"gta 5", "gta5"},
		{"GTA V", "gta5"},
		{"gta 6", "gta6"},
		{"rdr2", "rdr2"},
		{"rdr 2", "rdr2"},
		{"bg3", "bg3"},
		{"cod", "cod"},
		{"codmw3", "cod"},
		{"mw3", "cod"}, // subtitle after the colon
		{"tw3", "tw3"},
		{"w3", "tw3"}, // without "the"
		{"fh5", "fh5"},
		{"elden", "elden"}, // normal names still win on their own words
	}
	for _, c := range cases {
		got := idx.Query(c.q, 3)
		if len(got) == 0 || got[0].ID != c.want {
			t.Errorf("Query(%q) = %+v, want first %s", c.q, got, c.want)
		}
	}
	// "gta" is a prefix of both GTA acronyms: both are found
	if got := idx.Query("gta", 5); len(got) < 2 {
		t.Errorf("Query(gta) = %+v, want both GTA games", got)
	}
}

func TestAcronyms(t *testing.T) {
	cases := map[string][]string{
		"Grand Theft Auto V":               {"gtav", "gta5"},
		"Baldur's Gate 3":                  {"bg3"},
		"The Witcher 3: Wild Hunt GOTY":    {"tw3whg", "w3whg", "whg"},
		"Call of Duty: Modern Warfare III": {"codmwiii", "codmw3", "mwiii", "mw3"},
		"Elden Ring":                       {"er"},
	}
	for name, want := range cases {
		got := acronyms(name)
		if len(got) != len(want) {
			t.Errorf("acronyms(%q) = %v, want %v", name, got, want)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("acronyms(%q) = %v, want %v", name, got, want)
				break
			}
		}
	}
}
