package mapping

import (
	"aniverse/internal/provider/anilist"
	"aniverse/internal/provider/gogoanime"
	"aniverse/internal/types"
	"aniverse/internal/util"
	"fmt"
	"strings"
)

type GogoAnimeMapResult struct {
	Sub *types.AnimeInfo
	Dub *types.AnimeInfo
}

type AniListIDResponse struct {
	Data struct {
		Media struct {
			IDMal string `json:"idMal"`
		} `json:"Media"`
	} `json:"data"`
}

func GetGogoAnimeMap(anilistID string) (*GogoAnimeMapResult, error) {
	// Initialize providers
	aniListProvider := anilist.NewAniListBase()
	gogoAnimeProvider := gogoanime.NewGogoAnime()

	// Get anime title from AniList
	animeInfo, err := aniListProvider.GetMedia(anilistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get anime info from AniList: %v", err)
	}

	title := animeInfo.Title

	// Create dubbed title variations
	titleDub := types.Title{
		Romaji:  title.Romaji + " dub",
		English: title.English + " dub",
		Native:  title.Native + " dub",
	}

	// Sanitize the titles
	searchTitle := util.Sanitize(title.English)
	if searchTitle == "" {
		searchTitle = util.Sanitize(title.Romaji)
	}

	// Search GogoAnime with sanitized title
	searchResults, err := gogoAnimeProvider.Search(searchTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to search GogoAnime: %v", err)
	}

	// Collect titles from search results
	var (
		gogoTitles    []string
		gogoDubTitles []string
	)

	for _, result := range searchResults {
		gogoTitle := result.Title.Romaji
		gogoTitles = append(gogoTitles, gogoTitle)
		if strings.Contains(strings.ToLower(gogoTitle), "(dub)") {
			gogoDubTitles = append(gogoDubTitles, gogoTitle)
		}
	}

	// Find the best matching titles
	bestSubTitle := util.FindOriginalTitle(title, gogoTitles)
	bestDubTitle := util.FindOriginalTitle(titleDub, gogoDubTitles)

	// Find the corresponding AnimeInfo objects
	var (
		bestSub *types.AnimeInfo
		bestDub *types.AnimeInfo
	)

	for _, result := range searchResults {
		if result.Title.Romaji == bestSubTitle {
			bestSub = &result
		}
		if result.Title.Romaji == bestDubTitle {
			bestDub = &result
		}
	}

	return &GogoAnimeMapResult{
		Sub: bestSub,
		Dub: bestDub,
	}, nil
}
