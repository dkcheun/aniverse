package types

type Episode struct {
	ID           string   `json:"id"`
	Anime        Title    `json:"series"`
	Number       int      `json:"episode"`
	EpisodeTitle string   `json:"title,omitempty"`
	IsFiller     bool     `json:"isFiller"`
	Img          *string  `json:"img,omitempty"`
	HasDub       bool     `json:"hasDub"`
	Description  *string  `json:"description,omitempty"`
	Rating       *float64 `json:"rating,omitempty"`
	Source       Source   `json:"source,omitempty"`
}

// Source holds all relevant streaming information for a video episode.
type Source struct {
	Sources       []Quality         `json:"available_qualities"`
	Subtitles     []string          `json:"subtitles"`
	Audio         []string          `json:"audio"`
	IsM3U8        bool              `json:"is_m3u8"`
	Intro         EpisodeTiming     `json:"intro"`
	Outro         EpisodeTiming     `json:"outro"`
	Headers       map[string]string `json:"headers"`
	Thumbnail     string            `json:"thumbnail"`
	ThumbnailType string            `json:"thumbnailType"`
}

// Quality represents a specific video quality with its associated metadata.
type Quality struct {
	Name       string `json:"quality"`
	Bandwidth  int    `json:"bandwidth"`
	Resolution string `json:"resolution"`
	SubURL     string `json:"sub,omitempty"`
	DubURL     string `json:"dub,omitempty"`
}

// EpisodeTiming holds the timing information for intro and outro segments.
type EpisodeTiming struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type Subtitle struct {
	Language string
	URL      string
}

type AudioTrack struct {
	Language string
	Format   string
}
