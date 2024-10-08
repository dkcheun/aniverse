package extractor

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"aniverse/internal/crawler"
	"aniverse/internal/types"
)

// Constants for encryption and decryption keys, as well as the fixed initialization vector.
const (
	encryptionKey        = "37911490979715163134003223491201"
	decryptionKey        = "54674138327930866480207815084989"
	initializationVector = "3134003223491201"
)

// Common errors used throughout the extractor package.
var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrRequest         = errors.New("request error")
	ErrJSONParse       = errors.New("JSON parsing error")
	ErrScraping        = errors.New("scraping error")
	ErrNoContent       = errors.New("no content found")
	ErrInvalidRegex    = errors.New("invalid regex")
)

// Responsible for handling the decryption of video sources
// from GogoAnime. It holds the keys, initialization vector (IV), and a base crawler
// to fetch the content.
type Gogocdn struct {
	key             []byte
	decryptionKey   []byte
	iv              []byte
	baseCrawler     *crawler.BaseCrawler
	reEncryptedData *regexp.Regexp
}

// Initializes a new instance of Gogocdn, taking in a BaseCrawler.
// It uses predefined encryption and decryption keys, as well as a fixed IV.
func NewGogocdn(c *crawler.BaseCrawler) *Gogocdn {
	baseCrawler := ensureBaseCrawler(c)
	return &Gogocdn{
		key:             []byte(encryptionKey),
		decryptionKey:   []byte(decryptionKey),
		iv:              []byte(initializationVector),
		baseCrawler:     baseCrawler,
		reEncryptedData: regexp.MustCompile(`data-value="(.+?)"`),
	}
}

// The structure for the JSON data returned by GogoAnime.
type gogoCdnData struct {
	Data string `json:"data"`
}

// Holds the sources and backups (Bkp) for video files.
// It is extracted from the decrypted JSON response.
type gogoCdn struct {
	Source []struct {
		File string `json:"file"`
		Type string `json:"type"`
	} `json:"source"`
	Bkp []struct {
		File string `json:"file"`
		Type string `json:"type"`
	} `json:"source_bk"`
	Track interface{} `json:"track"`
}

// Retrieves the streaming sources for a given link. It first
// parses the page, decrypts the content, and returns a structured response.
func (g *Gogocdn) Extract(link string) (*types.Source, error) {
	// Initialize the Source struct with all necessary fields
	sources := &types.Source{
		Sources:       []types.Quality{},
		Subtitles:     []string{},
		Audio:         []string{},
		Headers:       make(map[string]string),
		Intro:         types.EpisodeTiming{Start: 0, End: 0},
		Outro:         types.EpisodeTiming{Start: 0, End: 0},
		Thumbnail:     "",
		ThumbnailType: "",
	}

	// Parse the URL and extract the content ID.
	parsedURL, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("Gogocdn Extract: %w : URL is not valid", ErrInvalidArgument)
	}

	// Extract the 'id' parameter from the query string.
	contentID := parsedURL.Query().Get("id")
	if contentID == "" {
		return nil, fmt.Errorf("Gogocdn Extract: %w : URL does not have 'id' query parameter", ErrInvalidArgument)
	}

	// Parse and retrieve the encrypted parameters from the page.
	encryptedParams, err := g.parsePage(link, contentID)
	if err != nil {
		return nil, fmt.Errorf("Gogocdn Extract: %w", err)
	}

	// Construct the URL to fetch the encrypted content.
	nextHost := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	apiURL := fmt.Sprintf("%s/encrypt-ajax.php?%s", nextHost, encryptedParams)
	headers := map[string]string{"X-Requested-With": "XMLHttpRequest"}

	// Send the request to fetch the encrypted video data.
	response, err := g.baseCrawler.Client.Get(apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("Gogocdn Extract: %w : %s", ErrRequest, err.Error())
	}

	// Unmarshal the JSON response into gogoCdnData structure.
	var gogoCdnResponse gogoCdnData
	err = json.Unmarshal(response, &gogoCdnResponse)
	if err != nil {
		return nil, fmt.Errorf("Gogocdn Extract: %w : %s", ErrJSONParse, err.Error())
	}

	// Decrypt the content.
	decData, err := g.aesDecrypt(gogoCdnResponse.Data, g.decryptionKey, g.iv)
	if err != nil {
		return nil, fmt.Errorf("Gogocdn Extract: %w : Cannot decrypt sources : %s", ErrScraping, err.Error())
	}

	// Unmarshal the decrypted data into a structured gogoCdn object.
	var dataFile gogoCdn
	err = json.Unmarshal(decData, &dataFile)
	if err != nil {
		return nil, fmt.Errorf("Gogocdn Extract: %w : %s", ErrJSONParse, err.Error())
	}

	// Iterate over the primary sources to extract the master m3u8 URL
	for _, s := range dataFile.Source {
		if s.File == "" {
			continue
		}
		// Extract intro/outro from the first M3U8 file
		introEnd, outroStart, err := g.calculateIntroOutro(s.File)
		if err != nil {
			return nil, fmt.Errorf("Gogocdn Extract: failed to extract intro/outro: %w", err)
		}

		// Parse the master m3u8 to extract qualities
		qualities, err := g.parseMasterM3U8(s.File)
		if err != nil {
			return nil, fmt.Errorf("Gogocdn Extract: failed to parse master m3u8: %w", err)
		}

		// Append the parsed qualities to the Source struct
		sources.Sources = append(sources.Sources, qualities...)
		sources.IsM3U8 = true
		sources.Intro = introEnd
		sources.Outro = outroStart
	}

	// Process backup sources and track data (if available).
	for _, s := range dataFile.Bkp {
		if s.File == "" {
			continue
		}
		if strings.Contains(strings.ToLower(s.Type), "subtitle") {
			sources.Subtitles = append(sources.Subtitles, s.File)
		} else if strings.Contains(strings.ToLower(s.Type), "audio") {
			sources.Audio = append(sources.Audio, s.File)
		}
	}

	// Handle track data (e.g., thumbnails).
	switch track := dataFile.Track.(type) {
	case []interface{}:
		// Handle if track is an array
		for _, item := range track {
			if trackMap, ok := item.(map[string]interface{}); ok {
				if kind, ok := trackMap["kind"].(string); ok && strings.ToLower(kind) == "thumbnails" {
					if file, ok := trackMap["file"].(string); ok {
						sources.Thumbnail = file
						sources.ThumbnailType = "Sprite"
					}
				}
			}
		}
	case map[string]interface{}:
		// Handle if track is a map
		if tracks, ok := track["tracks"].([]interface{}); ok {
			for _, item := range tracks {
				if trackMap, ok := item.(map[string]interface{}); ok {
					if kind, ok := trackMap["kind"].(string); ok && strings.ToLower(kind) == "thumbnails" {
						if file, ok := trackMap["file"].(string); ok {
							sources.Thumbnail = file
							sources.ThumbnailType = "Sprite"
						}
					}
				}
			}
		}
	}

	return sources, nil
}

// Fetches and parses the web page content, decrypting the necessary
// data and returning the parsed parameters needed for further extraction.
func (g *Gogocdn) parsePage(link string, contentID string) (string, error) {
	response, err := g.baseCrawler.Client.Get(link, nil)
	if err != nil {
		return "", fmt.Errorf("Gogocdn parsePage: %w", ErrRequest)
	}

	match := g.reEncryptedData.FindSubmatch(response)
	if len(match) < 2 {
		return "", fmt.Errorf("Gogocdn parsePage: %w", ErrInvalidRegex)
	}
	encryptedData := match[1]

	decryptedData, err := g.aesDecrypt(string(encryptedData), g.key, g.iv)
	if err != nil {
		return "", fmt.Errorf("Gogocdn parsePage: %w : decryption error : %s", ErrScraping, err.Error())
	}

	encryptedContentID, err := g.aesEncrypt([]byte(contentID), g.key, g.iv)
	if err != nil {
		return "", fmt.Errorf("Gogocdn parsePage: %w : encryption error : %s", ErrScraping, err.Error())
	}

	component := fmt.Sprintf("id=%s&alias=%s&%s", encryptedContentID, contentID, decryptedData)

	return component, nil
}

// pad applies PKCS#7 padding to the data to make it a multiple of the block size.
func (g *Gogocdn) pad(data []byte) []byte {
	padding := 16 - (len(data) % 16)
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

// unpad removes PKCS#7 padding from the data.
func (g *Gogocdn) unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("unpad: data is empty")
	}

	padding := int(data[len(data)-1])
	if padding == 0 || padding > len(data) {
		return nil, errors.New("unpad: invalid padding")
	}

	// Verify padding bytes
	for i := 0; i < padding; i++ {
		if data[len(data)-1-i] != byte(padding) {
			return nil, errors.New("unpad: invalid padding bytes")
		}
	}

	return data[:len(data)-padding], nil
}

// Performs AES encryption using the provided key and IV.
// It returns the base64-encoded ciphertext.
func (g *Gogocdn) aesEncrypt(data []byte, key []byte, iv []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aesEncrypt: %w", err)
	}

	paddedData := g.pad(data)

	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(paddedData))
	mode.CryptBlocks(ciphertext, paddedData)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Performs AES decryption using the provided key and IV.
// It returns the plaintext data after decryption and unpadding.
func (g *Gogocdn) aesDecrypt(data string, key []byte, iv []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("aesDecrypt: base64 decode failed: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aesDecrypt: %w", err)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("aesDecrypt: ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	unpadded, err := g.unpad(plaintext)
	if err != nil {
		return nil, fmt.Errorf("aesDecrypt: %w", err)
	}

	return unpadded, nil
}

// ensureBaseCrawler ensures that a BaseCrawler instance is available.
func ensureBaseCrawler(c *crawler.BaseCrawler) *crawler.BaseCrawler {
	if c == nil {
		return crawler.NewBaseCrawler()
	}
	return c
}

// calculateIntroOutro extracts the intro and outro segments from the m3u8 playlist
func (g *Gogocdn) calculateIntroOutro(m3u8URL string) (types.EpisodeTiming, types.EpisodeTiming, error) {
	resp, err := http.Get(m3u8URL)
	if err != nil {
		return types.EpisodeTiming{}, types.EpisodeTiming{}, fmt.Errorf("failed to fetch m3u8: %w", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var totalDuration float64
	var introDuration float64 = 90.0 // Example: Assume intro is always 90 seconds
	var outroStart float64
	var outroDuration float64 = 60.0 // Example: Assume outro is 60 seconds

	var currentDuration float64

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#EXTINF:") {
			// Parse segment duration
			durationStr := strings.TrimPrefix(line, "#EXTINF:")
			durationStr = strings.TrimRight(durationStr, ",")
			duration, err := strconv.ParseFloat(durationStr, 64)
			if err != nil {
				return types.EpisodeTiming{}, types.EpisodeTiming{}, fmt.Errorf("failed to parse EXTINF duration: %w", err)
			}
			currentDuration += duration
			totalDuration += duration
		}
	}

	if err := scanner.Err(); err != nil {
		return types.EpisodeTiming{}, types.EpisodeTiming{}, fmt.Errorf("error reading m3u8: %w", err)
	}

	// Set the outro start based on percentage of total duration
	outroStart = totalDuration * 0.85

	introTiming := types.EpisodeTiming{Start: 0, End: introDuration}
	outroTiming := types.EpisodeTiming{Start: outroStart, End: outroStart + outroDuration}

	return introTiming, outroTiming, nil
}

// extractDurationsFromM3U8 sums up the duration of media segments in the m3u8 playlist.
// func (g *Gogocdn) extractDurationsFromM3U8(m3u8URL string) (int, int, error) {
// 	resp, err := http.Get(m3u8URL)
// 	if err != nil {
// 		return 0, 0, fmt.Errorf("failed to fetch m3u8: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	scanner := bufio.NewScanner(resp.Body)
// 	var totalDuration int
// 	var segmentDurations []int

// 	for scanner.Scan() {
// 		line := strings.TrimSpace(scanner.Text())
// 		if strings.HasPrefix(line, "#EXTINF:") {
// 			durationStr := strings.TrimPrefix(line, "#EXTINF:")
// 			durationStr = strings.TrimRight(durationStr, ",")
// 			duration, err := strconv.ParseFloat(durationStr, 64)
// 			if err != nil {
// 				return 0, 0, fmt.Errorf("failed to parse EXTINF duration: %w", err)
// 			}
// 			segmentDurations = append(segmentDurations, int(duration))
// 			totalDuration += int(duration)
// 		}
// 	}

// 	if err := scanner.Err(); err != nil {
// 		return 0, 0, fmt.Errorf("error reading m3u8: %w", err)
// 	}

// 	introEnd := totalDuration / 10       // First 10%
// 	outroStart := totalDuration * 9 / 10 // Start outro at 90%

// 	return introEnd, outroStart, nil
// }

// parseMasterM3U8 fetches and parses the master .m3u8 playlist.
func (g *Gogocdn) parseMasterM3U8(masterURL string) ([]types.Quality, error) {
	// Fetch the master .m3u8 content
	resp, err := http.Get(masterURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch master m3u8: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	var qualities []types.Quality
	var currentQuality types.Quality

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "#EXT-X-STREAM-INF") {
			// Extract attributes
			attributes := parseAttributes(line)
			bandwidth, err := strconv.Atoi(attributes["BANDWIDTH"])
			if err != nil {
				return nil, fmt.Errorf("invalid BANDWIDTH value: %w", err)
			}

			currentQuality = types.Quality{
				Name:       attributes["NAME"],
				Bandwidth:  bandwidth,
				Resolution: attributes["RESOLUTION"],
			}
		} else if strings.HasSuffix(line, ".m3u8") && currentQuality.Name != "" {
			// Assign the URL to the current quality
			currentQuality.SubURL = resolveURL(masterURL, line)
			qualities = append(qualities, currentQuality)
			// Reset currentQuality for next entry
			currentQuality = types.Quality{}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading master m3u8: %w", err)
	}

	if len(qualities) == 0 {
		return nil, errors.New("no qualities found in master m3u8")
	}

	return qualities, nil
}

// parseAttributes parses the attributes from a #EXT-X-STREAM-INF line.
func parseAttributes(line string) map[string]string {
	attributes := make(map[string]string)
	// Remove the prefix
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return attributes
	}

	attrs := strings.Split(parts[1], ",")
	for _, attr := range attrs {
		kv := strings.SplitN(attr, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.Trim(strings.TrimSpace(kv[1]), `"`)
			attributes[key] = value
		}
	}
	return attributes
}

// resolveURL resolves relative URLs based on the master playlist URL.
func resolveURL(masterURL, relativeURL string) string {
	u, err := url.Parse(masterURL)
	if err != nil {
		return relativeURL // Fallback to the relative URL if parsing fails
	}
	base := u.Scheme + "://" + u.Host + path.Dir(u.Path)
	resolvedURL, err := url.Parse(relativeURL)
	if err != nil {
		return relativeURL
	}
	return base + "/" + resolvedURL.Path
}
