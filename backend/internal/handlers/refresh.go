package handlers

import (
	"hash/fnv"

	"github.com/recolab/backend/internal/gorse"
)

func candidateLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	return limit * 4
}

func selectRefreshWindow(scores []gorse.Score, refresh string, limit int) []gorse.Score {
	if limit <= 0 || len(scores) <= limit {
		return scores
	}
	if refresh == "" {
		return scores[:limit]
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(refresh))
	start := int(h.Sum32()) % len(scores)

	out := make([]gorse.Score, 0, limit)
	for i := 0; len(out) < limit && i < len(scores); i++ {
		out = append(out, scores[(start+i)%len(scores)])
	}
	return out
}
