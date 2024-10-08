package provider

import (
	"aniverse/internal/types"
)

type BaseProvider interface {
	ID() string
	URL() string
	Formats() []types.Format
	NeedsProxy() bool
	UseGoogleTranslate() bool
	Search(query string, mediaType types.MediaType, formats []types.Format, page int, perPage int) ([]types.AnimeInfo, error)
	GetMedia(id string) (*types.AnimeInfo, error)
}
