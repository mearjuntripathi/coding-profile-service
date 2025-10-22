package cache

import (
	"time"
	"sync"
	"coding-profile-service/pkg/model"
)

var (
	cacheMu sync.RWMutex
	cache   = map[string]model.StatsResponse{}
	ttl     = 2 * time.Minute // cache TTL, tune as needed
)

func SetCache(key string, resp model.StatsResponse, ttl time.Duration) {
	cacheMu.Lock()
	cache[key] = resp
	cacheMu.Unlock()
}

func GetCache(key string) (model.StatsResponse, bool) {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	if resp, exists := cache[key]; exists {
		return resp, true
	}
	return model.StatsResponse{}, false
}
