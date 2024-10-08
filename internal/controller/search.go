package controller

import (
	"aniverse/internal/mapping"
	"aniverse/internal/types"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func (provider *BaseController) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing 'q' parameter.")
	}

	// Optional
	pageParam := c.Query("page")
	perPageParam := c.Query("per_page")

	page := 1
	perPage := 10

	if pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err != nil {
			page = p
		}
	}

	if perPageParam != "" {
		if pp, err := strconv.Atoi(perPageParam); err == nil {
			perPage = pp
		}
	}

	// Check if the query is a numeric ID
	if id, err := strconv.Atoi(query); err == nil {
		// Query is numeric, search by ID
		result, err := provider.anilist.GetMedia(fmt.Sprintf("%d", id))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving anime by ID: " + err.Error())
		}
		return c.Status(fiber.StatusOK).JSON(result)
	}

	results, err := provider.anilist.Search(query, types.TypeAnime, provider.anilist.Formats(), page, perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(&results)
}

func (provider *BaseController) GetAnimeInfo(c *fiber.Ctx) error {
	// Get the 'id' parameter from the route
	id := c.Query("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing 'id' parameter")
	}

	// Fetch anime info from AniList using the provided AniListBase client
	info, err := provider.anilist.GetMedia(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error fetching anime info: " + err.Error())
	}
	// Fetch episodes from GogoAnime (or another provider)
	episodesResult, err := mapping.GetEpisodes(id) // GetEpisodes returns *EpisodesResult
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error fetching episodes: " + err.Error())
	}

	// Make sure the episodes are correctly assigned
	if episodesResult != nil {
		info.Episodes = mergeEpisodes(info.Episodes, episodesResult.Episodes)
	}
	// Return the populated AnimeInfo with episodes as JSON
	return c.Status(fiber.StatusOK).JSON(info)
}

func (provider *BaseController) SearchGogoAnime(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing 'q' parameter")
	}

	results, err := provider.gogoanime.Search(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(results)
}

// mergeEpisodes merges episodes fetched from AniList and GogoAnime (or other providers).
func mergeEpisodes(animeListEpisodes, providerEpisodes []types.Episode) []types.Episode {
	// Create a map to combine episodes by episode number
	episodeMap := make(map[int]types.Episode)

	// Add episodes from AniList to the map
	for _, ep := range animeListEpisodes {
		episodeMap[ep.Number] = ep
	}

	// Merge episodes from the provider, adding sources where applicable
	for _, ep := range providerEpisodes {
		if existingEp, found := episodeMap[ep.Number]; found {
			// If the episode exists, update the source and dub information
			existingEp.HasDub = ep.HasDub
			existingEp.Source = ep.Source // Replace or merge source info as needed
			episodeMap[ep.Number] = existingEp
		} else {
			// If not found, add the episode directly
			episodeMap[ep.Number] = ep
		}
	}

	// Convert the map back to a slice
	var mergedEpisodes []types.Episode
	for _, ep := range episodeMap {
		mergedEpisodes = append(mergedEpisodes, ep)
	}

	return mergedEpisodes
}
