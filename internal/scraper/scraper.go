package scraper

import (
	"errors"
	"io"
	"net/http"
	"time"
	// "regexp"
	// "strconv"
	// "strings"

	// "coding-profile-service/pkg/model"
	"github.com/PuerkitoBio/goquery"
	"fmt"
)

func fetchHTML(url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to fetch data")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func FetchDoc(url string) (*goquery.Document, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s (%d)", resp.Status, resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	return doc, nil
}
