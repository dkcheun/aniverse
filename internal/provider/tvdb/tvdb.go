package tvdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type AuthResponse struct {
	Token string `json:"token"`
}

type Series struct {
	ID         int    `json:"id"`
	SeriesName string `json:"seriesName"`
	// Add other fields as needed
}

type SeriesData struct {
	Data []Series `json:"data"`
}

type Client struct {
	APIKey string
	Token  string
	Client *http.Client
}

type Episode struct {
	ID          int    `json:"id"`
	EpisodeName string `json:"episodeName"`
	Overview    string `json:"overview"`
}

type EpisodeData struct {
	Data []Episode `json:"data"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
		Client: &http.Client{},
	}
}

func (c *Client) Authenticate() error {
	url := "https://api.thetvdb.com/login"
	payload := map[string]string{
		"apikey": os.Getenv("TVDB_CLIENT"),
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return err
	}

	c.Token = authResp.Token
	return nil
}

func (c *Client) SearchSeriesByName(name string) ([]Series, error) {
	searchURL := fmt.Sprintf("https://api.thetvdb.com/search/series?name=%s", url.QueryEscape(name))
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data SeriesData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Data, nil
}

func (c *Client) GetEpisodes(seriesID int) ([]Episode, error) {
	episodesURL := fmt.Sprintf("https://api.thetvdb.com/series/%d/episodes", seriesID)
	req, err := http.NewRequest("GET", episodesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var epData EpisodeData
	if err := json.NewDecoder(resp.Body).Decode(&epData); err != nil {
		return nil, err
	}

	return epData.Data, nil
}
