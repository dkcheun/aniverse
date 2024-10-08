package types

type Media struct {
	ID              int        `json:"id"`
	IDMal           int        `json:"idMal"`
	Title           Title      `json:"title"`
	Description     string     `json:"description"`
	CoverImage      Image      `json:"coverImage"`
	BannerImage     string     `json:"bannerImage"`
	IsAdult         bool       `json:"isAdult"`
	MeanScore       int        `json:"meanScore"`
	Popularity      int        `json:"popularity"`
	Format          string     `json:"format"`
	Status          string     `json:"status"`
	Episodes        int        `json:"episodes"`
	Duration        int        `json:"duration"`
	Season          string     `json:"season"`
	SeasonYear      int        `json:"seasonYear"`
	Genres          []string   `json:"genres"`
	Synonyms        []string   `json:"synonyms"`
	CountryOfOrigin string     `json:"countryOfOrigin"`
	Tags            []Tag      `json:"tags"`
	Characters      Characters `json:"characters"`
	Relations       Relations  `json:"relations"`
	Type            string     `json:"type"`
}

type Image struct {
	ExtraLarge string `json:"extraLarge"`
	Large      string `json:"large"`
	Color      string `json:"color"`
}

type Tag struct {
	Name string `json:"name"`
}

type Characters struct {
	Edges []CharacterEdge `json:"edges"`
}

type CharacterEdge struct {
	Node struct {
		Name struct {
			Full string `json:"full"`
		} `json:"name"`
		Image Image `json:"image"`
	} `json:"node"`
	VoiceActors []struct {
		Name struct {
			Full string `json:"full"`
		} `json:"name"`
		Image Image `json:"image"`
	} `json:"voiceActors"`
}

type Relations struct {
	Edges []RelationEdge `json:"edges"`
}

type RelationEdge struct {
	RelationType string `json:"relationType"`
	Node         struct {
		ID     int    `json:"id"`
		Title  Title  `json:"title"`
		Format string `json:"format"`
		Type   string `json:"type"`
	} `json:"node"`
}
