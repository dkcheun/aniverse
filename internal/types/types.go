package types

type Title struct {
	English string `json:"english"`
	Romaji  string `json:"romaji"`
	Native  string `json:"native"`
}

type Artwork struct {
	Type       string `json:"type"`
	Img        string `json:"img"`
	ProviderID string `json:"providerId"`
}

type VoiceActor struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type Character struct {
	VoiceActor *VoiceActor `json:"voiceActor,omitempty"`
	Image      string      `json:"image"`
	Name       string      `json:"name"`
}

type Relation struct {
	ID           string    `json:"id"`
	Format       Format    `json:"format"`
	RelationType string    `json:"relationType"`
	Title        Title     `json:"title"`
	Type         MediaType `json:"type"`
}

type AnimeInfo struct {
	ID              string      `json:"id"`
	IDMal           string      `json:"idMal"`
	Title           Title       `json:"title"`
	Trailer         *string     `json:"trailer,omitempty"`
	CurrentEpisode  int         `json:"currentEpisode"`
	Duration        *int        `json:"duration,omitempty"`
	CoverImage      *Image      `json:"coverImage,omitempty"`
	BannerImage     *string     `json:"bannerImage,omitempty"`
	Popularity      int         `json:"popularity"`
	Synonyms        []string    `json:"synonyms"`
	TotalEpisodes   int         `json:"totalEpisodes"`
	Episodes        []Episode   `json:"episodes"`
	Color           *string     `json:"color,omitempty"`
	Status          MediaStatus `json:"status"`
	Season          *Season     `json:"season,omitempty"`
	Genres          []string    `json:"genres"`
	Rating          *float64    `json:"rating,omitempty"`
	Description     *string     `json:"description,omitempty"`
	Format          Format      `json:"format"`
	Year            *int        `json:"year,omitempty"`
	Type            MediaType   `json:"type"`
	CountryOfOrigin *string     `json:"countryOfOrigin,omitempty"`
	Tags            []string    `json:"tags"`
	Artwork         []Artwork   `json:"artwork"`
	Relations       []Relation  `json:"relations"`
	Characters      []Character `json:"characters"`
}

type MangaInfo struct {
	ID              string      `json:"id"`
	Title           Title       `json:"title"`
	CoverImage      *string     `json:"coverImage,omitempty"`
	BannerImage     *string     `json:"bannerImage,omitempty"`
	Popularity      int         `json:"popularity"`
	Synonyms        []string    `json:"synonyms"`
	TotalChapters   int         `json:"totalChapters"`
	TotalVolumes    int         `json:"totalVolumes"`
	Color           *string     `json:"color,omitempty"`
	Status          MediaStatus `json:"status"`
	Genres          []string    `json:"genres"`
	Rating          *float64    `json:"rating,omitempty"`
	Description     *string     `json:"description,omitempty"`
	Format          Format      `json:"format"`
	Year            *int        `json:"year,omitempty"`
	Type            MediaType   `json:"type"`
	CountryOfOrigin *string     `json:"countryOfOrigin,omitempty"`
	Tags            []string    `json:"tags"`
	Artwork         []Artwork   `json:"artwork"`
	Relations       []Relation  `json:"relations"`
	Characters      []Character `json:"characters"`
	Author          *string     `json:"author,omitempty"`
	Publisher       *string     `json:"publisher,omitempty"`
}
