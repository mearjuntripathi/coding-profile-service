package scraper

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GFGProfile holds the scraped data temporarily
type GFGProfile struct {
	TotalSolved          int
	Streak               int
	EasySolved           int
	MediumSolved         int
	HardSolved           int
	ContestsParticipated int
	MaxRating            int
	GlobalRank           int
	CountryRank          int
}

// FetchGFGHTML scrapes the GFG user profile HTML
func FetchGFGHTML(username string) (*GFGProfile, error) {
	url := fmt.Sprintf("https://auth.geeksforgeeks.org/user/%s", username)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	profile := &GFGProfile{}

	// Total Problems Solved
	doc.Find(".scoreCard_head_left--text__KZ2S1").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		value := s.Next().Text()
		switch text {
		case "Problem Solved":
			profile.TotalSolved, _ = strconv.Atoi(strings.TrimSpace(value))
		case "Coding Score":
			profile.MaxRating, _ = strconv.Atoi(strings.TrimSpace(value))
		case "Article Published":
			// optional, ignore
		}
	})

	// Streak
	doc.Find(".circularProgressBar_head_mid_streakCnt__MFOF1").Each(func(i int, s *goquery.Selection) {
		streakText := strings.Split(s.Text(), "/")[0]
		profile.Streak, _ = strconv.Atoi(strings.TrimSpace(streakText))
	})

	// Problem difficulty breakdown
	doc.Find(".problemNavbar_head_nav__a4K6P").Each(func(i int, s *goquery.Selection) {
		text := s.Find(".problemNavbar_head_nav--text__UaGCx").Text()
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		parts := strings.Split(text, "(")
		if len(parts) != 2 {
			return
		}
		count, _ := strconv.Atoi(strings.TrimSuffix(parts[1], ")"))
		category := strings.ToUpper(strings.TrimSpace(parts[0]))
		switch category {
		case "SCHOOL", "BASIC", "EASY":
			profile.EasySolved += count
		case "MEDIUM":
			profile.MediumSolved = count
		case "HARD":
			profile.HardSolved = count
		}
	})

	// Contests Participated
	doc.Find(".scoreCards_head__G_uNQ .scoreCard_head_left--text__KZ2S1").Each(func(i int, s *goquery.Selection) {
		if s.Text() == "Contest Rating" {
			// optional: you can parse rating if needed
		}
	})
	doc.Find(".scoreCards_head__G_uNQ .scoreCard_head_left--score__oSi_x").Each(func(i int, s *goquery.Selection) {
		// Hack: total contests count appears in "Problem Solved" card? Otherwise may need another selector
	})

	return profile, nil
}