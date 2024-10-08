package util

import (
	"aniverse/internal/types"
	"strings"
)

func JaroWinkler(s1, s2 string) float64 {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	m := len(s1)
	n := len(s2)

	if m == 0 && n == 0 {
		return 1.0
	}
	if m == 0 || n == 0 {
		return 0.0
	}

	matchDistance := max(m, n)/2 - 1

	s1Matches := make([]bool, m)
	s2Matches := make([]bool, n)

	matches := 0
	transpositions := 0

	// Find matches
	for i := 0; i < m; i++ {
		start := max(0, i-matchDistance)
		end := min(n-1, i+matchDistance)

		for j := start; j <= end; j++ {
			if s2Matches[j] {
				continue
			}
			if s1[i] != s2[j] {
				continue
			}
			s1Matches[i] = true
			s2Matches[j] = true
			matches++
			break
		}
	}

	if matches == 0 {
		return 0.0
	}

	// Count transpositions
	k := 0
	for i := 0; i < m; i++ {
		if !s1Matches[i] {
			continue
		}
		for !s2Matches[k] {
			k++
		}
		if s1[i] != s2[k] {
			transpositions++
		}
		k++
	}

	transpositions /= 2

	matchScore := float64(matches) / float64(m)
	transpositionScore := float64(matches-transpositions) / float64(matches)
	similarity := (matchScore + float64(matches)/float64(n) + transpositionScore) / 3

	// Apply Winkler modification
	prefixLength := 0
	for i := 0; i < min(4, min(m, n)); i++ {
		if s1[i] == s2[i] {
			prefixLength++
		} else {
			break
		}
	}

	return similarity + float64(prefixLength)*0.1*(1-similarity)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Sanitize removes unnecessary parts from the title
func Sanitize(title string) string {
	title = strings.ToLower(title)
	title = strings.ReplaceAll(title, "[^\\p{L}\\p{N}\\s]", "")
	wordsToRemove := []string{"season", "cour", "part"}
	words := strings.Fields(title)
	var sanitizedWords []string
	for _, word := range words {
		remove := false
		for _, removeWord := range wordsToRemove {
			if word == removeWord {
				remove = true
				break
			}
		}
		if !remove {
			sanitizedWords = append(sanitizedWords, word)
		}
	}
	return strings.Join(sanitizedWords, " ")
}

// FindBestMatch finds the best matching title from a list of targets
func FindBestMatch(mainString string, targets []string) string {
	if len(targets) == 0 {
		return ""
	}

	bestMatch := targets[0]
	highestScore := JaroWinkler(mainString, bestMatch)

	for _, target := range targets[1:] {
		score := JaroWinkler(mainString, target)
		if score > highestScore {
			highestScore = score
			bestMatch = target
		}
	}

	return bestMatch
}

// FindOriginalTitle finds the best matching title among the given titles
func FindOriginalTitle(title types.Title, titles []string) string {
	romajiBestMatch := FindBestMatch(title.Romaji, titles)
	englishBestMatch := FindBestMatch(title.English, titles)
	nativeBestMatch := FindBestMatch(title.Native, titles)

	romajiScore := JaroWinkler(title.Romaji, romajiBestMatch)
	englishScore := JaroWinkler(title.English, englishBestMatch)
	nativeScore := JaroWinkler(title.Native, nativeBestMatch)

	if romajiScore >= englishScore && romajiScore >= nativeScore {
		return romajiBestMatch
	} else if englishScore >= romajiScore && englishScore >= nativeScore {
		return englishBestMatch
	} else {
		return nativeBestMatch
	}
}
