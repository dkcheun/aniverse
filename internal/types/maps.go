package types

import (
	"encoding/json"
	"fmt"
)

type AnimeMapping struct {
	AniListID     int    `json:"anilist_id"`
	MALID         int    `json:"mal_id"`
	AnimePlanetID string `json:"anime-planet_id"`
	KitsuID       int    `json:"kitsu_id"`
}

// UnmarshalJSON handles both string and number for AnimePlanetID.
func (a *AnimeMapping) UnmarshalJSON(data []byte) error {
	type Alias AnimeMapping
	aux := &struct {
		AnimePlanetID interface{} `json:"anime-planet_id"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert AnimePlanetID to string if necessary
	switch id := aux.AnimePlanetID.(type) {
	case string:
		a.AnimePlanetID = id
	case float64:
		a.AnimePlanetID = fmt.Sprintf("%.0f", id) // Convert number to string
	default:
		return fmt.Errorf("unexpected type for anime-planet_id: %T", id)
	}

	return nil
}
