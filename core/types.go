package core

type Action string
type MediaType string

const (
	ActionPlay     Action = "play"
	ActionDownload Action = "download"
	ActionCast     Action = "cast"

	Movie  MediaType = "movie"
	Series MediaType = "series"
)

const (
	FLIXHQ_BASE_URL   = "https://flixhq.to"
	FLIXHQ_SEARCH_URL = FLIXHQ_BASE_URL + "/search"
	FLIXHQ_AJAX_URL   = FLIXHQ_BASE_URL + "/ajax"
	DECODER           = "https://dec.eatmynerds.live"
)

type SearchResult struct {
	Title  string
	URL    string
	Type   MediaType
	Poster string
}

type Season struct {
	ID   string
	Name string
}

type Episode struct {
	ID   string
	Name string
}

type Server struct {
	ID   string
	Name string
}
