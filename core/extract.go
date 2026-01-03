package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetMediaID(url string, client *http.Client) (string, error) {
	req, _ := NewRequest("GET", url)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	id := doc.Find("#watch-block").AttrOr("data-id", "")
	if id == "" {
		id = doc.Find("div.detail_page-watch").AttrOr("data-id", "")
	}
	if id == "" {
		id = doc.Find("#movie_id").AttrOr("value", "")
	}

	if id == "" {
		return "", fmt.Errorf("could not find media ID")
	}
	return id, nil
}

func GetSeasons(mediaID string, client *http.Client) ([]Season, error) {
	url := fmt.Sprintf("%s/season/list/%s", FLIXHQ_AJAX_URL, mediaID)
	req, _ := NewRequest("GET", url)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var seasons []Season
	doc.Find(".dropdown-item").Each(func(i int, s *goquery.Selection) {
		id := s.AttrOr("data-id", "")
		name := strings.TrimSpace(s.Text())
		if id != "" {
			seasons = append(seasons, Season{ID: id, Name: name})
		}
	})
	return seasons, nil
}

func GetEpisodes(id string, isSeason bool, client *http.Client) ([]Episode, error) {
	endpoint := "movie/episodes"
	if isSeason {
		endpoint = "season/episodes"
	}
	url := fmt.Sprintf("%s/%s/%s", FLIXHQ_AJAX_URL, endpoint, id)

	req, _ := NewRequest("GET", url)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var episodes []Episode
	doc.Find(".nav-item a").Each(func(i int, s *goquery.Selection) {
		id := s.AttrOr("data-id", "")
		if id == "" {
			id = s.AttrOr("data-linkid", "")
		}
		name := strings.TrimSpace(s.AttrOr("title", s.Text()))
		if name == "" {
			name = s.Text()
		}
		if id != "" {
			episodes = append(episodes, Episode{ID: id, Name: name})
		}
	})

	if len(episodes) == 0 {
		doc.Find("a.eps-item").Each(func(i int, s *goquery.Selection) {
			id := s.AttrOr("data-id", "")
			name := strings.TrimSpace(s.AttrOr("title", s.Text()))
			if id != "" {
				episodes = append(episodes, Episode{ID: id, Name: name})
			}
		})
	}
	return episodes, nil
}

func GetServers(episodeID string, client *http.Client) ([]Server, error) {
	url := fmt.Sprintf("%s/episode/servers/%s", FLIXHQ_AJAX_URL, episodeID)
	req, _ := NewRequest("GET", url)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var servers []Server
	doc.Find(".nav-item a").Each(func(i int, s *goquery.Selection) {
		id := s.AttrOr("data-id", "")
		name := strings.TrimSpace(s.Find("span").Text())
		if name == "" {
			name = strings.TrimSpace(s.Text())
		}
		if id != "" {
			servers = append(servers, Server{ID: id, Name: name})
		}
	})
	return servers, nil
}

func GetLink(serverID string, client *http.Client) (string, error) {
	url := fmt.Sprintf("%s/episode/sources/%s", FLIXHQ_AJAX_URL, serverID)
	req, _ := NewRequest("GET", url)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		Link string `json:"link"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return res.Link, nil
}
