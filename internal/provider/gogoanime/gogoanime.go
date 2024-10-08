package gogoanime

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"aniverse/internal/extractor"
	"aniverse/internal/types"

	"github.com/PuerkitoBio/goquery"
)

type GogoAnime struct {
	baseURL string
	ajaxURL string
	gogoCDN *extractor.Gogocdn
}

func NewGogoAnime() *GogoAnime {
	return &GogoAnime{
		baseURL: "https://gogoanime3.co", // Ensure this is the correct base URL
		ajaxURL: "https://ajax.gogocdn.net",
		gogoCDN: extractor.NewGogocdn(nil),
	}
}

func (g *GogoAnime) ID() string {
	return "gogoanime"
}

func (g *GogoAnime) URL() string {
	return g.baseURL
}

func (g *GogoAnime) Formats() []types.Format {
	return []types.Format{
		types.FormatMovie,
		types.FormatONA,
		types.FormatOVA,
		types.FormatSpecial,
		types.FormatTV,
		types.FormatTVShort,
	}
}

func (g *GogoAnime) Search(query string) ([]types.AnimeInfo, error) {
	results := []types.AnimeInfo{}
	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("%s/search.html?keyword=%s", g.baseURL, encodedQuery)
	log.Printf("Searching GogoAnime with URL: %s", searchURL) // Added logging for debugging

	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	doc.Find("ul.items > li").Each(func(i int, s *goquery.Selection) {
		titleText := s.Find("p.name a").Text()
		idHref, exists := s.Find("div.img a").Attr("href")
		if !exists {
			return // Skip if href doesn't exist
		}
		id := strings.TrimPrefix(idHref, "/category/")

		releasedText := s.Find("p.released").Text()
		year := extractYear(releasedText)

		imgSrc, exists := s.Find("div.img a img").Attr("src")
		if !exists {
			imgSrc = ""
		}

		// Create AnimeInfo with available data
		animeInfo := types.AnimeInfo{
			ID:    id,
			Title: types.Title{English: titleText, Romaji: titleText, Native: titleText},
			CoverImage: &types.Image{
				ExtraLarge: imgSrc,
				Large:      imgSrc,
				Color:      "",
			},
			Year: func() *int {
				if year > 0 {
					return &year
				} else {
					return nil
				}
			}(),
			Format: types.FormatTV,
			Type:   types.TypeAnime,
		}

		results = append(results, animeInfo)
	})

	return results, nil
}

func (g *GogoAnime) FetchEpisodes(id string) ([]types.Episode, error) {
	log.Printf("Fetching episodes from ID: %s", id)

	// Ensure ID starts with a "/"
	if !strings.HasPrefix(id, "/") {
		id = "/category/" + id
	}

	resp, err := http.Get(g.baseURL + id)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	epStart, _ := doc.Find("#episode_page li").First().Find("a").Attr("ep_start")
	epEnd, _ := doc.Find("#episode_page li").Last().Find("a").Attr("ep_end")
	movieID, _ := doc.Find("#movie_id").Attr("value")
	alias, _ := doc.Find("#alias_anime").Attr("value")

	ajaxURL := fmt.Sprintf("%s/ajax/load-list-episode?ep_start=%s&ep_end=%s&id=%s&default_ep=0&alias=%s", g.ajaxURL, epStart, epEnd, movieID, alias)
	log.Printf("Constructed AJAX URL: %s", ajaxURL)

	ajaxResp, err := http.Get(ajaxURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make AJAX request: %w", err)
	}
	defer ajaxResp.Body.Close()

	if ajaxResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 AJAX response code: %d", ajaxResp.StatusCode)
	}

	ajaxDoc, err := goquery.NewDocumentFromReader(ajaxResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AJAX response: %w", err)
	}

	var episodes []types.Episode
	ajaxDoc.Find("#episode_related li").Each(func(i int, s *goquery.Selection) {
		epID, exists := s.Find("a").Attr("href")
		if !exists {
			return
		}

		epID = strings.TrimSpace(epID)

		numberText := strings.TrimSpace(s.Find("div.name").Text())
		numberText = strings.TrimPrefix(numberText, "EP ")
		number, err := strconv.Atoi(numberText)
		if err != nil {
			number = 0
		}

		episode := types.Episode{
			ID:       epID,
			Number:   number,
			HasDub:   strings.Contains(id, "-dub"),
			IsFiller: false,
			// Img, Description, Rating can be populated if available
		}
		episodes = append(episodes, episode)
	})

	// Reverse the episodes to match the original order
	for i, j := 0, len(episodes)-1; i < j; i, j = i+1, j-1 {
		episodes[i], episodes[j] = episodes[j], episodes[i]
	}

	return episodes, nil
}

func (g *GogoAnime) GetSource(episodeURL string) (string, error) {
	resp, err := http.Get(episodeURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch episode page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse episode page: %w", err)
	}

	// Example: Find the iframe that contains the streaming link
	iframeSrc, exists := doc.Find("iframe").Attr("src")
	if !exists {
		return "", fmt.Errorf("no iframe src found in episode page")
	}

	return iframeSrc, nil
}

// extractYear extracts the year from a string like "Released: 2021".
func extractYear(text string) int {
	parts := strings.Fields(text)
	if len(parts) >= 2 {
		year, err := strconv.Atoi(parts[1])
		if err == nil {
			return year
		}
	}
	return 0
}
