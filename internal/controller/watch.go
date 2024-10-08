package controller

import (
	"aniverse/internal/mapping"
	"aniverse/internal/types"
	"aniverse/view"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (provider *BaseController) WatchEpisode(c *fiber.Ctx) error {
	animeID := c.Query("id")
	episodeNumStr := c.Query("ep")

	if animeID == "" || episodeNumStr == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing 'id' or 'ep' query parameter")
	}

	episodeNum, err := strconv.Atoi(episodeNumStr)
	if err != nil || episodeNum < 1 {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid 'ep' parameter. It should be a positive integer.")
	}

	// Map AniList ID to GogoAnime IDs
	mappingResult, err := mapping.GetGogoAnimeMap(animeID)
	if err != nil {
		log.Printf("Error mapping AniList ID %s: %v", animeID, err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to map AniList ID to GogoAnime IDs.")
	}

	// Choose Subbed or Dubbed version (defaulting to Subbed)
	var gogoAnimeID string
	version := "sub" // default

	if mappingResult.Sub != nil {
		gogoAnimeID = mappingResult.Sub.ID
	} else if mappingResult.Dub != nil {
		gogoAnimeID = mappingResult.Dub.ID
		version = "dub"
	} else {
		return c.Status(fiber.StatusNotFound).SendString("No GogoAnime mapping found for this anime.")
	}

	log.Printf("Selected GogoAnime ID: %s (Version: %s)", gogoAnimeID, version)

	// Fetch episodes for the selected GogoAnime ID
	episodes, err := provider.gogoanime.FetchEpisodes("/category/" + gogoAnimeID)
	if err != nil {
		log.Printf("Error fetching episodes for GogoAnime ID %s: %v", gogoAnimeID, err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to fetch episodes.")
	}

	// Find the episode with the specified episode number
	var targetEpisode types.Episode
	found := false
	for _, ep := range episodes {
		if ep.Number == episodeNum {
			targetEpisode = ep
			found = true
			break
		}
	}

	if !found {
		return c.Status(fiber.StatusNotFound).SendString(fmt.Sprintf("Episode number %d not found.", episodeNum))
	}

	log.Printf("Found Episode: %s (Number: %d)", targetEpisode.ID, targetEpisode.Number)

	baseURL := strings.TrimSpace(provider.gogoanime.URL())
	epID := strings.TrimSpace(targetEpisode.ID)

	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		log.Printf("Error parsing base URL: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	parsedEpisode, err := url.Parse(epID)
	if err != nil {
		log.Printf("Error parsing episode ID: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Construct the GogoAnime episode URL
	episodeURL := parsedBase.ResolveReference(parsedEpisode).String()

	log.Printf("Constructed Episode URL: %s", episodeURL)

	// Fetch the streaming link from the episode page
	streamingLink, err := provider.gogoanime.GetSource(episodeURL)
	if err != nil {
		log.Printf("Error getting streaming link for URL %s: %v", episodeURL, err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to retrieve streaming link.")
	}

	log.Printf("Streaming Link: %s", streamingLink)

	// Use the extractor to get the video source
	source, err := provider.extractor.Extract(streamingLink)
	if err != nil {
		log.Printf("Error extracting video sources from streaming link %s: %v", streamingLink, err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to extract video sources.")
	}

	// Assign extracted source to the Episode's Source field.
	targetEpisode.Source = *source

	animeInfo, err := provider.anilist.GetMedia(animeID)
	if err != nil {
		log.Printf("Error fetching data from Anilist for ID %s: %v", animeID, err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to fetch anime data.")
	}

	targetEpisode.ID = animeInfo.ID
	targetEpisode.Anime = animeInfo.Title

	// Fetch episode titles from MAL if available (using idMal)
	if animeInfo.IDMal != "" {
		malEpisodes, err := provider.myanimelist.GetEpisodeTitles(animeInfo.IDMal, animeInfo.Title.English, episodeNum)
		if err != nil {
			log.Printf("Error fetching episode titles from MyAnimeList for ID %s: %v", animeInfo.IDMal, err)
		} else {
			// Update title from MAL if available
			if title, ok := malEpisodes[episodeNum]; ok {
				targetEpisode.EpisodeTitle = title
			}
		}
	}

	if targetEpisode.Source.Headers == nil {
		source.Headers = make(map[string]string)
	}

	targetEpisode.Source.Headers["Version"] = version

	log.Printf("Video Source Extracted: %+v", source)

	// Set headers and render the view
	c.Set("Content-Type", "text/html")
	if err := view.Watch(&targetEpisode).Render(c.Context(), c.Response().BodyWriter()); err != nil {
		log.Printf("Error rendering view: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to render view.")
	}

	// Just to verify JSON DATA
	// return c.Status(200).JSON(targetEpisode)
	return nil
}
