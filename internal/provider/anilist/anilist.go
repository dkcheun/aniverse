package anilist

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"aniverse/internal/types"
)

type AniListBase struct {
	BaseURL string
	query   string
}

func NewAniListBase() *AniListBase {
	return &AniListBase{
		BaseURL: "https://graphql.anilist.co",
		query: `
id
idMal
title {
  romaji
  english
  native
}
coverImage {
  extraLarge
  color
}
bannerImage
description
season
seasonYear
type
format
status(version: 2)
episodes
duration
genres
synonyms
isAdult
meanScore
popularity
countryOfOrigin
tags {
  name
}
characters {
  edges {
    node {
      name {
        full
      }
      image {
        large
      }
    }
    voiceActors {
      name {
        full
      }
      image {
        large
      }
    }
  }
}
relations {
  edges {
    relationType(version: 2)
    node {
      id
      title {
        romaji
        english
        native
      }
      format
      type
    }
  }
}
`,
	}
}

func (a *AniListBase) ID() string {
	return "anilist"
}

func (a *AniListBase) URL() string {
	return "https://anilist.co"
}

func (a *AniListBase) Formats() []types.Format {
	return []types.Format{
		types.FormatMovie,
		types.FormatONA,
		types.FormatOVA,
		types.FormatSpecial,
		types.FormatTV,
		types.FormatTVShort,
	}
}

func (a *AniListBase) NeedsProxy() bool {
	return true
}

func (a *AniListBase) UseGoogleTranslate() bool {
	return false
}

func (anilist *AniListBase) Search(query string, mediaType types.MediaType, formats []types.Format, page int, perPage int) ([]types.AnimeInfo, error) {
	graphqlQuery := `
query ($page: Int, $perPage: Int, $search: String, $type: MediaType, $format: [MediaFormat]) {
  Page(page: $page, perPage: $perPage) {
    media(type: $type, format_in: $format, search: $search) {
` + anilist.query + `
    }
  }
}
`

	variables := map[string]interface{}{
		"search":  query,
		"type":    mediaType,
		"format":  formats,
		"page":    page,
		"perPage": perPage,
	}

	payload := map[string]interface{}{
		"query":     graphqlQuery,
		"variables": variables,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", anilist.BaseURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://anilist.co")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Data struct {
			Page struct {
				Media []types.Media `json:"media"`
			} `json:"Page"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	mediaList := response.Data.Page.Media
	if mediaList == nil {
		return nil, errors.New("no media found")
	}

	var results []types.AnimeInfo
	for _, media := range mediaList {
		if media.IsAdult {
			continue
		}
		animeInfo := anilist.mapMediaToAnimeInfo(media)
		results = append(results, animeInfo)
	}

	return results, nil
}

func (a *AniListBase) GetMedia(id string) (*types.AnimeInfo, error) {
	graphqlQuery := `
query ($id: Int) {
  Media(id: $id) {
` + a.query + `
  }
}
`

	variables := map[string]interface{}{
		"id": id,
	}

	payload := map[string]interface{}{
		"query":     graphqlQuery,
		"variables": variables,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", a.BaseURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://anilist.co")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Data struct {
			Media types.Media `json:"Media"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	media := response.Data.Media
	if media.IsAdult {
		return nil, errors.New("media is adult content")
	}

	animeInfo := a.mapMediaToAnimeInfo(media)
	return &animeInfo, nil
}

func (a *AniListBase) mapMediaToAnimeInfo(media types.Media) types.AnimeInfo {
	title := types.Title{
		English: media.Title.English,
		Romaji:  media.Title.Romaji,
		Native:  media.Title.Native,
	}

	var rating *float64
	if media.MeanScore != 0 {
		val := float64(media.MeanScore) / 10
		rating = &val
	}

	artworks := []types.Artwork{
		{
			Type:       "poster",
			Img:        media.CoverImage.ExtraLarge,
			ProviderID: a.ID(),
		},
	}

	if media.BannerImage != "" {
		artworks = append(artworks, types.Artwork{
			Type:       "banner",
			Img:        media.BannerImage,
			ProviderID: a.ID(),
		})
	}

	characters := a.mapCharacters(media.Characters)
	relations := a.mapRelations(media.Relations)

	description := stripHTMLTags(media.Description)

	color := media.CoverImage.Color
	animeInfo := types.AnimeInfo{
		ID:              fmt.Sprintf("%d", media.ID),
		IDMal:           fmt.Sprintf("%d", media.IDMal),
		Title:           title,
		Description:     &description,
		CoverImage:      &media.CoverImage,
		BannerImage:     &media.BannerImage,
		Popularity:      media.Popularity,
		Synonyms:        media.Synonyms,
		TotalEpisodes:   media.Episodes,
		Status:          types.MediaStatus(media.Status),
		Season:          (*types.Season)(&media.Season),
		Genres:          media.Genres,
		Rating:          rating,
		Format:          types.Format(media.Format),
		Year:            &media.SeasonYear,
		Type:            types.MediaType(media.Type),
		CountryOfOrigin: &media.CountryOfOrigin,
		Tags:            a.extractTagNames(media.Tags),
		Artwork:         artworks,
		Characters:      characters,
		Relations:       relations,
		CurrentEpisode:  media.Episodes,
		Duration:        &media.Duration,
		Color:           &color,
	}

	return animeInfo
}

func (a *AniListBase) mapCharacters(characters types.Characters) []types.Character {
	var result []types.Character
	for _, edge := range characters.Edges {
		character := types.Character{
			Name:  edge.Node.Name.Full,
			Image: edge.Node.Image.Large,
		}
		if len(edge.VoiceActors) > 0 {
			va := edge.VoiceActors[0]
			character.VoiceActor = &types.VoiceActor{
				Name:  va.Name.Full,
				Image: va.Image.Large,
			}
		}
		result = append(result, character)
	}
	return result
}

func (a *AniListBase) mapRelations(relations types.Relations) []types.Relation {
	var result []types.Relation
	for _, edge := range relations.Edges {
		relation := types.Relation{
			ID:           fmt.Sprintf("%d", edge.Node.ID),
			Format:       types.Format(edge.Node.Format),
			RelationType: edge.RelationType,
			Title: types.Title{
				English: edge.Node.Title.English,
				Romaji:  edge.Node.Title.Romaji,
				Native:  edge.Node.Title.Native,
			},
			Type: types.MediaType(edge.Node.Type),
		}
		result = append(result, relation)
	}
	return result
}

func (a *AniListBase) extractTagNames(tags []types.Tag) []string {
	var names []string
	for _, tag := range tags {
		names = append(names, tag.Name)
	}
	return names
}

func stripHTMLTags(s string) string {
	// Optionally implement HTML tag stripping if needed
	return strings.TrimSpace(s)
}
