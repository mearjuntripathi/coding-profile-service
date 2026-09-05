// internal/handler/stats_handler.go
package handler

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "sync"
    "time"

    "coding-profile-service/internal/cache"
    "coding-profile-service/internal/scraper"
    "coding-profile-service/pkg/model"
)


// TTLs per platform — stats don't change that often
// stats_handler.go
var platformTTL = map[string]time.Duration{
    "leetcode":   30 * time.Minute,
    "gfg":        2 * time.Hour,
    "codechef":   2 * time.Hour,
    "hackerrank": 6 * time.Hour,
    "codeforces": 1 * time.Hour,
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"leetcode":   r.URL.Query().Get("leetcode"),
		"gfg":        r.URL.Query().Get("gfg"),
		"codechef":   r.URL.Query().Get("codechef"),
		"hackerrank": r.URL.Query().Get("hackerrank"),
		"codeforces": r.URL.Query().Get("codeforces"),
	}

	// Single platform mode fallback
	platform := strings.ToLower(r.URL.Query().Get("platform"))
	username := strings.TrimSpace(r.URL.Query().Get("username"))
	if platform != "" && username != "" {
		params[platform] = username
	}

	// Collect only requested platforms
	type job struct {
		platform string
		username string
	}
	var jobs []job
	for k, v := range params {
		if v != "" {
			jobs = append(jobs, job{k, v})
		}
	}

	// ✅ Run all platform fetches in parallel
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []model.StatsResponse
	)

	for _, j := range jobs {
		wg.Add(1)
		go func(p, u string) {
			defer wg.Done()
			resp := fetchWithCache(p, u)
			mu.Lock()
			results = append(results, resp)
			mu.Unlock()
		}(j.platform, j.username)
	}

	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"profiles": results,
	})
}

func fetchWithCache(platform, username string) model.StatsResponse {
	cacheKey := fmt.Sprintf("%s:%s", platform, username) // e.g. "leetcode:mearjuntripathi"

	// 1. Check Redis first
	if cached, ok := cache.GetCache(cacheKey); ok {
		cached.Cached = true // optional flag so you know it came from cache
		return cached
	}

	// 2. Cache miss — hit the scraper
	resp, err := fetchPlatformStats(platform, username)
	if err != nil {
		resp.Error = err.Error()
		return resp // don't cache errors
	}

	// 3. Store in Redis with platform-specific TTL
	ttl := platformTTL[platform]
	if ttl == 0 {
		ttl = 30 * time.Minute // default
	}
	cache.SetCache(cacheKey, resp, ttl)

	return resp
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
	case "codeforces":
		return scraper.FetchCodeforces(username)
	default:
		return model.StatsResponse{Platform: platform, Username: username, Error: "unsupported platform"}, nil
	}
}