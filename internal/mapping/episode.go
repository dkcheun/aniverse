package mapping

import (
	"aniverse/internal/provider/anilist"
	"aniverse/internal/provider/gogoanime"
	"aniverse/internal/provider/mal"
	"aniverse/internal/types"
	"fmt"
	"log"
)

// EpisodesResult contains the combined list of sub and dub episodes.
type EpisodesResult struct {
	Episodes []types.Episode
}

func GetEpisodes(anilistID string) (*EpisodesResult, error) {
	gogoAnimeMap, err := GetGogoAnimeMap(anilistID)
	if err != nil {
		return nil, err
	}

	gogoAnimeProvider := gogoanime.NewGogoAnime()
	aniListProvider := anilist.NewAniListBase()
	animeInfo, err := aniListProvider.GetMedia(anilistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get anime info from AniList: %v", err)
	}

	var subEpisodes []types.Episode
	var dubEpisodes []types.Episode

	// Fetch subbed episodes
	if gogoAnimeMap.Sub != nil {
		subEpisodes, err = gogoAnimeProvider.FetchEpisodes(gogoAnimeMap.Sub.ID)
		if err != nil {
			return nil, err
		}
	}

	// Fetch dubbed episodes
	if gogoAnimeMap.Dub != nil {
		dubEpisodes, err = gogoAnimeProvider.FetchEpisodes(gogoAnimeMap.Dub.ID)
		if err != nil {
			return nil, err
		}
	}

	// Fetch episode titles from MAL
	var malEpisodes map[int]string
	if animeInfo.IDMal != "" {
		mal := mal.NewMyAnimeList()
		malEpisodes, err = mal.GetEpisodeTitles(animeInfo.IDMal, animeInfo.Title.English, 0)
		if err != nil {
			log.Printf("Error scraping episode titles from MyAnimeList for ID %s: %v", animeInfo.IDMal, err)
		}
	}

	// Combine episodes into a single list, matching sub and dub by episode number
	episodeMap := make(map[int]types.Episode)

	// Add subbed episodes to the map
	for _, subEp := range subEpisodes {
		ep := types.Episode{
			ID:           subEp.ID,
			Number:       subEp.Number,
			EpisodeTitle: subEp.EpisodeTitle,
			IsFiller:     subEp.IsFiller,
			Img:          subEp.Img,
			HasDub:       false,        // No dub by default
			Source:       subEp.Source, // Initialize with subbed source
		}

		// Update title from MAL if available
		if title, ok := malEpisodes[subEp.Number]; ok {
			ep.EpisodeTitle = title
		}

		episodeMap[subEp.Number] = ep
	}

	// Add dubbed episodes and combine them with the subbed episodes
	for _, dubEp := range dubEpisodes {
		if ep, exists := episodeMap[dubEp.Number]; exists {
			ep.HasDub = true
			// Merge sub and dub links for each quality
			for i, subSource := range ep.Source.Sources {
				for _, dubSource := range dubEp.Source.Sources {
					if subSource.Name == dubSource.Name {
						// Add the dub URL to the matching sub quality
						ep.Source.Sources[i].DubURL = dubSource.SubURL
					}
				}
			}
			episodeMap[dubEp.Number] = ep
		} else {
			// If subbed episode does not exist, create a new one with only dub
			ep := types.Episode{
				ID:           dubEp.ID,
				Number:       dubEp.Number,
				EpisodeTitle: dubEp.EpisodeTitle,
				IsFiller:     dubEp.IsFiller,
				Img:          dubEp.Img,
				HasDub:       true,         // This is a dub episode
				Source:       dubEp.Source, // Initialize with dubbed source
			}

			// Update title from MAL if available
			if title, ok := malEpisodes[dubEp.Number]; ok {
				ep.EpisodeTitle = title
			}

			episodeMap[dubEp.Number] = ep
		}
	}

	// Convert map to slice
	var combinedEpisodes []types.Episode
	for _, episode := range episodeMap {
		combinedEpisodes = append(combinedEpisodes, episode)
	}

	// Return combined episodes in an EpisodesResult struct
	return &EpisodesResult{
		Episodes: combinedEpisodes,
	}, nil
}
