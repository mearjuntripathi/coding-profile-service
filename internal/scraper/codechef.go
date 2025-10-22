package scraper

import (
    "coding-profile-service/pkg/model"
    "fmt"
)

// FetchCodeChef is the function your API handler calls.
func FetchCodeChef(username string) (model.StatsResponse, error) {
    stats, err := FetchCodeChefHTML(username) // call the HTML scraper
    if err != nil {
        return model.StatsResponse{
            Platform: "codechef",
            Username: username,
            Error:    fmt.Sprintf("could not fetch CodeChef data for user: %s", username),
        }, err
    }

    return stats, nil
}
