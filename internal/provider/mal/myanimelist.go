package mal

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type MyAnimeList struct {
	BaseURL string
}

func NewMyAnimeList() *MyAnimeList {
	return &MyAnimeList{
		BaseURL: "https://myanimelist.net",
	}
}

func (m *MyAnimeList) GetEpisodeTitles(malID string, animeName string, episodeNum int) (map[int]string, error) {
	// Normalize the anime name for URL safety
	normalizedAnimeName := NormalizeAnimeTitle(animeName)

	// Construct the correct URL for episode listing
	url := fmt.Sprintf("%s/anime/%s/%s/episode", m.BaseURL, malID, normalizedAnimeName)

	// Make HTTP request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch MAL episode page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MAL episode page: %v", err)
	}

	// Map to store episode number and title
	episodeTitles := make(map[int]string)

	// Extract episode titles based on the new structure from MAL
	doc.Find("tr.episode-list-data").Each(func(i int, s *goquery.Selection) {
		// Extract episode number
		episodeNumStr := strings.TrimSpace(s.Find("td.episode-number").Text())

		// Extract episode title
		title := strings.TrimSpace(s.Find("td.episode-title a").Text())

		// Use regular expression to ensure only numbers are extracted from the episode number
		re := regexp.MustCompile(`\d+`)
		episodeNumStr = re.FindString(episodeNumStr)

		// Convert the episode number string to an integer
		episodeNumParsed, err := strconv.Atoi(episodeNumStr)
		if err != nil {
			log.Printf("Failed to parse episode number: %v", err)
			return
		}

		// Store the episode number and title
		episodeTitles[episodeNumParsed] = title
	})

	if len(episodeTitles) == 0 {
		return nil, fmt.Errorf("no episode titles found on MAL page")
	}

	// If the requested episodeNum exists in the map, return just that
	if title, ok := episodeTitles[episodeNum]; ok {
		return map[int]string{episodeNum: title}, nil
	}

	return episodeTitles, nil
}

// NormalizeAnimeTitle ensures the title follows MAL’s URL format (e.g., replacing spaces with underscores).
func NormalizeAnimeTitle(animeName string) string {
	normalized := strings.ToLower(animeName)
	normalized = strings.ReplaceAll(normalized, " ", "_")
	return normalized
}
