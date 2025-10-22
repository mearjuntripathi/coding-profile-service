package scraper

import (
    "coding-profile-service/pkg/model"
    "fmt"
    "github.com/PuerkitoBio/goquery"
    "net/http"
    "regexp"
    "strconv"
    "strings"
)

func FetchCodeChefHTML(username string) (model.StatsResponse, error) {
    url := fmt.Sprintf("https://www.codechef.com/users/%s", username)
    res, err := http.Get(url)
    if err != nil {
        return model.StatsResponse{}, err
    }
    defer res.Body.Close()

    if res.StatusCode != 200 {
        return model.StatsResponse{}, fmt.Errorf("failed to fetch CodeChef page: %d", res.StatusCode)
    }

    doc, err := goquery.NewDocumentFromReader(res.Body)
    if err != nil {
        return model.StatsResponse{}, err
    }

    stats := model.StatsResponse{Platform: "codechef", Username: username}

    // Rating
    stats.Rating, _ = strconv.Atoi(doc.Find(".rating-header .rating-number").Text())

    // Max Rating
    doc.Find(".rating-header small").Each(func(i int, s *goquery.Selection) {
        txt := s.Text()
        re := regexp.MustCompile(`\d+`)
        if match := re.FindString(txt); match != "" {
            stats.MaxRating, _ = strconv.Atoi(match)
        }
    })

    // Global and Country Rank
    ranks := doc.Find(".rating-ranks li strong")
    if ranks.Length() >= 2 {
        stats.GlobalRank, _ = strconv.Atoi(strings.TrimSpace(ranks.Eq(0).Text()))
        stats.CountryRank, _ = strconv.Atoi(strings.TrimSpace(ranks.Eq(1).Text()))
    }

    // Contests Participated
    doc.Find(".contest-participated-count b").Each(func(i int, s *goquery.Selection) {
        stats.ContestsParticipated, _ = strconv.Atoi(s.Text())
    })

    // Total Problems Solved
    doc.Find(".rating-data-section.problems-solved h3").Each(func(i int, s *goquery.Selection) {
        re := regexp.MustCompile(`\d+`)
        if match := re.FindString(s.Text()); match != "" {
            stats.TotalSolved, _ = strconv.Atoi(match)
        }
    })

    return stats, nil
}
