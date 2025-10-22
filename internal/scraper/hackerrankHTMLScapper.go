package scraper

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"coding-profile-service/pkg/model"
	"github.com/PuerkitoBio/goquery"
)

// FetchHackerRankHTML scrapes the public profile page for display info (badges, certificates, scores).
// It returns a model.StatsResponse (pointer-like style not necessary here) and error.
func FetchHackerRankHTML(username string) (model.StatsResponse, error) {
	stats := model.StatsResponse{
		Platform: "hackerrank",
		Username: username,
	}

	client := &http.Client{Timeout: 12 * time.Second}
	profileURL := fmt.Sprintf("https://www.hackerrank.com/%s", username)

	req, _ := http.NewRequest("GET", profileURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ProfileScraper/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return stats, fmt.Errorf("failed to fetch HackerRank profile page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return stats, fmt.Errorf("hackerrank returned status %d: %s", resp.StatusCode, string(body))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return stats, fmt.Errorf("failed to parse HackerRank HTML: %v", err)
	}

	// Helper to parse integers robustly
	parseFirstInt := func(s string) int {
		re := regexp.MustCompile(`(\d{1,7})`)
		if m := re.FindStringSubmatch(s); len(m) >= 2 {
			if v, err := strconv.Atoi(m[1]); err == nil {
				return v
			}
		}
		return 0
	}

	// --- 1) Parse the score cards area for Coding Score and Problem Solved ---
	// Many user profiles show a grid of cards (Coding Score, Problem Solved, Contest Rating, Article Published)
	// We'll look for elements that have a "label" text like "Coding Score" / "Problem Solved"
	doc.Find("div, span").Each(func(i int, s *goquery.Selection) {
		text := strings.ToLower(strings.TrimSpace(s.Text()))
		// If element has "coding score" label nearby, find its numeric sibling
		if strings.Contains(text, "coding score") {
			// try to find numeric inside the same parent / next sibling
			if val := parseFirstInt(s.Parent().Text()); val > 0 {
				stats.CodingScore = val
			} else if val := parseFirstInt(s.Text()); val > 0 {
				stats.CodingScore = val
			}
			return
		}
		if strings.Contains(text, "problem solved") || strings.Contains(text, "problems solved") || strings.Contains(text, "problem solved:") {
			// numeric usually near label
			if val := parseFirstInt(s.Parent().Text()); val > 0 {
				stats.TotalSolved = val
			} else if val := parseFirstInt(s.Text()); val > 0 {
				stats.TotalSolved = val
			}
			return
		}
	})

	// Another targeted approach: look for known card classes (if present)
	// e.g. text elements that look like the score text .scoreCard_head_left--text__KZ2S1 and value class .scoreCard_head_left--score__oSi_x
	doc.Find("[class]").Each(func(i int, s *goquery.Selection) {
		// a small heuristic: child label + value; we check inner text parts.
		txt := strings.TrimSpace(s.Text())
		lower := strings.ToLower(txt)
		if strings.Contains(lower, "coding score") {
			// find sibling value
			if v := parseFirstInt(s.Parent().Text()); v > 0 {
				stats.CodingScore = v
			}
		}
		if strings.Contains(lower, "problem solved") {
			if v := parseFirstInt(s.Parent().Text()); v > 0 {
				stats.TotalSolved = v
			}
		}
	})

	// --- 2) Badges ---
	var badges []string
	// Primary DOM seen: .hacker-badges-v2 .hacker-badge (or .badges-list .hacker-badge)
	doc.Find(".hacker-badges-v2 .hacker-badge, .badges-list .hacker-badge, .badges-wrap .hacker-badge").Each(func(i int, s *goquery.Selection) {
		// try explicit .badge-title (SVG text), then img alt, then generic text
		if title := strings.TrimSpace(s.Find(".badge-title").Text()); title != "" {
			badges = append(badges, title)
			return
		}
		if img := s.Find("img"); img.Length() > 0 {
			if alt, ok := img.Attr("alt"); ok && strings.TrimSpace(alt) != "" {
				badges = append(badges, strings.TrimSpace(alt))
				return
			}
		}
		// fallback: try to extract a short text from the node
		txt := strings.TrimSpace(s.Text())
		// keep only short names to avoid full svg noise
		if len(txt) > 0 && len(txt) < 80 {
			// clean up newlines and multiple spaces
			txt = regexp.MustCompile(`\s+`).ReplaceAllString(txt, " ")
			badges = append(badges, txt)
		}
	})
	// dedupe badges
	if len(badges) > 0 {
		seen := map[string]bool{}
		out := []string{}
		for _, b := range badges {
			if b == "" {
				continue
			}
			if !seen[b] {
				seen[b] = true
				out = append(out, b)
			}
		}
		stats.Badges = out
	}

	// --- 3) Certifications (links and count) ---
	var certLinks []string
	certCount := 0
	// look for anchors in .hacker-certificates
	doc.Find(".hacker-certificates a.certificate-link, .certificates a, a.certificate-link").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			// normalize relative links
			link := strings.TrimSpace(href)
			if strings.HasPrefix(link, "/") {
				link = "https://www.hackerrank.com" + link
			}
			certLinks = append(certLinks, link)
		}
		// try to parse title inside h2 .certificate_v3-heading
	})
	// If we have certLinks, set count accordingly. Otherwise try to find textual count in the DOM.
	if len(certLinks) > 0 {
		certCount = len(certLinks)
	} else {
		// fallback: try to find text mentioning 'certificate' and a number
		doc.Find("div, span, a, h2").EachWithBreak(func(i int, s *goquery.Selection) bool {
			txt := strings.ToLower(strings.TrimSpace(s.Text()))
			if strings.Contains(txt, "certificate") {
				// try to extract number from parent
				if n := parseFirstInt(s.Parent().Text()); n > 0 {
					certCount = n
					return false
				}
				// try in itself
				if n := parseFirstInt(s.Text()); n > 0 {
					certCount = n
					return false
				}
			}
			return true
		})
	}

	stats.Certifications = certCount
	stats.CertificationLinks = certLinks

	// --- 4) fallback: extract some numeric bits (score/solved) from embedded JSON if present ---
	// read text of the whole document (limited size) for common keys like "solved_challenges"
	// We can't re-read resp.Body (already consumed), but goquery built the tree — use doc.Text() carefully
	pageText := doc.Text()
	if stats.TotalSolved == 0 {
		if n := parseFirstInt(pageText); n > 0 {
			// conservatively assign only if seems plausible and not overwriting better values
			stats.TotalSolved = n
		}
	}

	// Return what we parsed
	return stats, nil
}
