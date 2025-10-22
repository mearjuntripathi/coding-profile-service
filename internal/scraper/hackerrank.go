package scraper

import (
	"coding-profile-service/pkg/model"
	"fmt"
)

// FetchHackerRank is the function your API handler should call.
// It delegates to FetchHackerRankHTML (and later you can add REST or chromedp fallback).
func FetchHackerRank(username string) (model.StatsResponse, error) {
	stats, err := FetchHackerRankHTML(username)
	if err != nil {
		return model.StatsResponse{
			Platform: "hackerrank",
			Username: username,
			Error:    fmt.Sprintf("could not fetch HackerRank data for user: %s (%v)", username, err),
		}, err
	}
	return stats, nil
}
