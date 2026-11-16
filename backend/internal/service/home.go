package service

// The homepage is built from the server's data only (the assignment: "the frontend only
// displays what it receives and nothing is hardcoded there"). Two parts (DECISIONS #35):
//   - HomeLayout: which blocks, in which order, and how each is drawn (GET /api/home);
//   - each block's own endpoint (GET /api/home/{id}) returns its text from HomeBlock
//     together with its items, so a block's content comes from one place.
// Texts are Lithuanian, like the rest of the UI.

// BlockRef is one homepage section in the layout.
type BlockRef struct {
	ID       string `json:"id"`                 // the block's endpoint is /api/home/{id}
	Type     string `json:"type"`               // how it is drawn: "banners", "categories" or "products"
	Personal bool   `json:"personal,omitempty"` // built per user: send login token or guest id, reload on sign-in
}

// BlockText is what a block shows around its items.
type BlockText struct {
	ID            string `json:"id"`
	Title         string `json:"title,omitempty"`
	Info          string `json:"info,omitempty"`    // the ⓘ pop-up: what the block shows, guest vs signed in
	MoreURL       string `json:"moreUrl,omitempty"` // "Rodyti visus" link, if any
	HideWhenEmpty bool   `json:"hideWhenEmpty,omitempty"`
}

// HomeLayout is the homepage, top to bottom.
var HomeLayout = []BlockRef{
	{ID: "banners", Type: "banners"},
	{ID: "categories", Type: "categories"},
	{ID: "recently-viewed", Type: "products", Personal: true},
	{ID: "recommended", Type: "products", Personal: true},
	{ID: "best-sellers", Type: "products"},
	{ID: "gift-cards", Type: "products"},
	{ID: "top-ups", Type: "products"},
	{ID: "pre-orders", Type: "products"},
}

var homeBlocks = map[string]BlockText{
	"recently-viewed": {
		Title:         "Neseniai peržiūrėta",
		Info:          "Prekės, kurias neseniai atidarei, naujausios pirmos. Kai esi prisijungęs, sąrašas saugomas tavo paskyroje ir matomas bet kuriame įrenginyje; kitu atveju jis saugomas tik šioje naršyklėje. Prisijungus naršyklės istorija perkeliama į paskyrą.",
		HideWhenEmpty: true,
	},
	"recommended": {
		Title: "Rekomenduojama tau",
		Info:  "Rekomendacijos grindžiamos tavo peržiūrėtomis, pamėgtomis ir pirktomis prekėmis: siūlome panašaus žanro, kategorijos ir platformos prekes, kurių dar neturi. Kai esi prisijungęs, rekomendacijos remiasi tavo paskyros istorija, kitu atveju – šios naršyklės. Kol istorijos nėra, rodomos perkamiausios prekės.",
	},
	"best-sellers": {
		Title:   "Perkamiausi",
		Info:    "Daugiausiai kartų parduotos prekės visoje parduotuvėje. Sąrašas vienodas visiems pirkėjams, nesvarbu, ar esi prisijungęs.",
		MoreURL: "/store",
	},
	"gift-cards": {
		Title:   "Populiariausios el. dovanų kortelės",
		Info:    "Populiariausios el. dovanų kortelės pagal pardavimus. Kodą gausi iškart po apmokėjimo. Sąrašas tas pats visiems; į krepšelį dėti gali visi, o apmokėti – tik prisijungę pirkėjai.",
		MoreURL: "/store?category=gift-cards",
	},
	"top-ups": {
		Title:   "Populiariausi papildymai",
		Info:    "Populiariausi žaidimų platformų papildymai (Steam, PlayStation, Xbox ir kt.) pagal pardavimus. Sąrašas vienodas visiems pirkėjams; apmokant reikės prisijungti.",
		MoreURL: "/store?category=gaming-cards",
	},
	"pre-orders": {
		Title: "Įsigyk iš anksto",
		Info:  "Dar neišleisti žaidimai. Užsisakyk dabar ir kodą gausi iškart; išleidimo dieną prekė savaime taps įprasta. Sąrašas vienodas visiems; apmokant reikės prisijungti.",
	},
	// banners and categories have no title: the hero and the icon row speak for themselves
}

// HomeBlock returns a block's text; blocks without text get just their id.
func HomeBlock(id string) BlockText {
	b := homeBlocks[id]
	b.ID = id
	return b
}
