package core

import (
	"errors"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func SearchContent(query string, client *http.Client) ([]SearchResult, error) {
	search := strings.ReplaceAll(query, " ", "-")
	req, _ := NewRequest("GET", FLIXHQ_SEARCH_URL+"/"+search)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	doc.Find("div.flw-item").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i >= 10 {
			return false
		}

		title := s.Find("h2.film-name a").AttrOr("title", "Unknown")
		href := s.Find("div.film-poster a").AttrOr("href", "")
		poster := s.Find("img.film-poster-img").AttrOr("data-src", "")
		typeStr := strings.TrimSpace(s.Find("span.fdi-type").Text())
		
		mediaType := Movie
		if strings.EqualFold(typeStr, "TV") || strings.EqualFold(typeStr, "Series") {
			mediaType = Series
		}

		if href != "" {
			results = append(results, SearchResult{
				Title:  title,
				URL:    FLIXHQ_BASE_URL + href,
				Type:   mediaType,
				Poster: poster,
			})
		}
		return true
	})

	if len(results) == 0 {
		return nil, errors.New("no results")
	}

	return results, nil
}
