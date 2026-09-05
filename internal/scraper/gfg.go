package scraper

import (
	"bytes"
	"coding-profile-service/pkg/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ─── Structs ────────────────────────────────────────────────────────────────

type GFGProfile struct {
	TotalSolved          int
	Streak               int
	EasySolved           int
	MediumSolved         int
	HardSolved           int
	ContestsParticipated int
	MaxRating            int
	CodingScore          int
	GlobalRank           int
	CountryRank          int
}

type gfgUserData struct {
	Score                      int    `json:"score"`
	MonthlyScore               int    `json:"monthly_score"`
	TotalProblemsSolved        int    `json:"total_problems_solved"`
	InstituteRank              string `json:"institute_rank"`
	PodSolvedLongestStreak     int    `json:"pod_solved_longest_streak"`
	PodSolvedCurrentStreak     int    `json:"pod_solved_current_streak"`
	PodCorrectSubmissionsCount int    `json:"pod_correct_submissions_count"`
}

type gfgAPIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Hard   map[string]gfgProblem `json:"Hard"`
		Medium map[string]gfgProblem `json:"Medium"`
		Easy   map[string]gfgProblem `json:"Easy"`
	} `json:"result"`
}

type gfgProblem struct {
	Slug        string `json:"slug"`
	ProblemName string `json:"pname"`
	Lang        string `json:"lang"`
	SubTime     string `json:"user_subtime"`
}

// ─── Main Entry Point ────────────────────────────────────────────────────────

func FetchGFG(username string) (model.StatsResponse, error) {
	type apiResult struct {
		easy, medium, hard int
		err                error
	}
	type profileResult struct {
		profile *GFGProfile
		err     error
	}

	apiCh     := make(chan apiResult, 1)
	profileCh := make(chan profileResult, 1)

	// Goroutine 1 — API for solved counts
	go func() {
		easy, medium, hard, err := fetchGFGSolvedCounts(username)
		apiCh <- apiResult{easy, medium, hard, err}
	}()

	// Goroutine 2 — HTML scraper for streak, rating, rank
	go func() {
		profile, err := FetchGFGProfile(username)
		profileCh <- profileResult{profile, err}
	}()

	api     := <-apiCh
	profile := <-profileCh

	resp := model.StatsResponse{
		Platform: "gfg",
		Username: username,
	}

	// Fill solved counts from API
	if api.err == nil {
		resp.EasySolved   = api.easy
		resp.MediumSolved = api.medium
		resp.HardSolved   = api.hard
		resp.TotalSolved  = api.easy + api.medium + api.hard
	}

	// Fill other fields from HTML scraper
	if profile.err == nil {
		resp.Streak               = profile.profile.Streak
		resp.MaxRating            = profile.profile.MaxRating
		resp.CodingScore          = profile.profile.CodingScore
		resp.GlobalRank           = profile.profile.GlobalRank
		resp.ContestsParticipated = profile.profile.ContestsParticipated
	}

	if api.err != nil && profile.err != nil {
		return resp, fmt.Errorf("could not fetch GFG data for user: %s", username)
	}

	return resp, nil
}

// ─── API: Fetch Solved Counts ─────────────────────────────────────────────────

func fetchGFGSolvedCounts(username string) (easy, medium, hard int, err error) {
	payload := map[string]string{
		"handle":      username,
		"requestType": "",
		"year":        "",
		"month":       "",
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST",
		"https://practiceapi.geeksforgeeks.org/api/v1/user/problems/submissions/",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return 0, 0, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Origin", "https://www.geeksforgeeks.org")
	req.Header.Set("Referer", "https://www.geeksforgeeks.org/")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, 0, err
	}
	defer resp.Body.Close()

	var apiResp gfgAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return 0, 0, 0, err
	}

	if apiResp.Status != "success" {
		return 0, 0, 0, fmt.Errorf("GFG API returned non-success for user: %s", username)
	}

	return len(apiResp.Result.Easy),
		len(apiResp.Result.Medium),
		len(apiResp.Result.Hard), nil
}

// ─── HTML Scraper: Fetch Profile Fields ──────────────────────────────────────

func FetchGFGProfile(username string) (*GFGProfile, error) {
	url := fmt.Sprintf("https://www.geeksforgeeks.org/user/%s/", username)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyStr := string(bodyBytes)

	profile := &GFGProfile{}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyStr))
	if err != nil {
		return nil, err
	}

	doc.Find("script").Each(func(_ int, s *goquery.Selection) {
		if parseNextJSPayload(s.Text(), profile) {
			return
		}
	})

	parseDifficultyFromHTML(bodyStr, profile)

	return profile, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func parseNextJSPayload(script string, profile *GFGProfile) bool {
	re := regexp.MustCompile(`"data":\{([^{}]*"total_problems_solved":[^{}]*)\}`)

	matches := re.FindStringSubmatch(script)
	if len(matches) >= 2 {
		jsonStr := `{"data":{` + matches[1] + `}}`
		var wrapper struct {
			Data gfgUserData `json:"data"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &wrapper); err == nil && wrapper.Data.TotalProblemsSolved > 0 {
			applyProfile(wrapper.Data, profile)
			return true
		}
	}

	unescaped := unescapeNextJS(script)
	if unescaped == "" {
		return false
	}

	matches2 := re.FindStringSubmatch(unescaped)
	if len(matches2) >= 2 {
		jsonStr := `{"data":{` + matches2[1] + `}}`
		var wrapper struct {
			Data gfgUserData `json:"data"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &wrapper); err == nil && wrapper.Data.TotalProblemsSolved > 0 {
			applyProfile(wrapper.Data, profile)
			return true
		}
	}

	return false
}

func applyProfile(d gfgUserData, profile *GFGProfile) {
	profile.CodingScore = d.Score
	profile.MaxRating   = d.Score
	profile.TotalSolved = d.TotalProblemsSolved
	profile.Streak      = d.PodSolvedLongestStreak
	profile.ContestsParticipated = d.PodCorrectSubmissionsCount
	if d.InstituteRank != "" {
		rank, _ := strconv.Atoi(d.InstituteRank)
		profile.GlobalRank = rank
	}
}

func parseDifficultyFromHTML(html string, profile *GFGProfile) {
	patterns := map[string]*int{
		`School\s*\((\d+)\)`: nil,
		`Basic\s*\((\d+)\)`:  nil,
		`Easy\s*\((\d+)\)`:   nil,
		`Medium\s*\((\d+)\)`: &profile.MediumSolved,
		`Hard\s*\((\d+)\)`:   &profile.HardSolved,
	}

	school, basic, easy := 0, 0, 0

	for pattern, target := range patterns {
		re := regexp.MustCompile(pattern)
		m := re.FindStringSubmatch(html)
		if len(m) < 2 {
			continue
		}
		val, _ := strconv.Atoi(m[1])
		switch {
		case strings.Contains(pattern, "School"):
			school = val
		case strings.Contains(pattern, "Basic"):
			basic = val
		case strings.Contains(pattern, "Easy"):
			easy = val
		default:
			if target != nil {
				*target = val
			}
		}
	}

	profile.EasySolved = school + basic + easy
}

func unescapeNextJS(s string) string {
	re := regexp.MustCompile(`self\.__next_f\.push\(\[1,"((?:[^"\\]|\\.)*)"\]\)`)
	matches := re.FindAllStringSubmatch(s, -1)

	var parts []string
	for _, m := range matches {
		if len(m) >= 2 {
			jsonBytes := []byte(`"` + m[1] + `"`)
			var unescaped string
			if err := json.Unmarshal(jsonBytes, &unescaped); err == nil {
				parts = append(parts, unescaped)
			}
		}
	}
	return strings.Join(parts, "")
}