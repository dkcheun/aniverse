package crawler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gocolly/colly/v2"
)

var (
	DefaultBaseCrawler = NewBaseCrawler()
)

type BaseCrawler struct {
	Client    *HttpClient
	Collector *colly.Collector
}

func NewBaseCrawler() *BaseCrawler {
	return &BaseCrawler{
		Client:    &HttpClient{},
		Collector: colly.NewCollector(),
	}
}

type HttpClient struct{}

func (c *HttpClient) Get(url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (c *BaseCrawler) OnHTML(selector string, callback func(*colly.HTMLElement)) {
	c.Collector.OnHTML(selector, callback)
}

func (c *BaseCrawler) Visit(url string) error {
	return c.Collector.Visit(url)
}
