package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"coding-profile-service/pkg/model"
)

func FetchLeetCode(username string) (model.StatsResponse, error) {
	query := fmt.Sprintf(`
	{
	  matchedUser(username: "%s") {
	    username
	    submitStatsGlobal {
	      acSubmissionNum {
	        difficulty
	        count
	      }
	    }
	  }
	}`, username)

	body, _ := json.Marshal(map[string]string{"query": query})
	req, _ := http.NewRequest("POST", "https://leetcode.com/graphql", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return model.StatsResponse{}, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return model.StatsResponse{}, err
	}

	data := result["data"].(map[string]interface{})["matchedUser"]
	if data == nil {
		return model.StatsResponse{}, fmt.Errorf("user not found")
	}

	stats := data.(map[string]interface{})["submitStatsGlobal"].(map[string]interface{})["acSubmissionNum"].([]interface{})
	respModel := model.StatsResponse{
		Platform: "leetcode",
		Username: username,
	}

	for _, s := range stats {
		item := s.(map[string]interface{})
		switch item["difficulty"].(string) {
		case "All":
			respModel.TotalSolved = int(item["count"].(float64))
		case "Easy":
			respModel.EasySolved = int(item["count"].(float64))
		case "Medium":
			respModel.MediumSolved = int(item["count"].(float64))
		case "Hard":
			respModel.HardSolved = int(item["count"].(float64))
		}
	}
	return respModel, nil
}
