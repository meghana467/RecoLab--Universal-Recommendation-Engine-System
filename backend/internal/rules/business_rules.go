package rules

import (
	"fmt"
	"strings"

	"github.com/recolab/backend/internal/gorse"
)

// FilterInactiveItems removes items not in the active set.
func FilterInactiveItems(items []gorse.Score, activeIds map[string]bool) []gorse.Score {
	if len(activeIds) == 0 {
		return items
	}
	result := make([]gorse.Score, 0, len(items))
	for _, item := range items {
		if activeIds[item.Id] {
			result = append(result, item)
		}
	}
	return result
}

// DeduplicateItems removes duplicate item IDs, keeping the first occurrence.
func DeduplicateItems(items []gorse.Score) []gorse.Score {
	seen := make(map[string]bool)
	result := make([]gorse.Score, 0, len(items))
	for _, item := range items {
		if !seen[item.Id] {
			seen[item.Id] = true
			result = append(result, item)
		}
	}
	return result
}

// AddExplanation generates a human-readable reason for a recommendation.
func AddExplanation(item gorse.Score, userLabels []string, itemLabels []string) string {
	if len(userLabels) == 0 && len(itemLabels) == 0 {
		return fmt.Sprintf("Recommended based on overall popularity (score: %.2f)", item.Score)
	}

	matching := intersect(userLabels, itemLabels)
	if len(matching) > 0 {
		return fmt.Sprintf("Matches your interests: %s", strings.Join(capitalize(matching), ", "))
	}

	if len(itemLabels) > 0 {
		sample := itemLabels
		if len(sample) > 3 {
			sample = sample[:3]
		}
		return fmt.Sprintf("Popular in: %s", strings.Join(capitalize(sample), ", "))
	}

	return "Recommended for you based on your activity"
}

func intersect(a, b []string) []string {
	set := make(map[string]bool)
	for _, v := range a {
		set[strings.ToLower(v)] = true
	}
	var result []string
	for _, v := range b {
		if set[strings.ToLower(v)] {
			result = append(result, v)
		}
	}
	return result
}

func capitalize(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		if len(s) > 0 {
			out[i] = strings.ToUpper(s[:1]) + s[1:]
		} else {
			out[i] = s
		}
	}
	return out
}
