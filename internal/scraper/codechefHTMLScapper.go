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

    // Rating — from #rating-block-all
    ratingBlock := doc.Find("#rating-block-all")

    // Rating number
    ratingText := strings.TrimSpace(ratingBlock.Find(".rating-number").Text())
    stats.Rating, _ = strconv.Atoi(ratingText)

    // Stars — count ★ symbols
    starText := ratingBlock.Find(".rating-star").Text()
    stats.Stars = strings.Count(starText, "★")

    // Max Rating — from <small>(Highest Rating 1734)</small>
    ratingBlock.Find("small").Each(func(i int, s *goquery.Selection) {
        txt := s.Text()
        if strings.Contains(txt, "Highest Rating") {
            re := regexp.MustCompile(`\d+`)
            if match := re.FindString(txt); match != "" {
                stats.MaxRating, _ = strconv.Atoi(match)
            }
        }
    })

    // Global and Country Rank
    ranks := ratingBlock.Find(".rating-ranks li strong")
    if ranks.Length() >= 2 {
        stats.GlobalRank, _ = strconv.Atoi(strings.TrimSpace(ranks.Eq(0).Text()))
        stats.CountryRank, _ = strconv.Atoi(strings.TrimSpace(ranks.Eq(1).Text()))
    }

    // Contests Participated
    doc.Find(".contest-participated-count b").Each(func(i int, s *goquery.Selection) {
        stats.ContestsParticipated, _ = strconv.Atoi(strings.TrimSpace(s.Text()))
    })

    // Total Problems Solved
    doc.Find(".rating-data-section.problems-solved h3").Each(func(i int, s *goquery.Selection) {
        re := regexp.MustCompile(`\d+`)
        if match := re.FindString(s.Text()); match != "" {
            n, _ := strconv.Atoi(match)
            if n > stats.TotalSolved {
                stats.TotalSolved = n // take highest number found
            }
        }
    })

    return stats, nil
}
