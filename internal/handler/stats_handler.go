package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"coding-profile-service/internal/scraper"
	"coding-profile-service/pkg/model"
)

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	// Read query params
	platform := strings.ToLower(r.URL.Query().Get("platform"))
	username := strings.TrimSpace(r.URL.Query().Get("username"))

	// Map of platform -> username for multi-platform mode
	params := map[string]string{
		"leetcode":   r.URL.Query().Get("leetcode"),
		"gfg":        r.URL.Query().Get("gfg"),
		"codechef":   r.URL.Query().Get("codechef"),
		"hackerrank": r.URL.Query().Get("hackerrank"),
	}

	var results []model.StatsResponse

	// --- Multi-platform Mode ---
	for key, val := range params {
		if val != "" {
			resp, err := fetchPlatformStats(key, val)
			if err != nil {
				resp.Error = err.Error()
			}
			results = append(results, resp)
		}
	}

	// --- Single-platform Mode ---
	if len(results) == 0 && platform != "" && username != "" {
		resp, err := fetchPlatformStats(platform, username)
		if err != nil {
			resp.Error = err.Error()
		}
		results = append(results, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"profiles": results,
	})
}

func fetchPlatformStats(platform, username string) (model.StatsResponse, error) {
	switch platform {
	case "leetcode":
		return scraper.FetchLeetCode(username)
	case "gfg":
		return scraper.FetchGFG(username)
	case "codechef":
		return scraper.FetchCodeChef(username)
	case "hackerrank":
		return scraper.FetchHackerRank(username)
	default:
		return model.StatsResponse{Platform: platform, Username: username, Error: "unsupported platform"}, nil
	}
}
